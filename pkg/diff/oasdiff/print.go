// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package oasdiff

import (
	"fmt"

	"github.com/tufin/oasdiff/diff"
)

var prefix = ""

func indent() {
	prefix = prefix + "  "
}

func dedent() {
	prefix = prefix[:len(prefix)-2]
}

func PrintSchemaDiff(diff *diff.SchemaDiff) {
	indent()
	defer dedent()

	if diff.SchemaAdded {
		fmt.Printf("%sSchemaAdded: %#v\n", prefix, diff.SchemaAdded)
	}
	if diff.SchemaDeleted {
		fmt.Printf("%sSchemaDeleted: %#v\n", prefix, diff.SchemaDeleted)
	}
	if diff.CircularRefDiff {
		fmt.Printf("%sCircularRefDiff: %#v\n", prefix, diff.CircularRefDiff)
	}
	if diff.ExtensionsDiff != nil {
		fmt.Printf("%sExtensionsDiff: %#v\n", prefix, diff.ExtensionsDiff)
	}
	if diff.OneOfDiff != nil {
		fmt.Printf("%sOneOfDiff: %#v\n", prefix, diff.OneOfDiff)
	}
	if diff.AnyOfDiff != nil {
		fmt.Printf("%sAnyOfDiff: %#v\n", prefix, diff.AnyOfDiff)
	}
	if diff.AllOfDiff != nil {
		fmt.Printf("%sAllOfDiff: %#v\n", prefix, diff.AllOfDiff)
	}
	if diff.NotDiff != nil {
		fmt.Printf("%sNotDiff: %#v\n", prefix, diff.NotDiff)
	}
	if diff.TypeDiff != nil {
		fmt.Printf("%sTypeDiff: %#v\n", prefix, diff.TypeDiff)
	}
	if diff.TitleDiff != nil {
		fmt.Printf("%sTitleDiff: %#v\n", prefix, diff.TitleDiff)
	}
	if diff.FormatDiff != nil {
		fmt.Printf("%sFormatDiff: %#v\n", prefix, diff.FormatDiff)
	}
	if diff.DescriptionDiff != nil {
		fmt.Printf("%sDescriptionDiff:\n", prefix)
		printValueDiff(diff.DescriptionDiff)
	}
	if diff.EnumDiff != nil {
		fmt.Printf("%sEnumDiff: %#v\n", prefix, diff.EnumDiff)
	}
	if diff.DefaultDiff != nil {
		fmt.Printf("%sDefaultDiff: %#v\n", prefix, diff.DefaultDiff)
	}
	if diff.ExampleDiff != nil {
		fmt.Printf("%sExampleDiff: %#v\n", prefix, diff.ExampleDiff)
	}
	if diff.ExternalDocsDiff != nil {
		fmt.Printf("%sExternalDocsDiff: %#v\n", prefix, diff.ExternalDocsDiff)
	}
	if diff.AdditionalPropertiesAllowedDiff != nil {
		fmt.Printf("%sAdditionalPropertiesAllowedDiff: %#v\n", prefix, diff.AdditionalPropertiesAllowedDiff)
	}
	if diff.UniqueItemsDiff != nil {
		fmt.Printf("%sUniqueItemsDiff: %#v\n", prefix, diff.UniqueItemsDiff)
	}
	if diff.ExclusiveMinDiff != nil {
		fmt.Printf("%sExclusiveMinDiff: %#v\n", prefix, diff.ExclusiveMinDiff)
	}
	if diff.ExclusiveMaxDiff != nil {
		fmt.Printf("%sExclusiveMaxDiff: %#v\n", prefix, diff.ExclusiveMaxDiff)
	}
	if diff.NullableDiff != nil {
		fmt.Printf("%sNullableDiff: %#v\n", prefix, diff.NullableDiff)
	}
	if diff.ReadOnlyDiff != nil {
		fmt.Printf("%sReadOnlyDiff: %#v\n", prefix, diff.ReadOnlyDiff)
	}
	if diff.WriteOnlyDiff != nil {
		fmt.Printf("%sWriteOnlyDiff: %#v\n", prefix, diff.WriteOnlyDiff)
	}
	if diff.AllowEmptyValueDiff != nil {
		fmt.Printf("%sAllowEmptyValueDiff: %#v\n", prefix, diff.AllowEmptyValueDiff)
	}
	if diff.XMLDiff != nil {
		fmt.Printf("%sXMLDiff: %#v\n", prefix, diff.XMLDiff)
	}
	if diff.DeprecatedDiff != nil {
		fmt.Printf("%sDeprecatedDiff: %#v\n", prefix, diff.DeprecatedDiff)
	}
	if diff.MinDiff != nil {
		fmt.Printf("%sMinDiff: %#v\n", prefix, diff.MinDiff)
	}
	if diff.MaxDiff != nil {
		fmt.Printf("%sMaxDiff: %#v\n", prefix, diff.MaxDiff)
	}
	if diff.MultipleOfDiff != nil {
		fmt.Printf("%sMultipleOfDiff: %#v\n", prefix, diff.MultipleOfDiff)
	}
	if diff.MinLengthDiff != nil {
		fmt.Printf("%sMinLengthDiff: %#v\n", prefix, diff.MinLengthDiff)
	}
	if diff.MaxLengthDiff != nil {
		fmt.Printf("%sMaxLengthDiff: %#v\n", prefix, diff.MaxLengthDiff)
	}
	if diff.PatternDiff != nil {
		fmt.Printf("%sPatternDiff: %#v\n", prefix, diff.PatternDiff)
	}
	if diff.MinItemsDiff != nil {
		fmt.Printf("%sMinItemsDiff: %#v\n", prefix, diff.MinItemsDiff)
	}
	if diff.MaxItemsDiff != nil {
		fmt.Printf("%sMaxItemsDiff: %#v\n", prefix, diff.MaxItemsDiff)
	}
	if diff.ItemsDiff != nil {
		fmt.Printf("%sItemsDiff:\n", prefix)
		PrintSchemaDiff(diff.ItemsDiff)
	}
	if diff.RequiredDiff != nil {
		fmt.Printf("%sRequiredDiff: %#v\n", prefix, diff.RequiredDiff)
	}
	if diff.PropertiesDiff != nil {
		fmt.Printf("%sPropertiesDiff:\n", prefix)
		PrintSchemasDiff(diff.PropertiesDiff)
	}
	if diff.MinPropsDiff != nil {
		fmt.Printf("%sMinPropsDiff: %#v\n", prefix, diff.MinPropsDiff)
	}
	if diff.MaxPropsDiff != nil {
		fmt.Printf("%sMaxPropsDiff: %#v\n", prefix, diff.MaxPropsDiff)
	}
	if diff.AdditionalPropertiesDiff != nil {
		fmt.Printf("%sAdditionalPropertiesDiff: %#v\n", prefix, diff.AdditionalPropertiesDiff)
	}
	if diff.DiscriminatorDiff != nil {
		fmt.Printf("%sDiscriminatorDiff: %#v\n", prefix, diff.DiscriminatorDiff)
	}
}

func PrintSchemasDiff(diff *diff.SchemasDiff) {
	indent()
	defer dedent()

	if len(diff.Added) > 0 {
		fmt.Printf("%sAdded:\n", prefix)
		indent()
		for _, v := range diff.Added {
			fmt.Printf("%s- %s\n", prefix, v)
		}
		dedent()
	}

	if len(diff.Deleted) > 0 {
		fmt.Printf("%sDeleted:\n", prefix)
		indent()
		for _, v := range diff.Deleted {
			fmt.Printf("%s- %s\n", prefix, v)
		}
		dedent()
	}

	if len(diff.Modified) > 0 {
		fmt.Printf("%sModified\n", prefix)
		PrintModifiedSchemas(diff.Modified)
	}
}

func PrintModifiedSchemas(diff diff.ModifiedSchemas) {
	indent()
	defer dedent()

	for k, v := range diff {
		fmt.Printf("%s%s:\n", prefix, k)
		PrintSchemaDiff(v)
	}
}

func printValueDiff(diff *diff.ValueDiff) {
	indent()
	defer dedent()

	fmt.Printf("%sFrom: %#v\n", prefix, diff.From)
	fmt.Printf("%s  To: %#v\n", prefix, diff.To)
}
