// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"go.xrstf.de/crdiff/pkg/compare"
	"go.xrstf.de/crdiff/pkg/loader"
)

type breakingCmdOptions struct {
	common commonCompareOptions
}

func (o *breakingCmdOptions) AddFlags(fs *pflag.FlagSet) {
	o.common.AddFlags(fs)
}

func BreakingCommand(globalOpts *globalOptions) *cobra.Command {
	cmdOpts := breakingCmdOptions{
		common: commonCompareOptions{
			output: outputFormatText,
		},
	}

	cmd := &cobra.Command{
		Use:          "breaking BASE REVISION",
		Short:        "Compare two or more CRD files/directories and print all breaking differences",
		RunE:         BreakingRunE(globalOpts, &cmdOpts),
		SilenceUsage: true,
	}

	cmdOpts.AddFlags(cmd.PersistentFlags())

	cmd.PreRunE = cmdOpts.common.PreRunE

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
		diffOpt := compare.CompareOptions{
			BreakingOnly:       true,
			IgnoreDescriptions: cmdOpts.common.ignoreDescriptions,
		}
		report, err := compareCRDs(log, baseCRDs, revisionCRDs, diffOpt)
		if err != nil {
			return fmt.Errorf("failed comparing CRDs: %v", err)
		}

		if !report.HasChanges() {
			log.Info("No changes detected.")
			// do not return, still print the report on stdout so we still
			// produce valid JSON in case --output=json is given.
			// return nil
		}

		outputReport(log, report, true, &cmdOpts.common)

		return nil
	})
}
