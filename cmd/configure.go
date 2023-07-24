package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Generate config file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("configure called")
		err := CreateConfig()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func prompt(p string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(p)
	res, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	res = strings.TrimSuffix(res, "\n")
	res = strings.TrimSuffix(res, "\r")

	return res, nil
}

// Create config file
func CreateConfig() error {
	fmt.Println("Creating a new config file...")
	fmt.Println("Getting Umbrella config")

	// Get API Hostname
	hostname, err := prompt("API Hostname:")
	if err != nil {
		return err
	}
	viper.Set("api_hostname", hostname)

	// Get API Version
	version, err := prompt("API Version:")
	if err != nil {
		return err
	}

	if strings.Contains(version, "v") {
		viper.Set("api_version", version)
	} else {
		viper.Set("api_version", fmt.Sprint("v"+version))
	}

	// Get Client ID
	key, err := prompt("Client ID:")
	if err != nil {
		return err
	}
	viper.Set("key", key)

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
		filepath, err := prompt("File Path:")
		if err != nil {
			return err
		}

		// Check if file exists
		if _, err := os.Stat(filepath); err == nil {
			files = append(files, filepath)

		} else if errors.Is(err, os.ErrNotExist) {
			fmt.Print("File not found:", filepath)
			fmt.Println("Please verify path is correct.")
			continue

		} else {
			fmt.Println("Something weird happened. Quitting.")
			return nil
		}

		// Ask if another file needs to be added
		res, err := prompt("Add another file? (Y/n):")
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
	viper.WriteConfigAs(cfgFile)
	err = os.Chmod(cfgFile, 0600)
	if err != nil {
		return err
	}

	fmt.Println("\nWrote config to", cfgFile)
	return nil
}

func init() {
	rootCmd.AddCommand(configureCmd)
}
