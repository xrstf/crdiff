// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"strings"

	oasdiffdiff "github.com/tufin/oasdiff/diff"

	"go.xrstf.de/crdiff/pkg/colors"
	"go.xrstf.de/crdiff/pkg/diff"

	"k8s.io/apimachinery/pkg/util/sets"
)

type DiffReport struct {
	Diffs map[string]diff.CRDDiff
}

func (r *DiffReport) Empty() bool {
	if r == nil {
		return true
	}

	for _, change := range r.Diffs {
		if !change.Empty() {
			return false
		}
	}

	return true
}

func (r *DiffReport) Print() {
	sortedIdentifiers := sets.List(sets.KeySet(r.Diffs))

	for _, crdIdentifier := range sortedIdentifiers {
		crdChanges := r.Diffs[crdIdentifier]

		if crdChanges.Empty() {
			continue
		}

		heading(crdIdentifier, "=", 0, "crd")
		fmt.Println("")

		extraLine := false
		if len(crdChanges.General) > 0 {
			extraLine = true
			for _, change := range crdChanges.General {
				fmt.Printf("  ~ %s\n", change.Modification)
			}
		}

		for _, version := range sets.List(crdChanges.Versions.AddedVersions) {
			extraLine = true
			fmt.Printf("  + %s %s\n", colors.ActionAdd.Render("added"), colors.Version.Render(version))
		}

		for _, version := range sets.List(crdChanges.Versions.DeletedVersions) {
			extraLine = true
			fmt.Printf("  - %s %s\n", colors.ActionRemove.Render("removed"), colors.Version.Render(version))
		}

		if extraLine {
			fmt.Println("")
		}

		changedVersions := sets.List(sets.KeySet(crdChanges.Versions.ChangedVersions))
		for _, version := range changedVersions {
			heading(version, "-", 2, "version")

			versionSchemasDiff := crdChanges.Versions.ChangedVersions[version]
			changedPaths := sets.List(sets.KeySet(versionSchemasDiff))

			for _, path := range changedPaths {
				pathChanges := versionSchemasDiff[path]

				fmt.Printf("    > %s:\n", colors.Path.Render(path))

				for _, field := range sets.List(pathChanges.AddedProperties) {
					fmt.Printf("      + %s %s\n", colors.ActionAdd.Render("added"), colors.Property.Render(field))
				}

				for _, field := range sets.List(pathChanges.DeletedProperties) {
					fmt.Printf("      - %s %s\n", colors.ActionRemove.Render("removed"), colors.Property.Render(field))
				}

				if d := pathChanges.Diff; d != nil {
					printSchemaDiff(d, "      ")
				}

				fmt.Println("")
			}

			fmt.Println("")
		}
	}
}

func heading(s, u string, padding int, style string) {
	pad := strings.Repeat(" ", padding)
	line := strings.Repeat(u, len(s))

	if style != "" {
		s = colors.Styles[style].Render(s)
		line = colors.Styles[style].Render(line)
	}

	fmt.Println(pad + s)
	fmt.Println(pad + line)
}

func printSchemaDiff(diff *oasdiffdiff.SchemaDiff, padding string) {
	if diff.ExtensionsDiff != nil {
		fmt.Printf("%sExtensionsDiff: %#v\n", padding, diff.ExtensionsDiff)
	}
	if diff.OneOfDiff != nil {
		fmt.Printf("%sOneOfDiff: %#v\n", padding, diff.OneOfDiff)
	}
	if diff.AnyOfDiff != nil {
		fmt.Printf("%sAnyOfDiff: %#v\n", padding, diff.AnyOfDiff)
	}
	if diff.AllOfDiff != nil {
		fmt.Printf("%sAllOfDiff: %#v\n", padding, diff.AllOfDiff)
	}
	if diff.NotDiff != nil {
		fmt.Printf("%sNotDiff: %#v\n", padding, diff.NotDiff)
	}
	if d := diff.TypeDiff; d != nil {
		printValueDiff("type", d, padding)
	}
	if d := diff.TitleDiff; d != nil {
		printValueDiff("title", d, padding)
	}
	if d := diff.FormatDiff; d != nil {
		printValueDiff("format", d, padding)
	}
	if d := diff.DescriptionDiff; d != nil {
		printValueDiff("description", d, padding)
	}
	if diff.EnumDiff != nil {
		fmt.Printf("%sEnumDiff: %#v\n", padding, diff.EnumDiff)
	}
	if d := diff.DefaultDiff; d != nil {
		printValueDiff("default value", d, padding)
	}
	if d := diff.ExampleDiff; d != nil {
		printValueDiff("example", d, padding)
	}
	if diff.ExternalDocsDiff != nil {
		fmt.Printf("%sExternalDocsDiff: %#v\n", padding, diff.ExternalDocsDiff)
	}
	if diff.AdditionalPropertiesAllowedDiff != nil {
		fmt.Printf("%sAdditionalPropertiesAllowedDiff: %#v\n", padding, diff.AdditionalPropertiesAllowedDiff)
	}
	if diff.UniqueItemsDiff != nil {
		fmt.Printf("%sUniqueItemsDiff: %#v\n", padding, diff.UniqueItemsDiff)
	}
	if diff.ExclusiveMinDiff != nil {
		fmt.Printf("%sExclusiveMinDiff: %#v\n", padding, diff.ExclusiveMinDiff)
	}
	if diff.ExclusiveMaxDiff != nil {
		fmt.Printf("%sExclusiveMaxDiff: %#v\n", padding, diff.ExclusiveMaxDiff)
	}
	if d := diff.NullableDiff; d != nil {
		printValueDiff("nullable", d, padding)
	}
	if diff.ReadOnlyDiff != nil {
		fmt.Printf("%sReadOnlyDiff: %#v\n", padding, diff.ReadOnlyDiff)
	}
	if diff.WriteOnlyDiff != nil {
		fmt.Printf("%sWriteOnlyDiff: %#v\n", padding, diff.WriteOnlyDiff)
	}
	if diff.AllowEmptyValueDiff != nil {
		fmt.Printf("%sAllowEmptyValueDiff: %#v\n", padding, diff.AllowEmptyValueDiff)
	}
	if diff.XMLDiff != nil {
		fmt.Printf("%sXMLDiff: %#v\n", padding, diff.XMLDiff)
	}
	if diff.DeprecatedDiff != nil {
		fmt.Printf("%sDeprecatedDiff: %#v\n", padding, diff.DeprecatedDiff)
	}
	if d := diff.MinDiff; d != nil {
		printValueDiff("minimum allowed value", d, padding)
	}
	if d := diff.MaxDiff; d != nil {
		printValueDiff("maximum allowed value", d, padding)
	}
	if diff.MultipleOfDiff != nil {
		fmt.Printf("%sMultipleOfDiff: %#v\n", padding, diff.MultipleOfDiff)
	}
	if d := diff.MinLengthDiff; d != nil {
		printValueDiff("minimum required length", d, padding)
	}
	if d := diff.MaxLengthDiff; d != nil {
		printValueDiff("maximum allowed length", d, padding)
	}
	if d := diff.PatternDiff; d != nil {
		printValueDiff("pattern", d, padding)
	}
	if d := diff.MinItemsDiff; d != nil {
		printValueDiff("minimum required items", d, padding)
	}
	if d := diff.MaxItemsDiff; d != nil {
		printValueDiff("maximum allowed items", d, padding)
	}
	if d := diff.RequiredDiff; d != nil {
		if items := d.Added; len(items) > 0 {
			fmt.Printf(
				"%s~ %s %v\n",
				padding,
				colors.ActionChange.Render("requires"),
				colors.Property.Render(items),
			)
		}

		if items := d.Deleted; len(items) > 0 {
			fmt.Printf(
				"%s~ %s %v\n",
				padding,
				colors.ActionChange.Render("unrequires"),
				colors.Property.Render(items),
			)
		}
	}
	if diff.MinPropsDiff != nil {
		fmt.Printf("%sMinPropsDiff: %#v\n", padding, diff.MinPropsDiff)
	}
	if diff.MaxPropsDiff != nil {
		fmt.Printf("%sMaxPropsDiff: %#v\n", padding, diff.MaxPropsDiff)
	}
	if diff.AdditionalPropertiesDiff != nil {
		fmt.Printf("%sAdditionalPropertiesDiff: %#v\n", padding, diff.AdditionalPropertiesDiff)
	}
	if diff.DiscriminatorDiff != nil {
		fmt.Printf("%sDiscriminatorDiff: %#v\n", padding, diff.DiscriminatorDiff)
	}
}

func isEmpty(s string) bool {
	return s == "" || s == "<nil>"
}

func printValueDiff(attribute string, diff *oasdiffdiff.ValueDiff, padding string) {
	from := fmt.Sprintf("%v", diff.From)

	if isEmpty(from) {
		fmt.Printf(
			"%s~ %s %s to %s\n",
			padding,
			colors.ActionChange.Render("set"),
			colors.Attribute.Render(attribute),
			colors.NewValue.Render(diff.To),
		)
	} else {
		to := fmt.Sprintf("%v", diff.To)

		if isEmpty(to) {
			fmt.Printf(
				"%s~ %s %s\n",
				padding,
				colors.ActionChange.Render("removed"),
				colors.Attribute.Render(attribute),
			)
		} else {
			fmt.Printf(
				"%s~ %s %s from %s to %s\n",
				padding,
				colors.ActionChange.Render("changed"),
				colors.Attribute.Render(attribute),
				colors.OldValue.Render(diff.From),
				colors.NewValue.Render(diff.To),
			)
		}
	}
}
