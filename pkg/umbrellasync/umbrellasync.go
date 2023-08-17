package umbrellasync

import (
	"fmt"
	"log"
	"strings"

	"github.com/thegrumpyape/umbrellasync/pkg/api"
	"github.com/thegrumpyape/umbrellasync/pkg/file"
)

func SyncFile(filepath string, destinationLists []api.DestinationList, umbrellaService api.UmbrellaService) error {
	blockFile, err := file.NewBlockFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Syncing file:", blockFile.Name)
	log.Println("Syncing file:", blockFile.Name)

	matchingDestinationList := FindMatchingDestinationList(blockFile.Name, destinationLists)

	if matchingDestinationList == (api.DestinationList{}) {
		matchingDestinationList, err = api.CreateDestinationList(blockFile.Name, umbrellaService)
		if err != nil {
			return err
		}
	}

	destinations, err := umbrellaService.GetDestinations(matchingDestinationList.ID, 100)
	if err != nil {
		log.Fatal(err)
	}

	destinationsToAdd, destinationsToRemove := Compare(blockFile.Data, destinations)

	if len(destinationsToAdd) != 0 {
		api.AddDestinationsToUmbrella(destinationsToAdd, matchingDestinationList, umbrellaService)
	}

	if len(destinationsToRemove) != 0 {
		api.RemoveDestinationsFromUmbrella(destinationsToRemove, destinations, matchingDestinationList, umbrellaService)
	}

	return nil
}

// Compares BlockFile with Destinations from DestinationList
func Compare(bl []string, dl []api.Destination) ([]string, []string) {
	var destsToAdd, destsToDelete []string

	mA, mB := make(map[string]bool), make(map[string]bool)

	for _, item := range bl {
		mA[item] = true
	}

	for _, item := range dl {
		mB[item.Destination] = true
	}

	for key := range mA {
		if !mB[key] {
			destsToAdd = append(destsToAdd, key)
		}
	}

	for key := range mB {
		if !mA[key] {
			destsToDelete = append(destsToDelete, key)
		}
	}

	return destsToAdd, destsToDelete
}

func FindMatchingDestinationList(blockFileName string, destinationLists []api.DestinationList) api.DestinationList {
	for _, destinationList := range destinationLists {
		if strings.Contains(destinationList.Name, blockFileName) {
			fmt.Println("Found match:", destinationList.Name)
			log.Println("Found match:", destinationList.Name)
			return destinationList
		}
	}
	return api.DestinationList{}
}
