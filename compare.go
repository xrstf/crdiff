// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/crdiff/pkg/crd"
	"go.xrstf.de/crdiff/pkg/diff"
)

func compareCRDs(log logrus.FieldLogger, baseCRDs, revisionCRDs map[string]crd.CRD, diffOpt diff.Options) (*DiffReport, error) {
	report := &DiffReport{
		Diffs: map[string]diff.CRDDiff{},
	}

	for crdIdentifier, baseCRD := range baseCRDs {
		revisionCRD, exists := revisionCRDs[crdIdentifier]
		if !exists {
			report.Diffs[crdIdentifier] = diff.CRDDiff{
				General: []diff.Change{{
					Modification: "CRD has been removed",
				}},
			}

			continue
		}

		crdChanges, err := diff.CompareCRDs(baseCRD, revisionCRD, diffOpt)
		if err != nil {
			return nil, fmt.Errorf("failed to compare %q: %w", crdIdentifier, err)
		}

		if crdChanges != nil {
			report.Diffs[crdIdentifier] = *crdChanges
		}
	}

	return report, nil
}
