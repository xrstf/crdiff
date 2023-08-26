// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package oasdiff

import (
	"encoding/json"
	"fmt"
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
