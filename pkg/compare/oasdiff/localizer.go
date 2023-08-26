// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package oasdiff

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type LocalizedMessage struct {
	Key  string        `json:"key"`
	Args []interface{} `json:"args"`
}

func jsonLocalizer(key string, args ...interface{}) string {
	m := LocalizedMessage{
		Key:  key,
		Args: args,
	}

	// trim oasdiff's quoting
	for i, arg := range args {
		if sarg, ok := arg.(string); ok {
			sarg = strings.TrimSuffix(sarg, "'")
			sarg = strings.TrimPrefix(sarg, "'")

			args[i] = sarg
		}
	}

	encoded, err := json.Marshal(m)
	if err != nil {
		panic(fmt.Sprintf("failed to JSON encode: %v", err))
	}

	return string(encoded)
}

// Disect will return a concrete Message struct based on the key.
// This is also where the concept of "properties" is turned into
// "paths" inside a spec.
func (m *LocalizedMessage) Disect() interface{} {
	// This switch statement only contains keys relevant for diffs
	// that can happen with crdiff, i.e. request header diffs are
	// ignored here.
	switch m.Key {
	case "new-required-request-property":
		return &NewRequiredPropertyMessage{Path: m.Args[0].(string)}
	case "request-property-became-required":
		return &PropertyBecameRequiredMessage{Path: m.Args[0].(string)}
	case "request-property-became-enum":
		return &PropertyBecameEnumMessage{Path: m.Args[0].(string)}
	case "request-property-removed":
		return &PropertyRemovedMessage{Path: m.Args[0].(string)}
	case "request-property-type-changed":
		return &PropertyTypeChangedMessage{
			Path: m.Args[0].(string),
			From: m.Args[1].(string),
			To:   m.Args[3].(string),
		}
	case "request-property-max-length-set":
		return &PropertyMaxLengthSetMessage{
			Path:   m.Args[0].(string),
			Length: parseInt(m.Args[1]),
		}
	case "request-property-min-length-set":
		return &PropertyMinLengthSetMessage{
			Path:   m.Args[0].(string),
			Length: parseInt(m.Args[1]),
		}
	case "request-property-min-length-increased":
		return &PropertyMinLengthIncreasedMessage{
			Path: m.Args[0].(string),
			From: parseInt(m.Args[1]),
			To:   parseInt(m.Args[2]),
		}
	case "request-property-min-items-set":
		return &PropertyMinItemsSetMessage{
			Path:  m.Args[0].(string),
			Items: parseInt(m.Args[1]),
		}
	case "request-property-min-items-increased":
		return &PropertyMinItemsIncreasedMessage{
			Path: m.Args[0].(string),
			From: parseInt(m.Args[1]),
			To:   parseInt(m.Args[2]),
		}
	case "request-property-pattern-added":
		return &PropertyPatternAddedMessage{
			Pattern: m.Args[0].(string),
			Path:    m.Args[1].(string),
		}
	case "request-property-pattern-changed":
		return &PropertyPatternChangedMessage{
			Pattern: m.Args[0].(string),
			Path:    m.Args[1].(string),
		}
	default:
		return m
	}
}

func parseInt(v interface{}) int {
	if i, ok := v.(int); ok {
		return i
	}

	if s, ok := v.(string); ok {
		i, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}

		return i
	}

	panic(fmt.Sprintf("cannot parse %v (%T) as int", v, v))
}

type NewRequiredPropertyMessage struct {
	Path string `json:"path" yaml:"path"`
}

type PropertyBecameRequiredMessage struct {
	Path string `json:"path" yaml:"path"`
}

type PropertyBecameEnumMessage struct {
	Path string `json:"path" yaml:"path"`
}

type PropertyRemovedMessage struct {
	Path string `json:"path" yaml:"path"`
}

type PropertyTypeChangedMessage struct {
	Path string `json:"path" yaml:"path"`
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

type PropertyMaxLengthSetMessage struct {
	Path   string `json:"path" yaml:"path"`
	Length int    `json:"length" yaml:"length"`
}

type PropertyMinLengthSetMessage struct {
	Path   string `json:"path" yaml:"path"`
	Length int    `json:"length" yaml:"length"`
}

type PropertyMinLengthIncreasedMessage struct {
	Path string `json:"path" yaml:"path"`
	From int    `json:"from" yaml:"from"`
	To   int    `json:"to" yaml:"to"`
}

type PropertyMinItemsSetMessage struct {
	Path  string `json:"path" yaml:"path"`
	Items int    `json:"items" yaml:"items"`
}

type PropertyMinItemsIncreasedMessage struct {
	Path string `json:"path" yaml:"path"`
	From int    `json:"from" yaml:"from"`
	To   int    `json:"to" yaml:"to"`
}

type PropertyPatternAddedMessage struct {
	Path    string `json:"path" yaml:"path"`
	Pattern string `json:"pattern" yaml:"pattern"`
}

type PropertyPatternChangedMessage struct {
	Path    string `json:"path" yaml:"path"`
	Pattern string `json:"pattern" yaml:"pattern"`
}
