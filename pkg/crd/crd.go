// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package crd

import apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

type CRD interface {
	Identifier() string
	Versions() ([]string, error)
	Scope() string
	Schema(version string) *apiextensionsv1.JSONSchemaProps
}
