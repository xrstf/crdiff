// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package loader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/crdiff/pkg/crd"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

type Options struct {
	FileExtensions []string
}

func NewDefaultOptions() *Options {
	return &Options{
		FileExtensions: []string{"yaml", "yml"},
	}
}

func LoadCRDs(source string, opt *Options, log logrus.FieldLogger) (map[string]crd.CRD, error) {
	if opt == nil {
		opt = NewDefaultOptions()
	}

	stat, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}

	if stat.IsDir() {
		absSource, err := filepath.Abs(source)
		if err != nil {
			return nil, fmt.Errorf("failed to determine absolute path: %w", err)
		}

		return forbidDuplicates(loadCRDsFromDirectory(absSource, opt, log))
	}

	return forbidDuplicates(loadCRDsFromFile(source, true, opt, log))
}

func forbidDuplicates(allCRDs []crd.CRD, err error) (map[string]crd.CRD, error) {
	if err != nil {
		return nil, err
	}

	result := map[string]crd.CRD{}
	for _, crdObj := range allCRDs {
		ident := crdObj.Identifier()

		if _, err := crdObj.Versions(); err != nil {
			return nil, fmt.Errorf("%s is invalid: %w", ident, err)
		}

		if _, exists := result[ident]; exists {
			return nil, fmt.Errorf("found multiple definitions of %s", ident)
		}

		result[ident] = crdObj
	}

	return result, nil
}

const (
	bufSize = 5 * 1024 * 1024
)

func loadCRDsFromFile(source string, logFullFilename bool, opt *Options, log logrus.FieldLogger) ([]crd.CRD, error) {
	if logFullFilename {
		log = log.WithField("filename", source)
	} else {
		log = log.WithField("filename", filepath.Base(source))
	}
	log.Debug("Reading file…")

	f, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	docSplitter := yamlutil.NewDocumentDecoder(f)
	defer docSplitter.Close()

	result := []crd.CRD{}

	for i := 1; true; i++ {
		buf := make([]byte, bufSize) // 5 MB, same as chunk size in decoder
		read, err := docSplitter.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("document %d is larger than the internal buffer", i)
		}

		crdObj, err := parseYAML(buf[:read])
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return nil, fmt.Errorf("document %d is invalid: %w", i, err)
		}

		if crdObj == nil {
			continue
		}

		result = append(result, crdObj)
	}

	return result, nil
}

func loadCRDsFromDirectory(rootDir string, opt *Options, log logrus.FieldLogger) ([]crd.CRD, error) {
	log = log.WithField("directory", rootDir)

	contents, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	result := []crd.CRD{}

	for _, entry := range contents {
		fullPath := filepath.Join(rootDir, entry.Name())

		if entry.IsDir() {
			subresult, err := loadCRDsFromDirectory(fullPath, opt, log)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory %s: %w", fullPath, err)
			}
			result = append(result, subresult...)
		} else if hasExtension(entry.Name(), opt.FileExtensions) {
			subresult, err := loadCRDsFromFile(fullPath, false, opt, log)
			if err != nil {
				return nil, fmt.Errorf("failed to read file %s: %w", fullPath, err)
			}
			result = append(result, subresult...)
		}
	}

	return result, nil
}

func hasExtension(filename string, extensions []string) bool {
	parts := strings.Split(filename, ".")
	extension := parts[len(parts)-1]

	for _, ext := range extensions {
		if ext == extension {
			return true
		}
	}

	return false
}

func parseYAML(data []byte) (crd.CRD, error) {
	candidate := unstructured.Unstructured{}

	err := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024).Decode(&candidate)
	if err != nil {
		return nil, fmt.Errorf("document is not valid YAML: %w", err)
	}

	if candidate.GetKind() != "CustomResourceDefinition" {
		return nil, nil
	}

	var crdObj crd.CRD

	decoder := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024)

	switch candidate.GetAPIVersion() {
	case apiextensionsv1.SchemeGroupVersion.String():
		crdInstance := apiextensionsv1.CustomResourceDefinition{}
		if err := decoder.Decode(&crdInstance); err != nil {
			return nil, fmt.Errorf("document is not valid apiextensions/v1 CustomResourceDefinition: %w", err)
		}
		crdObj = crd.NewV1(crdInstance)

	case apiextensionsv1beta1.SchemeGroupVersion.String():
		crdInstance := apiextensionsv1beta1.CustomResourceDefinition{}
		if err := decoder.Decode(&crdInstance); err != nil {
			return nil, fmt.Errorf("document is not valid apiextensions/v1beta1 CustomResourceDefinition: %w", err)
		}
		crdObj = crd.NewV1beta1(crdInstance)

	default:
		return nil, fmt.Errorf("document is using unrecognized API version %q", candidate.GetAPIVersion())
	}

	return crdObj, nil
}
