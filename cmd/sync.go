package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thegrumpyape/umbrellasync/pkg/blockfile"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrellasync"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync block files with Umbrella destination lists",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch the necessary parameters from the config.
		hostname := viper.GetString("api_hostname")
		version := viper.GetString("api_version")
		key := viper.GetString("key")
		secret := viper.GetString("secret")
		filepaths := viper.GetStringSlice("files")

		umbrellaSync, err := umbrellasync.New(hostname, version, key, secret, logger)
		if err != nil {
			logger.Fatal(err)
		}

		for _, filepath := range filepaths {
			// Create blockfile
			blockFile, err := blockfile.New(filepath)
			if err != nil {
				logger.Fatal(err)
			}

			// Sync blockfile
			logger.Println("Syncing file:", blockFile.Name)
			umbrellaSync.Sync(blockFile)
			logger.Println("Waiting for 60 seconds...")
			time.Sleep(60 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
