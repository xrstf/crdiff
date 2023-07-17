// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package diff

import (
	"fmt"

	"github.com/tufin/oasdiff/diff"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/diff/oasdiff"

	"k8s.io/apimachinery/pkg/util/sets"
)

type Change struct {
	Modification string
}

type CRDDiff struct {
	General  []Change
	Versions CRDVersionDiff
}

type CRDVersionDiff struct {
	AddedVersions   sets.Set[string]
	DeletedVersions sets.Set[string]
	ChangedVersions map[string]CRDSchemasDiff
}

type CRDSchemasDiff map[string]CRDSchemaDiff

type CRDSchemaDiff struct {
	AddedProperties   sets.Set[string]
	DeletedProperties sets.Set[string]
	Diff              *diff.SchemaDiff
}

func (d *CRDDiff) Empty() bool {
	if d == nil {
		return true
	}

	if len(d.General) > 0 {
		return false
	}

	if d.Versions.AddedVersions.Len() > 0 || d.Versions.DeletedVersions.Len() > 0 {
		return false
	}

	for _, versionDiff := range d.Versions.ChangedVersions {
		if len(versionDiff) > 0 {
			return false
		}
	}

	return true
}

type Options struct {
	Versions     []string
	BreakingOnly bool
}

func CompareCRDs(base, revision crd.CRD, opt Options) (*CRDDiff, error) {
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

	result := &CRDDiff{
		General: []Change{},
		Versions: CRDVersionDiff{
			AddedVersions:   sets.New[string](),
			DeletedVersions: sets.New[string](),
			ChangedVersions: map[string]CRDSchemasDiff{},
		},
	}

	if base.Scope() != revision.Scope() {
		result.General = append(result.General, Change{
			Modification: fmt.Sprintf("changed scope from %q to %q", base.Scope(), revision.Scope()),
		})
	}

	oasConfig := oasdiff.NewConfig()
	oasConfig.BreakingOnly = opt.BreakingOnly

	for _, version := range sets.List(baseVersionMap) {
		if !revisionVersionMap.Has(version) {
			result.Versions.DeletedVersions.Insert(version)
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

		result.Versions.ChangedVersions[version] = createCRDSchemasDiff(schemaChanges)
	}

	for _, version := range sets.List(revisionVersionMap) {
		// we already compared base and revision
		if baseVersionMap.Has(version) {
			continue
		}

		if !opt.BreakingOnly {
			result.Versions.AddedVersions.Insert(version)
		}
	}

	return result, nil
}

func createCRDSchemasDiff(diff *diff.SchemaDiff) CRDSchemasDiff {
	result := CRDSchemasDiff{}

	// printSchemaDiff(diff)

	collectChangesFromSchemaDiff(result, diff, "")

	return result
}

func hasLeafDiff(diff *diff.SchemaDiff) bool {
	return diff.ExtensionsDiff != nil ||
		diff.OneOfDiff != nil ||
		diff.AnyOfDiff != nil ||
		diff.AllOfDiff != nil ||
		diff.NotDiff != nil ||
		diff.TypeDiff != nil ||
		diff.TitleDiff != nil ||
		diff.FormatDiff != nil ||
		diff.DescriptionDiff != nil ||
		diff.EnumDiff != nil ||
		diff.DefaultDiff != nil ||
		diff.ExampleDiff != nil ||
		diff.ExternalDocsDiff != nil ||
		diff.AdditionalPropertiesAllowedDiff != nil ||
		diff.UniqueItemsDiff != nil ||
		diff.ExclusiveMinDiff != nil ||
		diff.ExclusiveMaxDiff != nil ||
		diff.NullableDiff != nil ||
		diff.ReadOnlyDiff != nil ||
		diff.WriteOnlyDiff != nil ||
		diff.AllowEmptyValueDiff != nil ||
		diff.XMLDiff != nil ||
		diff.DeprecatedDiff != nil ||
		diff.MinDiff != nil ||
		diff.MaxDiff != nil ||
		diff.MultipleOfDiff != nil ||
		diff.MinLengthDiff != nil ||
		diff.MaxLengthDiff != nil ||
		diff.PatternDiff != nil ||
		diff.MinItemsDiff != nil ||
		diff.MaxItemsDiff != nil ||
		diff.RequiredDiff != nil ||
		diff.MinPropsDiff != nil ||
		diff.MaxPropsDiff != nil ||
		diff.AdditionalPropertiesDiff != nil ||
		diff.DiscriminatorDiff != nil
}

func rootPath(path string) string {
	if path == "" {
		return "."
	}

	return path
}

func collectChangesFromSchemaDiff(result CRDSchemasDiff, sd *diff.SchemaDiff, path string) {
	if hasLeafDiff(sd) {
		result[rootPath(path)] = CRDSchemaDiff{
			Diff: sd,
		}
	}

	switch {
	case sd.ItemsDiff != nil:
		collectChangesFromSchemaDiff(result, sd.ItemsDiff, path+".[]")

	case sd.PropertiesDiff != nil:
		collectChangesFromSchemasDiff(result, sd.PropertiesDiff, path)
	}
}

func collectChangesFromSchemasDiff(result CRDSchemasDiff, sd *diff.SchemasDiff, path string) {
	if len(sd.Added) > 0 || len(sd.Deleted) > 0 {
		schemaDiff := result[rootPath(path)] // rely on Go's runtime defaulting for non-existing keys
		schemaDiff.AddedProperties = sets.New(sd.Added...)
		schemaDiff.DeletedProperties = sets.New(sd.Deleted...)

		result[rootPath(path)] = schemaDiff
	}

	for k, v := range sd.Modified {
		collectChangesFromSchemaDiff(result, v, path+"."+k)
	}
}

func limitVersions(allVersions, limited []string) sets.Set[string] {
	result := sets.New(allVersions...)

	if len(limited) > 0 {
		result = result.Intersection(sets.New(limited...))
	}

	return result
}
