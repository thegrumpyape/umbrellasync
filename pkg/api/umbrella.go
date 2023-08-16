package api

import (
	"fmt"
	"log"
	"strings"

	"github.com/thegrumpyape/umbrellasync/pkg/utils"
)

func CreateDestinationList(filename string, umbrellaService UmbrellaService) (DestinationList, error) {
	fmt.Println("Creating new blocklist in Umbrella: SOC Block", filename)
	log.Println("Creating new blocklist in Umbrella: SOC Block", filename)
	destinationList, err := umbrellaService.CreateDestinationList("block", false, "SOC Block "+filename)
	if err != nil {
		return DestinationList{}, err
	}

	return destinationList, nil
}

func AddDestinationsToUmbrella(destinationsToAdd []string, destinationList DestinationList, umbrellaService UmbrellaService) (int, error) {
	originalDestinationCount := destinationList.Meta.DestinationCount
	finalDestinationCount := originalDestinationCount
	destinationsToAdd = utils.ValidateDestinationValues(destinationsToAdd)

	chunkSize := 500
	for i := 0; i < len(destinationsToAdd); i += chunkSize {
		end := i + chunkSize

		// Avoid going over the slice bounds
		if end > len(destinationsToAdd) {
			end = len(destinationsToAdd)
		}

		var addPayload []NewDestination
		for _, destination := range destinationsToAdd[i:end] {
			addPayload = append(addPayload, NewDestination{Destination: destination})
		}

		dl, err := umbrellaService.AddDestinations(destinationList.ID, addPayload)
		if err != nil {
			return 0, err
		} else {
			finalDestinationCount = dl.Meta.DestinationCount
		}
	}

	return finalDestinationCount, nil
}

func RemoveDestinationsFromUmbrella(destinationsToRemove []string, existingDestinations []Destination, destinationList DestinationList, umbrellaService UmbrellaService) (int, error) {
	originalDestinationCount := destinationList.Meta.DestinationCount
	finalDestinationCount := originalDestinationCount
	destinationMap := mapDestinationIDs(existingDestinations)

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

		dl, err := umbrellaService.DeleteDestinations(destinationList.ID, removePayload)
		if err != nil {
			return 0, err
		} else {
			finalDestinationCount = dl.Meta.DestinationCount
		}
	}

	return finalDestinationCount, nil
}

func FindMatchingDestinationList(blockFileName string, destinationLists []DestinationList) DestinationList {
	for _, destinationList := range destinationLists {
		if strings.Contains(destinationList.Name, blockFileName) {
			fmt.Println("Found match:", destinationList.Name)
			log.Println("Found match:", destinationList.Name)
			return destinationList
		}
	}
	return DestinationList{}
}
