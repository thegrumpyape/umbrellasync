package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thegrumpyape/umbrellasync/pkg/blockfile"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
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

		// create umbrella service
		umbrellaService, err := umbrella.CreateClient(hostname, version, key, secret, logger)
		if err != nil {
			log.Fatal(err)
		}

		// get destination lists
		destinationLists, err := umbrellaService.GetDestinationLists(100)
		if err != nil {
			log.Fatal(err)
		}

		for _, filepath := range filepaths {
			// get block file data
			blockFile, err := blockfile.New(filepath)
			if err != nil {
				log.Fatal(err)
			}

			// find matching destination list for block file
			matchingDestinationList := umbrella.DestinationList{}

			for _, dl := range destinationLists {
				if strings.Contains(dl.Name, blockFile.Name) {
					matchingDestinationList = dl
					break
				}
			}

			// if no match is found, create a new destination list
			if matchingDestinationList == (umbrella.DestinationList{}) {
				log.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
				matchingDestinationList, err = umbrellaService.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
				if err != nil {
					log.Fatal(err)
				}
			}

			// sync file with matching destination list
			blockFile.Sync(umbrellaService, matchingDestinationList)
			fmt.Println("Waiting for 60 seconds...")
			time.Sleep(60 * time.Second)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
