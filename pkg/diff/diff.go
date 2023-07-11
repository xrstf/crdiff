// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package diff

import (
	"fmt"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/diff/oasdiff"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Change string
type Diff map[string][]Change

type Options struct {
	Versions     []string
	BreakingOnly bool
}

func CompareCRDs(base, revision crd.CRD, opt Options) (Diff, error) {
	if base.Identifier() != revision.Identifier() {
		return nil, fmt.Errorf("cannot compare to different CRDs (%q vs. %q)", base.Identifier(), revision.Identifier())
	}

	baseVersions, err := base.Versions()
	if err != nil {
		return nil, fmt.Errorf("failed to determine versions of base CRD: %w", err)
	}
	baseVersionMap := limitVersions(baseVersions, opt.Versions)

	revisionVersions, err := revision.Versions()
	if err != nil {
		return nil, fmt.Errorf("failed to determine versions of revision CRD: %w", err)
	}
	revisionVersionMap := limitVersions(revisionVersions, opt.Versions)

	result := Diff{
		"": []Change{},
	}

	if base.Scope() != revision.Scope() {
		result[""] = append(result[""], Change(fmt.Sprintf("changed Scope from %q to %q", base.Scope(), revision.Scope())))
	}

	oasConfig := oasdiff.NewConfig()
	oasConfig.BreakingOnly = opt.BreakingOnly

	for _, version := range sets.List(baseVersionMap) {
		if !revisionVersionMap.Has(version) {
			result[version] = []Change{
				Change("version has been removed"),
			}
			continue
		}

		baseSchema := base.Schema(version)
		revisionSchema := revision.Schema(version)

		schemaChanges, err := oasdiff.CompareSchemas(oasConfig, baseSchema, revisionSchema)
		if err != nil {
			return nil, fmt.Errorf("failed comparing version %v: %w", version, err)
		}

		// no changes in this version :)
		if schemaChanges == nil {
			continue
		}

		result[version] = append(result[version], Change(fmt.Sprintf("%v", schemaChanges)))
	}

	for _, version := range sets.List(revisionVersionMap) {
		// we already compared base and revision
		if baseVersionMap.Has(version) {
			continue
		}

		result[version] = []Change{
			Change("version has been added"),
		}
	}

	return result, nil
}

func limitVersions(allVersions, limited []string) sets.Set[string] {
	result := sets.New(allVersions...)

	if len(limited) > 0 {
		result = result.Intersection(sets.New(limited...))
	}

	return result
}
