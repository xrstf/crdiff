// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package oasdiff

import (
	"encoding/json"
	"io"

	"github.com/tufin/oasdiff/diff"
)

func DeepCopySchemaDiff(in *diff.SchemaDiff) *diff.SchemaDiff {
	if in == nil {
		return nil
	}

	reader, writer := io.Pipe()

	encoder := json.NewEncoder(writer)
	decoder := json.NewDecoder(reader)

	go func() {
		if err := encoder.Encode(in); err != nil {
			panic(err)
		}
	}()

	out := &diff.SchemaDiff{}
	if err := decoder.Decode(out); err != nil {
		panic(err)
	}

	return out
}
