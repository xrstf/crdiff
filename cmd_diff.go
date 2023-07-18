// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.xrstf.de/crdiff/pkg/diff"
	"go.xrstf.de/crdiff/pkg/loader"
)

type diffCmdOptions struct{}

func DiffCommand(globalOpts *globalOptions) *cobra.Command {
	cmdOpts := diffCmdOptions{}

	cmd := &cobra.Command{
		Use:          "diff BASE REVISION",
		Short:        "Compare two or more CRD files/directories and print the differences",
		RunE:         DiffRunE(globalOpts, &cmdOpts),
		SilenceUsage: true,
	}

	return cmd
}

func DiffRunE(globalOpts *globalOptions, cmdOpts *diffCmdOptions) cobraFuncE {
	return handleErrors(func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return cmd.Help()
		}

		loadOpts := loader.NewDefaultOptions()

		log.Debug("Loading base CRDs…")
		baseCRDs, err := loader.LoadCRDs(args[0], loadOpts, log)
		if err != nil {
			return fmt.Errorf("failed loading base CRDs: %v", err)
		}

		log.Debug("Loading revision CRDs…")
		revisionCRDs, err := loader.LoadCRDs(args[1], loadOpts, log)
		if err != nil {
			return fmt.Errorf("failed loading revision CRDs: %v", err)
		}

		log.Debug("Comparing CRDs…")
		diffOpt := diff.Options{}
		report, err := compareCRDs(log, baseCRDs, revisionCRDs, diffOpt)
		if err != nil {
			return fmt.Errorf("failed comparing CRDs: %v", err)
		}

		if report.Empty() {
			log.Info("No changes detected.")
			// do not return, still print the report on stdout so we still
			// produce valid JSON in case --output=json is given.
			// return nil
		}

		outputReport(log, report, globalOpts.output)

		return nil
	})
}
