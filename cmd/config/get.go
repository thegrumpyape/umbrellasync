package config

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewGetCommand(deps *ConfigCommandDependencies) *cobra.Command {
	getCommand := &cobra.Command{
		Use:   "get",
		Short: "Get configuration value",
		Long:  "Get value for specific key in umbrellasync config.yaml file. Defaults to $HOME/.umbrellasync/config.yaml",
		Run: func(cmd *cobra.Command, args []string) {
			value := get(deps, args[0])
			fmt.Println(value)
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("requires exactly 1 arguments")
			}

			return validateKey(args[0])
		},
	}

	return getCommand
}

func get(deps *ConfigCommandDependencies, key string) interface{} {
	return deps.ConfigurationManager.Get(key)
}
