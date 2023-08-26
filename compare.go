// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/crdiff/pkg/compare"
	"go.xrstf.de/crdiff/pkg/compare/report"
	"go.xrstf.de/crdiff/pkg/crd"
)

func compareCRDs(log logrus.FieldLogger, baseCRDs, revisionCRDs map[string]crd.CRD, diffOpt compare.CompareOptions) (*report.Report, error) {
	report := &report.Report{
		Diffs: map[string]compare.CRDDiff{},
	}

	for crdIdentifier, baseCRD := range baseCRDs {
		revisionCRD, exists := revisionCRDs[crdIdentifier]
		if !exists {
			report.Diffs[crdIdentifier] = compare.CRDDiff{
				General: []compare.Change{{
					Breaking:    true,
					Description: "CRD has been removed",
				}},
			}

			continue
		}

		crdChanges, err := compare.CompareCRDs(baseCRD, revisionCRD, diffOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to compare %q: %w", crdIdentifier, err)
		}

		if crdChanges.HasChanges() {
			report.Diffs[crdIdentifier] = *crdChanges
		}
	}

	return report, nil
}

func outputReport(log logrus.FieldLogger, report *report.Report, breakingOnly bool, opts *commonCompareOptions) {
	switch opts.output {
	case outputFormatText:
		report.Print(breakingOnly)
	case outputFormatJSON:
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			log.Errorf("Failed to render output as JSON: %v", err)
		}
	default:
		log.Errorf("This should never happen: Do not know how to handle %s output format.", opts.output)
	}
}
