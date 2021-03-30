package cmd

import (
	"flag"
	"log"
	"os"

	"github.com/vmkteam/mfd-generator/api"
	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/generators/vt"
	"github.com/vmkteam/mfd-generator/generators/vt-template"
	"github.com/vmkteam/mfd-generator/generators/xml"
	"github.com/vmkteam/mfd-generator/generators/xml-lang"
	"github.com/vmkteam/mfd-generator/generators/xml-vt"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "mfd",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		if flag := cmd.Flag("export"); flag != nil && flag.Value.String() != "" {
			if err := exportTemplates(flag.Value.String()); err != nil {
				log.Printf("export error: %s", err)
				os.Exit(0)
				return
			}

			log.Printf("export complete to path %s", flag.Value.String())
			os.Exit(0)
		}

		if !cmd.HasSubCommands() {
			if err := cmd.Help(); err != nil {
				panic("help not found")
			}
			os.Exit(0)
		}
	},
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
}

var debug = flag.Bool("debug", false, "enable debug output")

func init() {
	flags := root.Flags()
	flags.StringP("export", "e", "", "path to export templates, rest commands wil be ignored")

	root.AddCommand(
		xml.CreateCommand(),
		xmlvt.CreateCommand(),
		xmllang.CreateCommand(),
		model.CreateCommand(),
		repo.CreateCommand(),
		vt.CreateCommand(),
		vttmpl.CreateCommand(),
		api.CreateCommand(),
	)
}

// Execute runs root cmd
func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
