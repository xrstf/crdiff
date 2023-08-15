// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package oasdiff

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/checker/localizations"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/load"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func NewConfig() *diff.Config {
	return diff.NewConfig()
}

func CompareSchemas(cfg *diff.Config, base, revision *apiextensionsv1.JSONSchemaProps) (diffResult *diff.SchemaDiff, breakingChanges checker.Changes, err error) {
	openapiBase, err := jsonschemaToOpenapiSchema(base)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert base CRD into openapi spec: %w", err)
	}
	baseSpecInfo := &load.SpecInfo{Url: "http://example.com", Spec: openapiBase}

	openapiRevision, err := jsonschemaToOpenapiSchema(revision)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert revision CRD into openapi spec: %w", err)
	}
	revisionSpecInfo := &load.SpecInfo{Url: "http://example.com", Spec: openapiRevision}

	// step 1: diff both schemas against each other to determine all individual changes

	if cfg == nil {
		cfg = diff.NewConfig()
	}

	changes, opSources, err := diff.GetWithOperationsSourcesMap(cfg, baseSpecInfo, revisionSpecInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to compare specs: %w", err)
	}

	if changes == nil {
		return nil, nil, nil
	}

	completeDiff := changes.PathsDiff.Modified[dummySchemaPath].OperationsDiff.Modified["POST"].RequestBodyDiff.ContentDiff.MediaTypeModified[dummyContentType].SchemaDiff

	// step 2: run the checker to determine backwards-incompatible (= breaking) changes
	// this is adapted from oasdiff/internal/changelog.go

	// oasdiff sets this to 5 by default
	openapi3.CircularReferenceCounter = 5

	includeChecks := []string{}
	deprecationDaysBeta := 0   // we do not do deprecation periods based on days
	deprecationDaysStable := 0 // same

	bcConfig := checker.GetAllChecks(includeChecks, deprecationDaysBeta, deprecationDaysStable)
	bcConfig.Localizer = *localizations.New("en", "en")

	breakingChanges = checker.CheckBackwardCompatibilityUntilLevel(bcConfig, changes, opSources, checker.WARN)

	// filter the breaking changes and remove misleading context
	for k, change := range breakingChanges {
		breakingChanges[k] = checker.ApiChange{
			Id:    change.GetId(),
			Text:  change.GetText(),
			Level: change.GetLevel(),
		}
	}

	return completeDiff, breakingChanges, nil
}

type openapi3Schema struct {
	// Components openapi3SchemaComponents     `json:"components"`
	Paths map[string]*openapi3PathItem `json:"paths"`
}

// type openapi3SchemaComponents struct {
// 	Schemas map[string]interface{} `json:"schemas"`
// }

type openapi3PathItem struct {
	Post openapi3PostOperation `json:"post"`
}

type openapi3PostOperation struct {
	RequestBody openapi3RequestBodyRef `json:"requestBody"`
}

type openapi3RequestBodyRef struct {
	Required bool                              `json:"required"`
	Content  map[string]openapi3RequestContent `json:"content"`
}

type openapi3RequestContent struct {
	Schema interface{} `json:"schema"`
}

const (
	dummySchemaKey   = "crd"
	dummySchemaPath  = "/foo"
	dummyContentType = "application/json"
)

func jsonschemaToOpenapiSchema(jsonschema *apiextensionsv1.JSONSchemaProps) (*openapi3.T, error) {
	dummyT := openapi3Schema{
		// Component diffs do not support the normal/breaking change distinction.
		// Components: openapi3SchemaComponents{
		// 	Schemas: map[string]interface{}{
		// 		dummySchemaKey: jsonschema,
		// 	},
		// },
		Paths: map[string]*openapi3PathItem{
			dummySchemaPath: {
				Post: openapi3PostOperation{
					RequestBody: openapi3RequestBodyRef{
						Required: true,
						Content: map[string]openapi3RequestContent{
							dummyContentType: {
								Schema: jsonschema,
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	if err := encoder.Encode(dummyT); err != nil {
		return nil, err
	}

	return openapi3.NewLoader().LoadFromData(buf.Bytes())
}
