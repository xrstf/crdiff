// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package compare

import (
	"encoding/json"
	"fmt"

	"github.com/tufin/oasdiff/checker"
	"github.com/tufin/oasdiff/diff"
	"github.com/tufin/oasdiff/utils"

	"go.xrstf.de/crdiff/pkg/compare/oasdiff"
	"go.xrstf.de/crdiff/pkg/crd"

	"k8s.io/apimachinery/pkg/util/sets"
)

type CompareOptions struct {
	Versions           []string
	BreakingOnly       bool
	IgnoreDescriptions bool
}

func CompareCRDs(base, revision crd.CRD, opt CompareOptions) (*CRDDiff, error) {
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
		General:         []Change{},
		AddedVersions:   utils.StringList{},
		DeletedVersions: utils.StringList{},
		ChangedVersions: map[string]CRDVersionDiff{},
	}

	// detect general, non-schema related changes

	if base.Scope() != revision.Scope() {
		result.General = append(result.General, Change{
			Breaking:    true,
			Description: fmt.Sprintf("changed scope from %q to %q", base.Scope(), revision.Scope()),
		})
	}

	// compare schemas

	oasConfig := oasdiff.NewConfig()

	for _, version := range sets.List(baseVersionMap) {
		if !revisionVersionMap.Has(version) {
			result.DeletedVersions = append(result.DeletedVersions, version)
			continue
		}

		baseSchema := base.Schema(version)
		revisionSchema := revision.Schema(version)

		completeDiff, breakingChanges, err := oasdiff.CompareSchemas(oasConfig, baseSchema, revisionSchema)
		if err != nil {
			return nil, fmt.Errorf("failed comparing version %v: %w", version, err)
		}

		// no changes in this version :)
		if completeDiff == nil {
			continue
		}

		result.ChangedVersions[version] = createCRDVersionDiff(completeDiff, breakingChanges, &opt)
	}

	// detect newly added versions in this CRD

	for _, version := range sets.List(revisionVersionMap) {
		// we already compared base and revision
		if baseVersionMap.Has(version) {
			continue
		}

		if !opt.BreakingOnly {
			result.AddedVersions = append(result.AddedVersions, version)
		}
	}

	result.AddedVersions.Sort()
	result.DeletedVersions.Sort()

	return result, nil
}

func createCRDVersionDiff(diff *diff.SchemaDiff, breaking checker.Changes, opt *CompareOptions) CRDVersionDiff {
	result := CRDVersionDiff{
		SchemaChanges:   map[string]CRDSchemaDiff{},
		BreakingChanges: make([]BreakingChange, len(breaking)),
	}

	for i, change := range breaking {
		msg := oasdiff.LocalizedMessage{}

		// unwrap the localizer data we sneakily injected by using a JSON localizer
		if text := change.GetText(); text != "" {
			if err := json.Unmarshal([]byte(text), &msg); err != nil {
				panic(err)
			}
		}

		result.BreakingChanges[i] = BreakingChange{
			ID:        change.GetId(),
			Level:     change.GetLevel(),
			InfoKey:   msg.Key,
			Arguments: msg.Args,
		}
	}

	collectChangesFromSchemaDiff(result, diff, opt, "")

	return result
}

func hasLeafDiff(diff *diff.SchemaDiff, opt *CompareOptions) bool {
	// to keep things easy, we modify the struct if the user disabled certain checks
	if opt.IgnoreDescriptions {
		diff.DescriptionDiff = nil
	}

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
		// diff.ItemsDiff is not considered a leaf
		diff.RequiredDiff != nil ||
		// diff.PropertiesDiff is not considered a leaf
		diff.MinPropsDiff != nil ||
		diff.MaxPropsDiff != nil ||
		// diff.AdditionalPropertiesDiff != nil ||
		diff.DiscriminatorDiff != nil
}

func rootPath(path string) string {
	if path == "" {
		return "."
	}

	return path
}

func collectChangesFromSchemaDiff(result CRDVersionDiff, sd *diff.SchemaDiff, opt *CompareOptions, path string) {
	if hasLeafDiff(sd, opt) {
		// strip Items- & Properties Diff, as those are treated differently
		// and should not show up twice in the resulting diff
		copied := oasdiff.DeepCopySchemaDiff(sd)
		copied.PropertiesDiff = nil
		copied.ItemsDiff = nil

		result.SchemaChanges[rootPath(path)] = CRDSchemaDiff{
			Diff: copied,
		}
	}

	switch {
	case sd.ItemsDiff != nil:
		collectChangesFromSchemaDiff(result, sd.ItemsDiff, opt, path+".[]")

	case sd.PropertiesDiff != nil:
		collectChangesFromSchemasDiff(result, sd.PropertiesDiff, opt, path)

	case sd.AdditionalPropertiesDiff != nil:
		collectChangesFromSchemaDiff(result, sd.AdditionalPropertiesDiff, opt, path+".*")
	}
}

func collectChangesFromSchemasDiff(result CRDVersionDiff, sd *diff.SchemasDiff, opt *CompareOptions, path string) {
	if len(sd.Added) > 0 || len(sd.Deleted) > 0 {
		schemaDiff := result.SchemaChanges[rootPath(path)] // rely on Go's runtime defaulting for non-existing keys
		schemaDiff.AddedProperties = sd.Added
		schemaDiff.DeletedProperties = sd.Deleted

		result.SchemaChanges[rootPath(path)] = schemaDiff
	}

	for k, v := range sd.Modified {
		collectChangesFromSchemaDiff(result, v, opt, path+"."+k)
	}
}

func limitVersions(allVersions, limited []string) sets.Set[string] {
	result := sets.New(allVersions...)

	if len(limited) > 0 {
		result = result.Intersection(sets.New(limited...))
	}

	return result
}
