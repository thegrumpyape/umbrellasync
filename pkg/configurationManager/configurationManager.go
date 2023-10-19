package configurationManager

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/viper"
	"github.com/thegrumpyape/umbrellasync/pkg/fileManager"
	"github.com/thegrumpyape/umbrellasync/pkg/utils"
	"golang.org/x/term"
)

var configPath string

type ConfigurationManager struct {
	TokenGenerationUrl string
}

func New() *ConfigurationManager {
	return &ConfigurationManager{
		TokenGenerationUrl: "https://api.umbrella.com/auth/v2/token",
	}
}

func (cm *ConfigurationManager) InitConfigFile() {
	configHome, configName, configType, err := setViperConfig()

	if err != nil {
		log.Fatal(err)
	}

	configPath = filepath.Join(configHome, configName+"."+configType)

	isDirExists, err := fileManager.IsExists(configHome)
	if err != nil {
		log.Fatal(err)
	}
	if !isDirExists {
		osMkdirErr := os.Mkdir(configHome, os.ModePerm)
		if osMkdirErr != nil {
			log.Fatal(osMkdirErr)
		}
	}

	isConfigExists, err := fileManager.IsExists(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if !isConfigExists {
		_, osCreateErr := os.Create(configPath)
		if osCreateErr != nil {
			log.Fatal(osCreateErr)
		}
	}

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			viper.SetConfigFile(configPath)
		} else {
			log.Fatal(err)
		}
	}
}

func setViperConfig() (string, string, string, error) {
	configHome := fmt.Sprintf("%s\\.umbrellasync\\", fileManager.GetHomedir())
	configName := "config"
	configType := "yaml"

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(configHome)

	return configHome, configName, configType, nil
}

func (cm *ConfigurationManager) Prompt() error {
	err := promptConfigValue("API Hostname:", "apihostname")
	if err != nil {
		return err
	}

	err = promptConfigValue("API Version:", "apiversion")
	if err != nil {
		return err
	}

	// Appends "v" if not provided in prompt
	if !strings.Contains(viper.Get("apiversion").(string), "v") {
		number := viper.Get("apiversion").(string)
		viper.Set("api_version", fmt.Sprint("v"+number))
	}

	// Get Client ID
	err = promptConfigValue("Client ID:", "key")
	if err != nil {
		return err
	}

	// Get Client Secret
	fmt.Print("Client Secret: ")
	s, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}
	secret := string(s)
	secret = strings.TrimSuffix(secret, "\n")
	// Windows
	secret = strings.TrimSuffix(secret, "\r")
	viper.Set("secret", secret)

	// Get Files
	var files []string
	fmt.Println("\n\nWhat files do you want to sync?")
	for {
		filepath, err := utils.GetUserInput("File Path:")
		if err != nil {
			return err
		}

		// Check if file exists
		if _, err := os.Stat(filepath); err == nil {
			files = append(files, filepath)

		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Println("File not found:", filepath)
			fmt.Println("Please verify path is correct.")
			continue

		} else {
			fmt.Println("Something weird happened. Quitting.")
			return nil
		}

		// Ask if another file needs to be added
		res, err := utils.GetUserInput("Add another file? (Y/n):")
		if err != nil {
			return err
		}
		res = strings.ToLower(res)
		if res == "n" {
			break
		}
	}

	viper.Set("files", files)

	// Write Viper Config
	viper.WriteConfigAs(configPath)
	err = os.Chmod(configPath, 0600)
	if err != nil {
		return err
	}

	fmt.Println("\nWrote config to", configPath)
	return nil
}

func (cm *ConfigurationManager) Set(key string, value string) error {
	viper.Set(key, value)
	writeClientIdErr := viper.WriteConfigAs(configPath)
	if writeClientIdErr != nil {
		return writeClientIdErr
	}
	return nil
}

func (cm *ConfigurationManager) Add(key string, value string) error {
	viper.Set(key, value)
	writeClientIdErr := viper.WriteConfigAs(configPath)
	if writeClientIdErr != nil {
		return writeClientIdErr
	}
	return nil
}

func (cm *ConfigurationManager) Append(key string, value string) error {
	var slice []interface{}
	initial := viper.Get(key)

	contains := func(s []interface{}, e string) bool {
		for _, a := range s {
			if a == e {
				return true
			}
		}
		return false
	}

	switch t := initial.(type) {
	case nil:
		slice = []interface{}{value}
	case []interface{}:
		if !contains(t, value) {
			slice = append(t, value)
		} else {
			slice = t
		}
	case interface{}:
		if t != value {
			slice = []interface{}{t, value}
		} else {
			slice = []interface{}{t}
		}
	default:
		return fmt.Errorf("unsupported type: %T", t)
	}

	viper.Set(key, slice)
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		return err
	}

	return nil
}

func (cm *ConfigurationManager) Get(key string) any {
	return viper.Get(key)
}

func (cm *ConfigurationManager) Clear(key string) error {
	fullConfig := viper.AllSettings()
	delete(fullConfig, key)
	viper.Reset()
	setViperConfig()
	for k, v := range fullConfig {
		viper.Set(k, v)
	}

	writeClientIdErr := viper.WriteConfigAs(configPath)
	if writeClientIdErr != nil {
		return writeClientIdErr
	}
	return nil
}

func promptConfigValue(prompt string, key string) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	res, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	res = strings.TrimSuffix(res, "\n")
	res = strings.TrimSuffix(res, "\r")

	viper.Set(key, res)
	return nil
}
