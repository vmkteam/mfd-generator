package api

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func CreateCommand() *cobra.Command {
	server := NewServer()

	command := &cobra.Command{
		Use:   "server",
		Short: "Run web server with generators",
		Long:  "",
		Run: func(command *cobra.Command, args []string) {
			if !command.HasFlags() {
				if err := command.Help(); err != nil {
					log.Printf("help not found, error: %s", err)
				}
				os.Exit(0)
				return
			}

			if err := server.ReadFlags(command); err != nil {
				log.Printf("read flags error: %s", err)
				return
			}

			if err := server.Serve(); err != nil {
				log.Printf("serve error: %s", err)
				return
			}
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}

	server.AddFlags(command)

	return command
}
