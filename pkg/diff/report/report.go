// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package report

import (
	"go.xrstf.de/crdiff/pkg/diff"
)

type Report struct {
	Diffs map[string]diff.CRDDiff `json:"diffs"`
}

func (r *Report) Empty() bool {
	if r == nil {
		return true
	}

	for _, change := range r.Diffs {
		if !change.Empty() {
			return false
		}
	}

	return true
}
