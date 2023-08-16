// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// These variables get set by ldflags during compilation.
var (
	BuildTag    string
	BuildCommit string
	BuildDate   string // RFC3339 format ("2006-01-02T15:04:05Z07:00")
)

type versionCmdOptions struct {
	output string
}

func (o *versionCmdOptions) PreRunE(cmd *cobra.Command, args []string) error {
	// set the log format on the global log variable
	switch o.output {
	case outputFormatText:
		// NOP
	case outputFormatJSON:
		log.SetFormatter(&logrus.JSONFormatter{})
	default:
		log.Errorf("Invalid flags: Unknown output format %q.", o.output)
		return fmt.Errorf("unknown output format %q", o.output)
	}

	return nil
}

func (o *versionCmdOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.output, "output", "o", o.output, "output format (one of [text, json])")
}

func VersionCommand(globalOpts *globalOptions) *cobra.Command {
	cmdOpts := versionCmdOptions{
		output: outputFormatText,
	}

	cmd := &cobra.Command{
		Use:          "version",
		Short:        "Print the application version and then exit",
		RunE:         VersionRunE(globalOpts, &cmdOpts),
		SilenceUsage: true,
	}

	cmdOpts.AddFlags(cmd.PersistentFlags())

	cmd.PreRunE = cmdOpts.PreRunE

	return cmd
}

type versionOutput struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	Date    string `json:"date"`
	Runtime string `json:"runtime"`
}

func VersionRunE(globalOpts *globalOptions, cmdOpts *versionCmdOptions) cobraFuncE {
	return handleErrors(func(cmd *cobra.Command, args []string) error {
		if cmdOpts.output == outputFormatJSON {
			o := versionOutput{
				Version: BuildTag,
				Commit:  BuildCommit,
				Date:    BuildDate,
				Runtime: runtime.Version(),
			}

			return json.NewEncoder(os.Stdout).Encode(o)
		}

		builtOn := BuildDate

		parsed, err := time.Parse(time.RFC3339, BuildDate)
		if err == nil {
			builtOn = parsed.Format(time.RFC1123)

			diff := time.Now().Sub(parsed)
			if hours := diff.Hours(); hours > 24 {
				builtOn += fmt.Sprintf(" (%d days ago)", int(hours/24))
			}
		}

		fmt.Printf(
			"CRDiff %s (%s), built with %s on %s\n",
			BuildTag,
			BuildCommit[:10],
			runtime.Version(),
			builtOn,
		)

		return nil
	})
}
