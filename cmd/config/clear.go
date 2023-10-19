package config

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewClearCommand(deps *ConfigCommandDependencies) *cobra.Command {
	clearCommand := &cobra.Command{
		Use:   "clear",
		Short: "Clear configuration value",
		Long:  "Remove value for specific key in umbrellasync config.yaml file. Defaults to $HOME/.umbrellasync/config.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			err := clear(deps, args[0])
			if err != nil {
				fmt.Printf("Failed clearing %s. Error: %s", args[0], err)
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly 1 argument")
			}

			return validateKey(args[0])
		},
	}

	return clearCommand
}

func clear(deps *ConfigCommandDependencies, key string) error {
	return deps.ConfigurationManager.Clear(key)
}
