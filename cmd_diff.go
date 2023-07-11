// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"go.xrstf.de/crdiff/pkg/diff"
	"go.xrstf.de/crdiff/pkg/loader"

	"github.com/spf13/cobra"
)

type isRunningOptions struct{}

func DiffCommand(globalOpts *globalOptions) *cobra.Command {
	cmdOpts := isRunningOptions{}

	cmd := &cobra.Command{
		Use:          "diff BASE REVISION",
		Short:        "Compare two or more CRD files/directories and print the differences",
		RunE:         DiffRunE(globalOpts, &cmdOpts),
		SilenceUsage: true,
	}

	return cmd
}

type filesChanges map[string]diff.Diff

func DiffRunE(globalOpts *globalOptions, cmdOpts *isRunningOptions) cobraFuncE {
	return handleErrors(func(cmd *cobra.Command, args []string) error {
		log := createLogger(cmd, globalOpts)

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

		diffs := filesChanges{}
		diffOpt := diff.Options{}

		log.Debug("Comparing CRDs…")
		for identifier, baseCRD := range baseCRDs {

			revisionCRD, exists := revisionCRDs[identifier]
			if !exists {
				diffs[identifier] = diff.Diff{
					"": []diff.Change{
						diff.Change("CRD has been removed"),
					},
				}

				continue
			}

			crdChanges, err := diff.CompareCRDs(baseCRD, revisionCRD, diffOpt)
			if err != nil {
				return fmt.Errorf("failed to compare %q: %w", identifier, err)
			}

			if crdChanges != nil {
				diffs[identifier] = crdChanges
			}
		}

		if len(diffs) == 0 {
			log.Info("No changes detected.")
			return nil
		}

		for identifier, changesPerVersion := range diffs {
			fmt.Println(identifier)

			nonVersionChanges, exist := changesPerVersion[""]
			if exist && len(nonVersionChanges) > 0 {
				for _, change := range nonVersionChanges {
					fmt.Printf("  * %s\n", change)
				}
				fmt.Println("")
			}

			delete(changesPerVersion, "")
			for version, versionChanges := range changesPerVersion {
				fmt.Printf("  %s\n", version)

				for _, change := range versionChanges {
					fmt.Printf("    * %s\n", change)
				}
				fmt.Println("")
			}
		}

		return nil
	})
}
