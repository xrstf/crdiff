// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package compare

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/loader"

	"k8s.io/apimachinery/pkg/api/equality"
)

func TestCompareCRDs(t *testing.T) {
	testcases, err := filepath.Glob("testdata/*.base.yaml")
	if err != nil {
		t.Fatalf("Failed to find testcases: %v", err)
	}

	for _, baseFile := range testcases {
		basename := strings.Replace(filepath.Base(baseFile), ".base.yaml", "", -1)

		t.Run(basename, func(t *testing.T) {
			testCompareSingleCRD(t, baseFile)
		})
	}
}

func testCompareSingleCRD(t *testing.T, baseFile string) {
	t.Helper()

	revisionFile := strings.Replace(baseFile, ".base.yaml", ".revision.yaml", -1)
	diffFile := strings.Replace(baseFile, ".base.yaml", ".diff.yaml", -1)

	log := logrus.New()
	log.SetOutput(io.Discard)

	baseCRD, err := loadCRD(log, baseFile)
	if err != nil {
		t.Fatalf("Failed to load base CRD: %v", err)
	}

	revisionCRD, err := loadCRD(log, revisionFile)
	if err != nil {
		t.Fatalf("Failed to load revision CRD: %v", err)
	}

	opt := CompareOptions{}

	result, err := CompareCRDs(baseCRD, revisionCRD, opt)
	if err != nil {
		t.Fatalf("Failed to compare CRDs: %v", err)
	}

	var encoded bytes.Buffer

	encoder := yaml.NewEncoder(&encoded)
	encoder.SetIndent(2)

	if err := encoder.Encode(result); err != nil {
		t.Fatalf("Failed to encode resulting diff as YAML: %v", err)
	}

	if err := compareExpectedResult(encoded.Bytes(), diffFile); err != nil {
		t.Fatalf("Result did not meet expectations: %v", err)
	}
}

func compareExpectedResult(actualBytes []byte, diffFile string) error {
	var (
		actual   map[string]interface{}
		expected map[string]interface{}
	)

	if err := yaml.NewDecoder(bytes.NewReader(actualBytes)).Decode(&actual); err != nil {
		return fmt.Errorf("failed to decode actual result: %w", err)
	}

	content, err := os.ReadFile(diffFile)
	if err != nil {
		return fmt.Errorf("failed to read expected diff file: %w", err)
	}

	if len(content) > 0 {
		if err := yaml.NewDecoder(bytes.NewReader(content)).Decode(&expected); err != nil {
			return fmt.Errorf("failed to decode expected diff result: %w", err)
		}
	}

	if !equality.Semantic.DeepEqual(actual, expected) {
		diff := stringDiff(string(content), string(actualBytes))
		return fmt.Errorf("expectation not met:\n\n%s\n", diff)
	}

	return nil
}

func loadCRD(log logrus.FieldLogger, filename string) (crd.CRD, error) {
	crds, err := loader.LoadCRDs(filename, nil, log)
	if err != nil {
		return nil, fmt.Errorf("failed to load CRDs: %w", err)
	}
	if len(crds) != 1 {
		return nil, fmt.Errorf("expected 1 CRD, but found %d.", len(crds))
	}

	for k := range crds {
		return crds[k], nil
	}

	return nil, errors.New("this code can never be reached")
}

func stringDiff(expected, actual string) string {
	expected = strings.TrimSpace(expected)
	actual = strings.TrimSpace(actual)

	diff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		ToFile:   "Actual",
		Context:  3,
	}

	unidifiedDiff, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		return fmt.Sprintf("<failed to generate diff: %v>", err)
	}

	return unidifiedDiff
}
