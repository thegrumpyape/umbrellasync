package umbrellasync

import (
	"log"
	"strings"

	"github.com/thegrumpyape/umbrellasync/pkg/blockfile"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

type UmbrellaSync struct {
	client           umbrella.UmbrellaService
	destinationLists []umbrella.DestinationList
	logger           *log.Logger
}

func New(hostname string, version string, key string, secret string, logger *log.Logger) (UmbrellaSync, error) {
	// create umbrella service
	umbrellaService, err := umbrella.CreateClient(hostname, version, key, secret, logger)
	if err != nil {
		return UmbrellaSync{}, err
	}

	// get destination lists
	destinationLists, err := umbrellaService.GetDestinationLists(100)
	if err != nil {
		return UmbrellaSync{}, err
	}

	return UmbrellaSync{client: umbrellaService, destinationLists: destinationLists, logger: logger}, nil
}

func (u *UmbrellaSync) Sync(blockFile blockfile.BlockFile) error {
	// find matching destination list for block file
	matchingDestinationList := umbrella.DestinationList{}

	for _, dl := range u.destinationLists {
		if strings.Contains(dl.Name, blockFile.Name) {
			matchingDestinationList = dl
			break
		}
	}

	// if no match is found, create a new destination list
	if matchingDestinationList == (umbrella.DestinationList{}) {
		u.logger.Println("Creating new blocklist in Umbrella: SOC Block", blockFile.Name)
		var err error
		matchingDestinationList, err = u.client.CreateDestinationList("block", false, "SOC Block "+blockFile.Name)
		if err != nil {
			return err
		}
	}

	destinations, err := u.client.GetDestinations(matchingDestinationList.ID, 100)
	if err != nil {
		return err
	}

	var destinationData []string
	for _, destination := range destinations {
		destinationData = append(destinationData, destination.Destination)
	}

	destinationsToAdd, destinationsToRemove := compareLists(blockFile.Data, destinationData)

	if len(destinationsToAdd) != 0 {
		matchingDestinationList, err = u.client.AddDestinations(matchingDestinationList, destinationsToAdd, 500)
		if err != nil {
			return err
		}
	}

	if len(destinationsToRemove) != 0 {
		matchingDestinationList, err = u.client.DeleteDestinations(matchingDestinationList, destinationsToRemove, destinations, 500)
		if err != nil {
			return err
		}
	}

	return nil
}

// Compares BlockFile with Destinations from DestinationList
func compareLists(blocklistData []string, destinationListData []string) ([]string, []string) {
	var destsToAdd, destsToDelete []string

	blMap, dlMap := make(map[string]bool), make(map[string]bool)

	for _, item := range blocklistData {
		blMap[item] = true
	}

	for _, item := range destinationListData {
		dlMap[item] = true
	}

	for key := range blMap {
		if !dlMap[key] {
			destsToAdd = append(destsToAdd, key)
		}
	}

	for key := range dlMap {
		if !blMap[key] {
			destsToDelete = append(destsToDelete, key)
		}
	}

	return destsToAdd, destsToDelete
}
