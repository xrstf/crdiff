// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package crd

import (
	"bytes"
	"encoding/json"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type v1beta1 struct {
	crd apiextensionsv1beta1.CustomResourceDefinition
}

func NewV1beta1(crd apiextensionsv1beta1.CustomResourceDefinition) CRD {
	return &v1beta1{crd}
}

func (c *v1beta1) Identifier() string {
	return fmt.Sprintf("%s/%s", c.crd.Spec.Group, c.crd.Spec.Names.Kind)
}

func (c *v1beta1) Scope() string {
	return string(c.crd.Spec.Scope)
}

func (c *v1beta1) Versions() ([]string, error) {
	versions := sets.New[string]()
	for _, v := range c.crd.Spec.Versions {
		if versions.Has(v.Name) {
			return nil, fmt.Errorf("defines version %q multiple times", v.Name)
		}
		versions.Insert(v.Name)
	}

	return sets.List(versions), nil
}

// Schema converts the v1beta1 schema to v1, as those types are thankfully
// identical and it makes diffing easier if all CRDs use the same type for
// their schemas.
func (c *v1beta1) Schema(version string) *apiextensionsv1.JSONSchemaProps {
	for _, v := range c.crd.Spec.Versions {
		if v.Name == version {
			var buf bytes.Buffer
			if json.NewEncoder(&buf).Encode(v.Schema.OpenAPIV3Schema) != nil {
				return nil
			}

			var v1 apiextensionsv1.JSONSchemaProps
			if json.NewDecoder(&buf).Decode(&v1) != nil {
				return nil
			}

			return &v1
		}
	}

	return nil
}
