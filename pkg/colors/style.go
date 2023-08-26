// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package colors

import "github.com/gookit/color"

var (
	CRD                    = color.New(color.FgLightCyan)
	Version                = color.New(color.FgLightMagenta)
	BreakingChangesHeading = color.New(color.FgRed)
	Path                   = color.New(color.FgLightYellow)
	Property               = color.New(color.FgLightYellow)
	Attribute              = color.New(color.FgLightGreen)
	Action                 = color.New(color.FgYellow)
	ActionAdd              = color.New(color.Green)
	ActionChange           = color.New(color.Yellow)
	ActionRemove           = color.New(color.Red)
	OldValue               = color.New(color.FgGreen)
	NewValue               = color.New(color.FgLightGreen)

	Styles = map[string]color.Style{
		"crd":                      CRD,
		"version":                  Version,
		"breaking-changes-heading": BreakingChangesHeading,
		"path":                     Path,
		"property":                 Property,
		"attribute":                Attribute,
		"action":                   Action,
		"action-add":               ActionAdd,
		"action-change":            ActionChange,
		"action-remove":            ActionRemove,
		"old-value":                OldValue,
		"new-value":                NewValue,
	}
)
