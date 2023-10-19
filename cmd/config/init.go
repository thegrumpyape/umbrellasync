package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewInitCommand(deps *ConfigCommandDependencies) *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialze the config file",
		Long:  "Initialzes the umbrellasync config.yaml file. Defaults to $HOME/.umbrellasync/config.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			err := initialize(deps)
			if err != nil {
				fmt.Printf("Failed initializing config. Error: %s", err)
			}
		},
	}

	return initCommand
}

func initialize(deps *ConfigCommandDependencies) error {
	return deps.ConfigurationManager.Prompt()
}
