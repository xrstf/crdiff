// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"os"

	"github.com/gookit/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	verbose    bool
	forceColor bool
	noColor    bool
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
		if err := c.Usage(); err != nil {
			return err
		}

		// ensure we exit with code 1 later on
		return err
	})

	rootCmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&opts.forceColor, "color", false, "enable colored output (can also use $FORCE_COLOR)")
	rootCmd.PersistentFlags().BoolVar(&opts.noColor, "no-color", false, "disable colored output (can also use $NO_COLOR)")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if opts.forceColor && opts.noColor {
			return errors.New("cannot combine --no-color with --color")
		}

		if opts.forceColor {
			color.Enable = true
		} else if opts.noColor {
			color.Enable = false
		}

		logger := logrus.New()
		if opts.verbose {
			logger.SetLevel(logrus.DebugLevel)
		}
		log = logger

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
