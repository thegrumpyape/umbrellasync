package sync

import (
	"log"
	"strings"

	"github.com/thegrumpyape/umbrellasync/pkg/io"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

func SyncFile(filepath string, destinationLists []umbrella.DestinationList, umbrellaService umbrella.UmbrellaService) error {
	chunkSize := 500
	blockFile, err := io.NewBlockFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Syncing file:", blockFile.Name)

	matchingDestinationList := umbrella.DestinationList{}
	for _, destinationList := range destinationLists {
		if strings.Contains(destinationList.Name, blockFile.Name) {
			log.Println("Found match:", destinationList.Name)
			matchingDestinationList = destinationList
		}
	}

	if matchingDestinationList == (umbrella.DestinationList{}) {
		log.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
		matchingDestinationList, err = umbrellaService.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
		if err != nil {
			return err
		}
	}

	destinations, err := umbrellaService.GetDestinations(matchingDestinationList.ID, 100)
	if err != nil {
		log.Fatal(err)
	}

	destinationsToAdd, destinationsToRemove := compare(blockFile.Data, destinations)

	if len(destinationsToAdd) != 0 {
		umbrellaService.AddDestinations(matchingDestinationList, destinationsToAdd, chunkSize)
	}

	if len(destinationsToRemove) != 0 {
		umbrellaService.DeleteDestinations(matchingDestinationList, destinationsToRemove, destinations, chunkSize)
	}

	return nil
}

// Compares BlockFile with Destinations from DestinationList
func compare(bl []string, dl []umbrella.Destination) ([]string, []string) {
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
