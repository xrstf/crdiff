// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"errors"
	"fmt"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	outputFormatText = "text"
	outputFormatJSON = "json"
)

type commonCompareOptions struct {
	forceColor         bool
	noColor            bool
	output             string
	ignoreDescriptions bool
}

func (o *commonCompareOptions) PreRunE(cmd *cobra.Command, args []string) error {
	// We silence errors so we can have full control over their handling,
	// but this means nobody is processing the error returned from this
	// function; so instead we print it ourselves and just return it to
	// make cobra exit with 1.
	fail := func(err error) error {
		log.Errorf("Invalid flags: %v.", err)
		return err
	}

	// set the log format on the global log variable
	switch o.output {
	case outputFormatText:
		// NOP
	case outputFormatJSON:
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		return fail(fmt.Errorf("unknown output format %q", o.output))
	}

	// configure gookit
	if o.forceColor && o.noColor {
		return fail(errors.New("cannot combine --no-color with --color"))
	}

	if o.forceColor {
		color.Enable = true
	} else if o.noColor {
		color.Enable = false
	}

	return nil
}

func (o *commonCompareOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.forceColor, "color", o.forceColor, "enable colored output (can also use $FORCE_COLOR)")
	fs.BoolVar(&o.noColor, "no-color", o.noColor, "disable colored output (can also use $NO_COLOR)")
	fs.BoolVar(&o.ignoreDescriptions, "ignore-descriptions", o.ignoreDescriptions, "ignore changes to field descriptions")
	fs.StringVarP(&o.output, "output", "o", o.output, "output format (one of [text, json])")
}
