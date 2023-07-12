// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type globalOptions struct {
	verbose bool
}

type cobraFuncE func(cmd *cobra.Command, args []string) error

func main() {
	opts := globalOptions{}
	ctx := context.Background() // signals.SetupSignalHandler()

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

	rootCmd.AddCommand(
		DiffCommand(&opts),
	)

	// we don't need this
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

const loggerCtxKey = iota

func handleErrors(action cobraFuncE) cobraFuncE {
	return func(cmd *cobra.Command, args []string) error {
		err := action(cmd, args)
		if err != nil {
			log := cmd.Context().Value(loggerCtxKey).(logrus.FieldLogger)
			log.Errorf("Operation failed: %v.", err)
		}

		return err
	}
}

func createLogger(cmd *cobra.Command, globalOpts *globalOptions) *logrus.Logger {
	logger := logrus.New()
	if globalOpts.verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	cmd.SetContext(context.WithValue(cmd.Context(), loggerCtxKey, logger))

	return logger
}

func heading(s, u string, padding int) {
	pad := strings.Repeat(" ", padding)

	fmt.Println(pad + s)
	fmt.Println(pad + strings.Repeat(u, len(s)))
}
