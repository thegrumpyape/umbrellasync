package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thegrumpyape/umbrellasync/pkg/api"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrellasync"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync block files with Umbrella destination lists",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch the necessary parameters from the config.
		hostname, version, key, secret, filepaths := fetchConfigParameters()

		umbrellaService, err := api.NewUmbrellaService(hostname, version, key, secret, logger)
		if err != nil {
			log.Fatal(err)
		}
		destinationLists, err := umbrellaService.GetDestinationLists(100)
		if err != nil {
			log.Fatal(err)
		}

		for _, filepath := range filepaths {
			umbrellasync.SyncFile(filepath, destinationLists, umbrellaService)
			fmt.Println("Waiting for 60 seconds...")
			time.Sleep(60 * time.Second)
		}
	},
}

func fetchConfigParameters() (string, string, string, string, []string) {
	hostname := viper.GetString("api_hostname")
	version := viper.GetString("api_version")
	key := viper.GetString("key")
	secret := viper.GetString("secret")
	filepaths := viper.GetStringSlice("files")

	return hostname, version, key, secret, filepaths
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
