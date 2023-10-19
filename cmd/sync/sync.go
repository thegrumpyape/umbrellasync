package sync

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thegrumpyape/umbrellasync/pkg/configurationManager"
	"github.com/thegrumpyape/umbrellasync/pkg/fileManager"
	"github.com/thegrumpyape/umbrellasync/pkg/logging"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

type SyncCommandDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
	Logger               logging.Logger
}

type SyncUmbrellaDependencies struct {
	ConfigurationManager *configurationManager.ConfigurationManager
	UmbrellaConnector    *umbrella.UmbrellaConnector
	Logger               logging.Logger
}

func New(deps *SyncCommandDependencies) *cobra.Command {
	syncCmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync threat intel with Umbrella",
		Long:  "Sync threat intel with Umbrella",
		PreRun: func(cmd *cobra.Command, cmdArgs []string) {

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			umbrellaClient := umbrella.CreateUmbrellaClient(*deps.ConfigurationManager, deps.Logger)
			umbrellaConnector, err := umbrella.New(umbrellaClient, *deps.ConfigurationManager, deps.Logger)
			if err != nil {
				log.Fatal(err)
			}

			syncUmbrellaDeps := &SyncUmbrellaDependencies{
				ConfigurationManager: deps.ConfigurationManager,
				UmbrellaConnector:    umbrellaConnector,
				Logger:               deps.Logger,
			}

			return executeSync(syncUmbrellaDeps)
		},
		PostRun: func(cmd *cobra.Command, args []string) {

		},
	}

	return syncCmd
}

func executeSync(deps *SyncUmbrellaDependencies) error {
	values, ok := deps.ConfigurationManager.Get("files").([]interface{})
	if !ok {
		return fmt.Errorf("Could not get files from config.yaml")
	}

	filepaths := make([]string, len(values))
	for i, v := range values {
		str, ok := v.(string)
		if !ok {
			return fmt.Errorf("Element at index %d is not a string", i)
		}
		filepaths[i] = str
	}

	destinationLists, err := deps.UmbrellaConnector.GetDestinationLists(100)
	if err != nil {
		return err
	}

	for _, filepath := range filepaths {
		fileInfo, err := fileManager.FileInfo(filepath)
		if err != nil {
			return err
		}
		fileData, err := fileManager.ReadFile(filepath)
		if err != nil {
			return err
		}

		// Sync blockfile
		deps.Logger.Info("Syncing file ", filepath)
		var matchingDestinationList umbrella.DestinationList

		for _, dl := range destinationLists {
			if strings.Contains(dl.Name, fileInfo.Name()) {
				deps.Logger.Info("Found matching destination list ", dl.Name)
				matchingDestinationList = dl
				break
			}
		}

		// if no match is found, create a new destination list
		if matchingDestinationList == (umbrella.DestinationList{}) {
			var err error
			matchingDestinationList, err = deps.UmbrellaConnector.CreateDestinationList("block", false, "SOC Block "+fileInfo.Name())
			deps.Logger.Info("Created destination list: ", matchingDestinationList.Name)
			if err != nil {
				return err
			}
		}

		deps.Logger.Info("Reading ", matchingDestinationList.Meta.DestinationCount, " destinations from ", matchingDestinationList.Name)
		destinations, err := deps.UmbrellaConnector.GetDestinations(matchingDestinationList.ID, 100)
		if err != nil {
			return err
		}

		var destinationData []string
		for _, destination := range destinations {
			destinationData = append(destinationData, destination.Destination)
		}

		fileLines := fileManager.ToLines(fileData)
		destinationsToAdd, destinationsToRemove := compareLists(fileLines, destinationData)

		if len(destinationsToAdd) != 0 {
			deps.Logger.Info(len(destinationsToAdd), " destinations missing from ", matchingDestinationList.Name)
			matchingDestinationList, err = deps.UmbrellaConnector.AddDestinations(matchingDestinationList, destinationsToAdd, 500)
			if err != nil {
				return err
			}
		}

		if len(destinationsToRemove) != 0 {
			deps.Logger.Info(len(destinationsToRemove), " destinations missing from ", filepath)
			matchingDestinationList, err = deps.UmbrellaConnector.DeleteDestinations(matchingDestinationList, destinationsToRemove, destinations, 500)
			if err != nil {
				return err
			}
		}

		//deps.Logger.Debug("Sleeping for 60 seconds")
		//time.Sleep(60 * time.Second)
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
