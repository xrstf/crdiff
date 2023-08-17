// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package indent

import (
	"fmt"
	"strings"
)

type Indenter struct {
	lines []string
	depth int
}

func NewIndenter() *Indenter {
	return &Indenter{
		depth: 0,
		lines: []string{},
	}
}

func (i *Indenter) Indent() *Indenter {
	i.depth++
	return i
}

func (i *Indenter) Dedent() *Indenter {
	i.depth--
	if i.depth < 0 {
		i.depth = 0
	}
	return i
}

func (i *Indenter) padding() string {
	return strings.Repeat("  ", i.depth)
}

func (i *Indenter) Add(chunk *Indenter) *Indenter {
	for _, line := range chunk.lines {
		i.lines = append(i.lines, i.padding()+line)
	}

	return i
}

func (i *Indenter) AddLine(s string) *Indenter {
	lines := strings.Split(s, "\n")

	for _, line := range lines {
		i.lines = append(i.lines, i.padding()+line)
	}

	return i
}

func (i *Indenter) PrependLine(s string) *Indenter {
	lines := strings.Split(s, "\n")
	i.lines = append(lines, i.lines...)

	return i
}

func (i *Indenter) AddLinef(pattern string, args ...interface{}) *Indenter {
	return i.AddLine(fmt.Sprintf(pattern, args...))
}

func (i *Indenter) PrependLinef(pattern string, args ...interface{}) *Indenter {
	return i.PrependLine(fmt.Sprintf(pattern, args...))
}

func (i *Indenter) String() string {
	return strings.Join(i.lines, "\n")
}

func (i *Indenter) Empty() bool {
	for _, line := range i.lines {
		if strings.TrimSpace(line) != "" {
			return false
		}
	}

	return true
}
