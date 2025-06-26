package cmd

import (
	"fmt"
	"os"

	"github.com/vmkteam/mfd-generator/api"
	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/generators/vt"
	vttmpl "github.com/vmkteam/mfd-generator/generators/vt-template"
	"github.com/vmkteam/mfd-generator/generators/xml"
	xmllang "github.com/vmkteam/mfd-generator/generators/xml-lang"
	xmlvt "github.com/vmkteam/mfd-generator/generators/xml-vt"
	"github.com/vmkteam/mfd-generator/mfd"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:   "mfd-generator",
	Short: "",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
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

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		//nolint:forbidigo
		fmt.Println("MFD Generator", mfd.Version)
	},
}

func init() {
	root.AddCommand(
		xml.CreateCommand(),
		xmlvt.CreateCommand(),
		xmllang.CreateCommand(),
		model.CreateCommand(),
		repo.CreateCommand(),
		vt.CreateCommand(),
		vttmpl.CreateCommand(),
		api.CreateCommand(),
		versionCmd,
	)
}

// Execute runs root cmd
func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
