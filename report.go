// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import "go.xrstf.de/crdiff/pkg/diff"

type DiffReport struct {
	Changes map[string]diff.Diff
}

func (r *DiffReport) Empty() bool {
	if r == nil {
		return true
	}

	for _, change := range r.Changes {
		if !change.Empty() {
			return false
		}
	}

	return true
}
