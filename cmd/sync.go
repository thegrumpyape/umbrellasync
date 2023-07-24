package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	umbrellasync "github.com/thegrumpyape/umbrellasync/pkg"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sync called")
		hostname := viper.GetString("api_hostname")
		version := viper.GetString("api_version")
		key := viper.GetString("key")
		secret := viper.GetString("secret")
		filepaths := viper.GetStringSlice("files")

		usvc := umbrella.NewUmbrellaSvc(hostname, version, key, secret)
		dls, err := usvc.GetDestinationLists(100)
		if err != nil {
			log.Fatal(err)
		}

		for _, fp := range filepaths {

			blockFile, err := umbrellasync.NewBlockFile(fp)

			log.Println("Syncing file:", blockFile.Name)

			var dlmatch umbrella.DestinationList

			// Find a matching Umbrella destination list for blocklist file
			for _, dl := range dls {
				if strings.Contains(dl.Name, blockFile.Name) {
					log.Println("Found match:", dl.Name)
					dlmatch = dl
				}
			}

			if dlmatch == (umbrella.DestinationList{}) {
				log.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
				dlmatch, err = usvc.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
				if err != nil {
					log.Fatal(err)
				}
			}

			// Retrieve
			destinations, err := usvc.GetDestinations(dlmatch.ID, 100)
			if err != nil {
				log.Fatal(err)
			}

			destsToAdd, destsToRemove := umbrellasync.Compare(blockFile.Data, destinations)

			log.Println("Added", len(destsToAdd), "destinations to Umbrella:", dlmatch.Name)
			var addPayload []umbrella.NewDestination
			for _, dest := range destsToAdd {
				addPayload = append(addPayload, umbrella.NewDestination{Destination: dest})
			}
			usvc.AddDestinations(dlmatch.ID, addPayload)

			log.Println("Added", len(destsToAdd), "destination from Umbrella:", dlmatch.Name)
			var rmPayload []int
			for _, dest := range destsToRemove {
				for _, uDest := range destinations {
					if dest == uDest.Destination {
						intID, err := strconv.Atoi(uDest.ID)
						if err != nil {
							log.Fatal(err)
						}
						rmPayload = append(rmPayload, intID)
					}
				}
			}
			usvc.DeleteDestinations(dlmatch.ID, rmPayload)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
