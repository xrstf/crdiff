// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.xrstf.de/crdiff/pkg/diff"
	"go.xrstf.de/crdiff/pkg/loader"
)

type breakingCmdOptions struct{}

func BreakingCommand(globalOpts *globalOptions) *cobra.Command {
	cmdOpts := breakingCmdOptions{}

	cmd := &cobra.Command{
		Use:          "breaking BASE REVISION",
		Short:        "Compare two or more CRD files/directories and print all breaking differences",
		RunE:         BreakingRunE(globalOpts, &cmdOpts),
		SilenceUsage: true,
	}

	return cmd
}

func BreakingRunE(globalOpts *globalOptions, cmdOpts *breakingCmdOptions) cobraFuncE {
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
		diffOpt := diff.Options{
			BreakingOnly: true,
		}
		report, err := compareCRDs(log, baseCRDs, revisionCRDs, diffOpt)
		if err != nil {
			return fmt.Errorf("failed comparing CRDs: %v", err)
		}

		if report.Empty() {
			log.Info("No changes detected.")
			return nil
		}

		report.Print()

		return nil
	})
}
