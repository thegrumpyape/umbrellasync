package cmd

import (
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
	Short: "Sync block files with Umbrella destination lists",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Fetch the necessary parameters from the config.
		hostname, version, key, secret, filepaths := fetchConfigParameters()

		umbrellaService := umbrella.NewUmbrellaService(hostname, version, key, secret)
		destinationLists, err := umbrellaService.GetDestinationLists(100)
		if err != nil {
			log.Fatal(err)
		}

		for _, filepath := range filepaths {
			syncFile(filepath, destinationLists, umbrellaService)
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

func syncFile(filepath string, destinationLists []umbrella.DestinationList, umbrellaService umbrella.UmbrellaService) {
	blockFile, err := umbrellasync.NewBlockFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Syncing file:", blockFile.Name)

	matchingDestinationList := findMatchingDestinationList(blockFile, destinationLists)

	if matchingDestinationList == (umbrella.DestinationList{}) {
		matchingDestinationList = createDestinationList(blockFile, umbrellaService)
	}

	destinations, err := umbrellaService.GetDestinations(matchingDestinationList.ID, 100)
	if err != nil {
		log.Fatal(err)
	}

	destinationsToAdd, destinationsToRemove := umbrellasync.Compare(blockFile.Data, destinations)

	addDestinationsToUmbrella(destinationsToAdd, matchingDestinationList, umbrellaService)
	removeDestinationsFromUmbrella(destinationsToRemove, destinations, matchingDestinationList, umbrellaService)
}

func findMatchingDestinationList(blockFile umbrellasync.BlockFile, destinationLists []umbrella.DestinationList) umbrella.DestinationList {
	for _, destinationList := range destinationLists {
		if strings.Contains(destinationList.Name, blockFile.Name) {
			log.Println("Found match:", destinationList.Name)
			return destinationList
		}
	}
	return umbrella.DestinationList{}
}

func createDestinationList(blockFile umbrellasync.BlockFile, umbrellaService umbrella.UmbrellaService) umbrella.DestinationList {
	log.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
	destinationList, err := umbrellaService.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
	if err != nil {
		log.Fatal(err)
	}

	return destinationList
}

func addDestinationsToUmbrella(destinationsToAdd []string, destinationList umbrella.DestinationList, umbrellaService umbrella.UmbrellaService) {
	log.Println("Added", len(destinationsToAdd), "destinations to Umbrella:", destinationList.Name)
	var addPayload []umbrella.NewDestination
	for _, destination := range destinationsToAdd {
		addPayload = append(addPayload, umbrella.NewDestination{Destination: destination})
	}
	umbrellaService.AddDestinations(destinationList.ID, addPayload)
}

func removeDestinationsFromUmbrella(destinationsToRemove []string, existingDestinations []umbrella.Destination, destinationList umbrella.DestinationList, umbrellaService umbrella.UmbrellaService) {
	destinationMap := mapDestinationIDs(existingDestinations)

	log.Println("Removed", len(destinationsToRemove), "destinations from Umbrella:", destinationList.Name)
	var removePayload []int
	for _, destination := range destinationsToRemove {
		if id, ok := destinationMap[destination]; ok {
			removePayload = append(removePayload, id)
		}
	}
	umbrellaService.DeleteDestinations(destinationList.ID, removePayload)
}

func mapDestinationIDs(destinations []umbrella.Destination) map[string]int {
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
