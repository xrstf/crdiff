// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package crd

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type v1 struct {
	crd apiextensionsv1.CustomResourceDefinition
}

func NewV1(crd apiextensionsv1.CustomResourceDefinition) CRD {
	return &v1{crd}
}

func (c *v1) Identifier() string {
	return fmt.Sprintf("%s/%s", c.crd.Spec.Group, c.crd.Spec.Names.Kind)
}

func (c *v1) Scope() string {
	return string(c.crd.Spec.Scope)
}

func (c *v1) Versions() ([]string, error) {
	versions := sets.New[string]()
	for _, v := range c.crd.Spec.Versions {
		if versions.Has(v.Name) {
			return nil, fmt.Errorf("defines version %q multiple times", v.Name)
		}
		versions.Insert(v.Name)
	}

	return sets.List(versions), nil
}

func (c *v1) Schema(version string) *apiextensionsv1.JSONSchemaProps {
	for _, v := range c.crd.Spec.Versions {
		if v.Name == version {
			return v.Schema.OpenAPIV3Schema
		}
	}

	return nil
}
