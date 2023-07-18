// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/diff"
	"go.xrstf.de/crdiff/pkg/diff/report"
)

func compareCRDs(log logrus.FieldLogger, baseCRDs, revisionCRDs map[string]crd.CRD, diffOpt diff.Options) (*report.Report, error) {
	report := &report.Report{
		Diffs: map[string]diff.CRDDiff{},
	}

	for crdIdentifier, baseCRD := range baseCRDs {
		revisionCRD, exists := revisionCRDs[crdIdentifier]
		if !exists {
			report.Diffs[crdIdentifier] = diff.CRDDiff{
				General: []diff.Change{{
					Description: "CRD has been removed",
				}},
			}

			continue
		}

		crdChanges, err := diff.CompareCRDs(baseCRD, revisionCRD, diffOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to compare %q: %w", crdIdentifier, err)
		}

		if !crdChanges.Empty() {
			report.Diffs[crdIdentifier] = *crdChanges
		}
	}

	return report, nil
}

func outputReport(log logrus.FieldLogger, report *report.Report, format string) {
	switch format {
	case outputFormatText:
		report.Print()
	case outputFormatJSON:
		json.NewEncoder(os.Stdout).Encode(report)
	default:
		log.Errorf("This should never happen: Do not know how to handle %s output format.", format)
	}
}
