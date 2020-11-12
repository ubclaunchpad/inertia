/*

Inertia-publish is a tool for helping publish Inertia releases to various Launch Pad
distribution channels.

For example, to publish to the Launch Pad Hombrew Tap (https://github.com/ubclaunchpad/homebrew-tap):

	inertia contrib publish homebrew

Learn more about `inertia/contrib` tools:

	inertia contrib -h

*/
package main

import (
	"bytes"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/out"
	templatedata "github.com/ubclaunchpad/inertia/contrib/inertia-publish/templates"
)

// channelConfigs defines the supported release channels at Launch Pad.
//
// See the `channelConfig` struct documentation for more details.
var channelConfigs = map[string]channelConfig{
	"homebrew": {
		Repository: "ubclaunchpad/homebrew-tap",
		Template:   "homebrew.rb",
		PublishAs:  "inertia.rb",
	},
	"scoop": {
		Repository: "ubclaunchpad/scoop-bucket",
		Template:   "scoop.json",
		PublishAs:  "inertia.json",
	},
}

func main() {
	var (
		flagOutput = "output"
		flagDryRun = "dryrun"
	)
	var publish = &cobra.Command{
		Use:    "inertia-publish [version] [channels]",
		Hidden: true,
		Args:   cobra.MinimumNArgs(1),
		Short:  "Publish Inertia releases to various distribution channels.",
		Run: func(cmd *cobra.Command, args []string) {
			version := args[0]
			outdir, _ := cmd.Flags().GetString(flagOutput)
			dryRun, _ := cmd.Flags().GetBool(flagDryRun)

			os.MkdirAll(outdir, os.ModePerm)
			data, err := newTemplateVariables(version, outdir)
			if err != nil {
				out.Fatalf("failed to load release data: %s\n", err)
			}

			for _, channel := range args[1:] {
				conf, exists := channelConfigs[channel]
				if !exists {
					out.Fatalf("unknown channel %q", channel)
				}
				out.Printf("Preparing %q publish...\n", channel)

				rawTemplate, err := templatedata.ReadFile(filepath.Join("templates", conf.Template))
				if err != nil {
					out.Fatal(err)
				}
				tmpl, err := template.New(channel).Parse(string(rawTemplate))
				if err != nil {
					out.Fatal(err)
				}
				var rendered bytes.Buffer
				if err := tmpl.Execute(&rendered, data); err != nil {
					out.Fatal(err)
				}
				if err := conf.Publish(data.Version, outdir, rendered.Bytes(), dryRun); err != nil {
					out.Fatal(err)
				}
			}
		},
	}
	publish.Flags().String(flagOutput, "dist", "output directory for rendered templates and binaries")
	publish.Flags().Bool(flagDryRun, true, "disable publish step")
	if err := publish.Execute(); err != nil {
		out.Fatal(err)
	}
}
