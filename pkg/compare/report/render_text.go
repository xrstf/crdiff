// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package report

import (
	"fmt"
	"strings"

	"github.com/tufin/oasdiff/diff"

	"go.xrstf.de/crdiff/pkg/colors"
	"go.xrstf.de/crdiff/pkg/compare"
	"go.xrstf.de/crdiff/pkg/indent"

	"k8s.io/apimachinery/pkg/util/sets"
)

func (r *Report) Render(breakingOnly bool) *indent.Indenter {
	printer := indent.NewIndenter()

	sortedIdentifiers := sets.List(sets.KeySet(r.Diffs))

	for i, crdIdentifier := range sortedIdentifiers {
		crdChanges := r.Diffs[crdIdentifier]

		if !shouldPrintCRD(&crdChanges, breakingOnly) {
			continue
		}

		if i > 0 {
			printer.AddLine("")
		}

		crdRendered := renderCRDDiffAsText(crdIdentifier, &crdChanges, breakingOnly)
		printer.Add(crdRendered)
	}

	return printer
}

func (r *Report) Print(breakingOnly bool) {
	fmt.Println(r.Render(breakingOnly))
}

func renderCRDDiffAsText(crdIdentifier string, crdChanges *compare.CRDDiff, breakingOnly bool) *indent.Indenter {
	printer := indent.NewIndenter()
	printer.AddLine(heading(crdIdentifier, "=", "crd"))
	printer.Indent()

	unversionedChanges := indent.NewIndenter()

	for _, change := range crdChanges.General {
		unversionedChanges.AddLinef("~ %s", change.Description)
	}

	if !breakingOnly {
		for _, version := range crdChanges.AddedVersions {
			unversionedChanges.AddLinef("+ %s %s", colors.ActionAdd.Render("added"), colors.Version.Render(version))
		}
	}

	for _, version := range crdChanges.DeletedVersions {
		unversionedChanges.AddLinef("- %s %s", colors.ActionRemove.Render("removed"), colors.Version.Render(version))
	}

	if !unversionedChanges.Empty() {
		printer.AddLine("")
		printer.Add(unversionedChanges)
	}

	// ensure a stable, sorted order of versions
	changedVersions := sets.List(sets.KeySet(crdChanges.ChangedVersions))
	for _, version := range changedVersions {
		versionDiff := crdChanges.ChangedVersions[version]
		if !shouldPrintCRDVersion(versionDiff, breakingOnly) {
			continue
		}

		renderedVersionDiff := renderCRDVersionDiffAsText(version, &versionDiff, breakingOnly)
		if renderedVersionDiff != nil {
			printer.AddLine("")
			printer.Add(renderedVersionDiff)
		}
	}

	return printer
}

func renderCRDVersionDiffAsText(version string, versionDiff *compare.CRDVersionDiff, breakingOnly bool) *indent.Indenter {
	blocks := []*indent.Indenter{}

	// ensure a stable, sorted order of paths
	changedPaths := sets.List(sets.KeySet(versionDiff.SchemaChanges))
	for _, path := range changedPaths {
		pathChanges := versionDiff.SchemaChanges[path]

		changes := indent.NewIndenter()

		for _, field := range pathChanges.AddedProperties {
			changes.AddLinef("+ %s %s", colors.ActionAdd.Render("added"), colors.Property.Render(field))
		}

		for _, field := range pathChanges.DeletedProperties {
			changes.AddLinef("- %s %s", colors.ActionRemove.Render("removed"), colors.Property.Render(field))
		}

		if d := pathChanges.Diff; d != nil {
			printSchemaDiff(d, changes)
		}

		if !changes.Empty() {
			block := indent.NewIndenter()
			block.AddLinef("%s:", colors.Path.Render(path))
			block.Indent()
			block.Add(changes)

			blocks = append(blocks, block)
		}
	}

	breaking := indent.NewIndenter()
	for _, b := range versionDiff.BreakingChanges {
		breaking.AddLine(fmt.Sprintf("%v", b))
	}

	if !breaking.Empty() {
		blocks = append(blocks, breaking)
	}

	if len(blocks) == 0 {
		return nil
	}

	result := indent.NewIndenter()
	result.AddLine(heading(version, "-", "version"))
	result.Indent()

	for i := range blocks {
		result.AddLine("")
		result.Add(blocks[i])
	}

	return result
}

func shouldPrintCRD(d *compare.CRDDiff, breakingOnly bool) bool {
	if breakingOnly {
		return d.HasBreakingChanges()
	}

	return d.HasChanges()
}

func shouldPrintCRDVersion(d compare.CRDVersionDiff, breakingOnly bool) bool {
	if breakingOnly {
		return d.HasBreakingChanges()
	}

	return d.HasChanges()
}

func heading(s, u string, style string) string {
	line := strings.Repeat(u, len(s))

	if style != "" {
		s = colors.Styles[style].Render(s)
		line = colors.Styles[style].Render(line)
	}

	return fmt.Sprintf("%s\n%s", s, line)
}

func printSchemaDiff(diff *diff.SchemaDiff, printer *indent.Indenter) {
	if diff.ExtensionsDiff != nil {
		printer.AddLinef("ExtensionsDiff: %#v", diff.ExtensionsDiff)
	}
	if diff.OneOfDiff != nil {
		printer.AddLinef("OneOfDiff: %#v", diff.OneOfDiff)
	}
	if diff.AnyOfDiff != nil {
		printer.AddLinef("AnyOfDiff: %#v", diff.AnyOfDiff)
	}
	if diff.AllOfDiff != nil {
		printer.AddLinef("AllOfDiff: %#v", diff.AllOfDiff)
	}
	if diff.NotDiff != nil {
		printer.AddLinef("NotDiff: %#v", diff.NotDiff)
	}
	if d := diff.TypeDiff; d != nil {
		printValueDiff("type", d, printer)
	}
	if d := diff.TitleDiff; d != nil {
		printValueDiff("title", d, printer)
	}
	if d := diff.FormatDiff; d != nil {
		printValueDiff("format", d, printer)
	}
	if d := diff.DescriptionDiff; d != nil {
		printValueDiff("description", d, printer)
	}
	if diff.EnumDiff != nil {
		printer.AddLinef("EnumDiff: %#v", diff.EnumDiff)
	}
	if d := diff.DefaultDiff; d != nil {
		printValueDiff("default value", d, printer)
	}
	if d := diff.ExampleDiff; d != nil {
		printValueDiff("example", d, printer)
	}
	if diff.ExternalDocsDiff != nil {
		printer.AddLinef("ExternalDocsDiff: %#v", diff.ExternalDocsDiff)
	}
	if diff.AdditionalPropertiesAllowedDiff != nil {
		printer.AddLinef("AdditionalPropertiesAllowedDiff: %#v", diff.AdditionalPropertiesAllowedDiff)
	}
	if diff.UniqueItemsDiff != nil {
		printer.AddLinef("UniqueItemsDiff: %#v", diff.UniqueItemsDiff)
	}
	if diff.ExclusiveMinDiff != nil {
		printer.AddLinef("ExclusiveMinDiff: %#v", diff.ExclusiveMinDiff)
	}
	if diff.ExclusiveMaxDiff != nil {
		printer.AddLinef("ExclusiveMaxDiff: %#v", diff.ExclusiveMaxDiff)
	}
	if d := diff.NullableDiff; d != nil {
		printValueDiff("nullable", d, printer)
	}
	if diff.ReadOnlyDiff != nil {
		printer.AddLinef("ReadOnlyDiff: %#v", diff.ReadOnlyDiff)
	}
	if diff.WriteOnlyDiff != nil {
		printer.AddLinef("WriteOnlyDiff: %#v", diff.WriteOnlyDiff)
	}
	if diff.AllowEmptyValueDiff != nil {
		printer.AddLinef("AllowEmptyValueDiff: %#v", diff.AllowEmptyValueDiff)
	}
	if diff.XMLDiff != nil {
		printer.AddLinef("XMLDiff: %#v", diff.XMLDiff)
	}
	if diff.DeprecatedDiff != nil {
		printer.AddLinef("DeprecatedDiff: %#v", diff.DeprecatedDiff)
	}
	if d := diff.MinDiff; d != nil {
		printValueDiff("minimum allowed value", d, printer)
	}
	if d := diff.MaxDiff; d != nil {
		printValueDiff("maximum allowed value", d, printer)
	}
	if diff.MultipleOfDiff != nil {
		printer.AddLinef("MultipleOfDiff: %#v", diff.MultipleOfDiff)
	}
	if d := diff.MinLengthDiff; d != nil {
		printValueDiff("minimum required length", d, printer)
	}
	if d := diff.MaxLengthDiff; d != nil {
		printValueDiff("maximum allowed length", d, printer)
	}
	if d := diff.PatternDiff; d != nil {
		printValueDiff("pattern", d, printer)
	}
	if d := diff.MinItemsDiff; d != nil {
		printValueDiff("minimum required items", d, printer)
	}
	if d := diff.MaxItemsDiff; d != nil {
		printValueDiff("maximum allowed items", d, printer)
	}
	if d := diff.RequiredDiff; d != nil {
		if items := d.Added; len(items) > 0 {
			printer.AddLinef("~ %s %v", colors.ActionChange.Render("requires"), colors.Property.Render(items))
		}

		if items := d.Deleted; len(items) > 0 {
			printer.AddLinef("~ %s %v", colors.ActionChange.Render("unrequires"), colors.Property.Render(items))
		}
	}
	if diff.MinPropsDiff != nil {
		printer.AddLinef("MinPropsDiff: %#v", diff.MinPropsDiff)
	}
	if diff.MaxPropsDiff != nil {
		printer.AddLinef("MaxPropsDiff: %#v", diff.MaxPropsDiff)
	}
	if diff.DiscriminatorDiff != nil {
		printer.AddLinef("DiscriminatorDiff: %#v", diff.DiscriminatorDiff)
	}
}

func isEmpty(s string) bool {
	return s == "" || s == "<nil>"
}

func printValueDiff(attribute string, diff *diff.ValueDiff, printer *indent.Indenter) {
	from := fmt.Sprintf("%v", diff.From)

	if isEmpty(from) {
		printer.AddLinef(
			"~ %s %s to %s",
			colors.ActionChange.Render("set"),
			colors.Attribute.Render(attribute),
			colors.NewValue.Render(diff.To),
		)
	} else {
		to := fmt.Sprintf("%v", diff.To)

		if isEmpty(to) {
			printer.AddLinef(
				"~ %s %s",
				colors.ActionChange.Render("removed"),
				colors.Attribute.Render(attribute),
			)
		} else {
			printer.AddLinef(
				"~ %s %s from %s to %s",
				colors.ActionChange.Render("changed"),
				colors.Attribute.Render(attribute),
				colors.OldValue.Render(diff.From),
				colors.NewValue.Render(diff.To),
			)
		}
	}
}
