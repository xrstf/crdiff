// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package report

import (
	"go.xrstf.de/crdiff/pkg/compare"
)

// Report is the result of one execution of crdiff.
// It contains the diffs for every found CRDs.
type Report struct {
	Diffs map[string]compare.CRDDiff `json:"diffs"`
}

func (r *Report) HasChanges() bool {
	if r == nil {
		return false
	}

	for _, change := range r.Diffs {
		if change.HasChanges() {
			return true
		}
	}

	return false
}
