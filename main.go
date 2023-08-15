// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Project build specific vars
var (
	Tag    string
	Commit string
)

func printVersion() {
	fmt.Printf(
		"version: %s\nbuilt with: %s\ntag: %s\ncommit: %s\n",
		strings.TrimPrefix(Tag, "v"),
		runtime.Version(),
		Tag,
		Commit,
	)
}

const (
	outputFormatText = "text"
	outputFormatJSON = "json"
)

type globalOptions struct {
	verbose    bool
	forceColor bool
	noColor    bool
	output     string
}

type cobraFuncE func(cmd *cobra.Command, args []string) error

var (
	log logrus.FieldLogger
)

func main() {
	opts := globalOptions{}
	ctx := context.Background()

	rootCmd := &cobra.Command{
		Use:           "crdiff",
		Short:         "Compare Kubernetes CRDs",
		SilenceErrors: true,
	}

	// cobra does not make any distinction between "error that happened because of bad flags"
	// and "error that happens because of something going bad inside the RunE function", and
	// so would always show the Usage, no matter what error occurred. To work around this, we
	// set SilenceUsages on all commands and manually print the error using the FlagErrorFunc.
	rootCmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		fmt.Println(err)
		if err := c.Usage(); err != nil {
			return err
		}

		// ensure we exit with code 1 later on
		return err
	})

	rootCmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&opts.forceColor, "color", false, "enable colored output (can also use $FORCE_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&opts.noColor, "no-color", false, "disable colored output (can also use $NO_COLOR)")
	rootCmd.PersistentFlags().StringVarP(&opts.output, "output", "o", "text", "output format (one of [text, json])")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// We silence errors so we can have full control over their handling,
		// but this means nobody is processing the error returned from this
		// function; so instead we print it ourselves and just return it to
		// make cobra exit with 1.
		fail := func(err error) error {
			log.Errorf("Invalid flags: %v.", err)
			return err
		}

		// init logger as early as possible, as any error handling depends on it
		logger := logrus.New()
		if opts.verbose {
			logger.SetLevel(logrus.DebugLevel)
		}
		log = logger

		switch opts.output {
		case outputFormatText:
			// NOP
		case outputFormatJSON:
			logger.SetFormatter(&logrus.JSONFormatter{})
		default:
			return fail(fmt.Errorf("unknown output format %q", opts.output))
		}

		if opts.forceColor && opts.noColor {
			return fail(errors.New("cannot combine --no-color with --color"))
		}

		if opts.forceColor {
			color.Enable = true
		} else if opts.noColor {
			color.Enable = false
		}

		return nil
	}

	rootCmd.AddCommand(
		DiffCommand(&opts),
		BreakingCommand(&opts),
	)

	// we don't need this
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func handleErrors(action cobraFuncE) cobraFuncE {
	return func(cmd *cobra.Command, args []string) error {
		err := action(cmd, args)
		if err != nil {
			log.Errorf("Operation failed: %v.", err)
		}

		return err
	}
}
