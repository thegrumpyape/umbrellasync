package config

import (
	"fmt"
	"reflect"

	"github.com/spf13/cobra"
	"github.com/thegrumpyape/umbrellasync/pkg/configurationManager"
)

var ConfigAvailableKeys = []string{"apihostname", "apiversion", "key", "secret", "files"}

type ConfigCommandDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
}

func New(deps *ConfigCommandDependencies) *cobra.Command {
	configCommand := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "Internal configuration management for umbrellasync config file",
	}

	configCommand.AddCommand(NewGetCommand(deps))
	configCommand.AddCommand(NewSetCommand(deps))
	configCommand.AddCommand(NewClearCommand(deps))
	configCommand.AddCommand(NewInitCommand(deps))

	return configCommand
}

func validateKey(key string) error {
	validKeys := make(map[string]bool)

	for _, key := range ConfigAvailableKeys {
		validKeys[key] = true
	}

	if val, ok := validKeys[key]; !ok || !val {
		return fmt.Errorf("key must be one of: %s", reflect.ValueOf(validKeys).MapKeys())
	}
	return nil
}
