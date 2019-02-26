package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"github.com/ubclaunchpad/inertia/cfg"
	"github.com/ubclaunchpad/inertia/cmd"

	"github.com/ubclaunchpad/inertia/cmd/core"
	"github.com/ubclaunchpad/inertia/cmd/core/utils/output"
	remotescmd "github.com/ubclaunchpad/inertia/cmd/remotes"
)

// Version denotes the version of the binary
var (
	Version string

	mdReadmeTemplate = `# Inertia Command Reference

Click [here](/inertia.md) for the Inertia CLI command reference. It is generated
automatically using ` + "`inertia-docgen`." + `

For a more general usage guide, refer to the [Inertia Usage Guide](https://inertia.ubclaunchpad.com).

For documentation regarding the daemon API, refer to the [API Reference](https://inertia.ubclaunchpad.com/api).

* Generated: %s
* Version: %s
`
)

func main() {
	var root = cmd.NewInertiaCmd(Version, "~/.inertia/inertia.global")
	if err := newDocgenCmd(root).Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func newDocgenCmd(root *core.Cmd) *cobra.Command {
	const (
		flagOutput = "output"
		flagFormat = "format"
	)
	var docs = &cobra.Command{
		Use:     "inertia-docgen",
		Hidden:  true,
		Version: Version,
		Short:   "Generate command reference for the Inertia CLI.",
		Run: func(cmd *cobra.Command, args []string) {
			var out, _ = cmd.Flags().GetString(flagOutput)
			var format, _ = cmd.Flags().GetString(flagFormat)

			// create *full* Inertia tree, for sake of documentation
			remotescmd.AttachRemoteHostCmd(root,
				remotescmd.CmdOptions{
					RemoteCfg: &cfg.Remote{Name: "${remote_name}"},
				},
				false)

			// set up file tree
			os.MkdirAll(out, os.ModePerm)

			// gen docs
			switch format {
			case "man":
				if err := doc.GenManTree(root.Command, &doc.GenManHeader{
					Title: "Inertia CLI Command Reference",
					Source: fmt.Sprintf(
						"Generated by inertia-docgen %s",
						root.Version),
					Manual: "https://inertia.ubclaunchpad.com",
				}, out); err != nil {
					output.Fatal(err.Error())
				}
			default:
				if err := doc.GenMarkdownTree(root.Command, out); err != nil {
					output.Fatal(err.Error())
				}
				var readme = fmt.Sprintf(mdReadmeTemplate, time.Now().Format("2006-Jan-02"), Version)
				ioutil.WriteFile(filepath.Join(out, "README.md"), []byte(readme), os.ModePerm)
			}

			fmt.Printf("%s documentation generated in %s\n", format, out)
		},
	}
	docs.Flags().StringP(flagOutput, "o", "./docs/cli", "output file path")
	docs.Flags().StringP(flagFormat, "f", "md", "format to generate (md|man)")
	return docs
}
