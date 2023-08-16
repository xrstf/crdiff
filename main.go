// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type globalOptions struct {
	verbose bool
}

func (o *globalOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.verbose, "verbose", "v", o.verbose, "enable verbose logging")
}

type cobraFuncE func(cmd *cobra.Command, args []string) error

var (
	log *logrus.Logger
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

	opts.AddFlags(rootCmd.PersistentFlags())

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		fmt.Println("rootCmd.PersistentPreRun")
		// init logger as early as possible, as any error handling depends on it
		log = logrus.New()
		if opts.verbose {
			log.SetLevel(logrus.DebugLevel)
		}
	}

	rootCmd.AddCommand(
		DiffCommand(&opts),
		BreakingCommand(&opts),
		VersionCommand(&opts),
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
