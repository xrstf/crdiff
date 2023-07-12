// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/diff"
	"go.xrstf.de/crdiff/pkg/loader"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/sirupsen/logrus"
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

		log.Debug("Comparing CRDs…")
		report, err := compareCRDs(log, baseCRDs, revisionCRDs)
		if err != nil {
			return fmt.Errorf("failed comparing CRDs: %v", err)
		}

		if report.Empty() {
			log.Info("No changes detected.")
			return nil
		}

		sortedIdentifiers := sets.List(sets.KeySet(report.Changes))

		for _, crdIdentifier := range sortedIdentifiers {
			crdChanges := report.Changes[crdIdentifier]

			if crdChanges.Empty() {
				continue
			}

			heading(crdIdentifier, "=", 0)

			if len(crdChanges.Unversioned) > 0 {
				for _, change := range crdChanges.Unversioned {
					fmt.Printf("  * %s\n", change)
				}
			}
			fmt.Println("")

			for version, versionChanges := range crdChanges.Versioned {
				heading(version, "-", 2)

				for _, change := range versionChanges {
					fmt.Printf("    * %s\n", change)
				}
				fmt.Println("")
			}
		}

		return nil
	})
}

func compareCRDs(log logrus.FieldLogger, baseCRDs, revisionCRDs map[string]crd.CRD) (*DiffReport, error) {
	report := &DiffReport{
		Changes: map[string]diff.Diff{},
	}
	diffOpt := diff.Options{}

	for crdIdentifier, baseCRD := range baseCRDs {
		revisionCRD, exists := revisionCRDs[crdIdentifier]
		if !exists {
			report.Changes[crdIdentifier] = diff.Diff{
				Unversioned: []diff.Change{
					diff.Change("CRD has been removed"),
				},
			}

			continue
		}

		crdChanges, err := diff.CompareCRDs(baseCRD, revisionCRD, diffOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to compare %q: %w", crdIdentifier, err)
		}

		if crdChanges != nil {
			report.Changes[crdIdentifier] = *crdChanges
		}
	}

	return report, nil
}
