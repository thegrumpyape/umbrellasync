package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	umbrellasync "github.com/thegrumpyape/umbrellasync/pkg"
	"github.com/thegrumpyape/umbrellasync/pkg/api"
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
			syncFile(filepath, destinationLists, umbrellaService)
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

func syncFile(filepath string, destinationLists []api.DestinationList, umbrellaService api.UmbrellaService) {
	blockFile, err := umbrellasync.NewBlockFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Syncing file:", blockFile.Name)
	log.Println("Syncing file:", blockFile.Name)

	matchingDestinationList := findMatchingDestinationList(blockFile, destinationLists)

	if matchingDestinationList == (api.DestinationList{}) {
		matchingDestinationList = createDestinationList(blockFile, umbrellaService)
	}

	destinations, err := umbrellaService.GetDestinations(matchingDestinationList.ID, 100)
	if err != nil {
		log.Fatal(err)
	}

	destinationsToAdd, destinationsToRemove := umbrellasync.Compare(blockFile.Data, destinations)

	if len(destinationsToAdd) != 0 {
		addDestinationsToUmbrella(destinationsToAdd, matchingDestinationList, umbrellaService)
	}

	if len(destinationsToRemove) != 0 {
		removeDestinationsFromUmbrella(destinationsToRemove, destinations, matchingDestinationList, umbrellaService)
	}
}

func findMatchingDestinationList(blockFile umbrellasync.BlockFile, destinationLists []api.DestinationList) api.DestinationList {
	for _, destinationList := range destinationLists {
		if strings.Contains(destinationList.Name, blockFile.Name) {
			fmt.Println("Found match:", destinationList.Name)
			log.Println("Found match:", destinationList.Name)
			return destinationList
		}
	}
	return api.DestinationList{}
}

func createDestinationList(blockFile umbrellasync.BlockFile, umbrellaService api.UmbrellaService) api.DestinationList {
	fmt.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
	log.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
	destinationList, err := umbrellaService.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
	if err != nil {
		log.Fatal(err)
	}

	return destinationList
}

func addDestinationsToUmbrella(destinationsToAdd []string, destinationList api.DestinationList, umbrellaService api.UmbrellaService) {
	fmt.Println("Added", len(destinationsToAdd), "destinations to Umbrella:", destinationList.Name)

	chunkSize := 500
	for i := 0; i < len(destinationsToAdd); i += chunkSize {
		end := i + chunkSize

		// Avoid going over the slice bounds
		if end > len(destinationsToAdd) {
			end = len(destinationsToAdd)
		}

		var addPayload []api.NewDestination
		for _, destination := range destinationsToAdd[i:end] {
			addPayload = append(addPayload, api.NewDestination{Destination: destination})
		}

		fmt.Println(addPayload)
		umbrellaService.AddDestinations(destinationList.ID, addPayload)
	}
}

func removeDestinationsFromUmbrella(destinationsToRemove []string, existingDestinations []api.Destination, destinationList api.DestinationList, umbrellaService api.UmbrellaService) {
	destinationMap := mapDestinationIDs(existingDestinations)

	fmt.Println("Removed", len(destinationsToRemove), "destinations from Umbrella:", destinationList.Name)

	chunkSize := 500
	for i := 0; i < len(destinationsToRemove); i += chunkSize {
		end := i + chunkSize

		// Avoid going over the slice bounds
		if end > len(destinationsToRemove) {
			end = len(destinationsToRemove)
		}

		var removePayload []int
		for _, destination := range destinationsToRemove[i:end] {
			if id, ok := destinationMap[destination]; ok {
				removePayload = append(removePayload, id)
			}
		}

		umbrellaService.DeleteDestinations(destinationList.ID, removePayload)
	}
}

func mapDestinationIDs(destinations []api.Destination) map[string]int {
	destinationMap := make(map[string]int)
	for _, destination := range destinations {
		id, _ := strconv.Atoi(destination.ID)
		destinationMap[destination.Destination] = id
	}
	return destinationMap
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
