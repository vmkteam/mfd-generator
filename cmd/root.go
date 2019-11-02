package cmd

import (
	"flag"
	"os"

	"github.com/vmkteam/mfd-generator/generators/model"
	"github.com/vmkteam/mfd-generator/generators/repo"
	"github.com/vmkteam/mfd-generator/generators/vt"
	"github.com/vmkteam/mfd-generator/generators/vt-template"
	"github.com/vmkteam/mfd-generator/generators/xml"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var root = &cobra.Command{
	Use:   "mfd",
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

var debug = flag.Bool("debug", false, "enable debug output")

func init() {
	var config zap.Config

	// using stdlib here, because cobra is not executed yet
	flag.Parse()
	if debug != nil && *debug {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.OutputPaths = []string{"stdout"}
	config.Encoding = "console"
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	root.AddCommand(
		xml.CreateCommand(logger),
		model.CreateCommand(logger),
		repo.CreateCommand(logger),
		vt.CreateCommand(logger),
		vttmpl.CreateCommand(logger),
	)
}

// Execute runs root cmd
func Execute() {
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
