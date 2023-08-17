package io

import (
	"bufio"
	"log"
	"os"

	"github.com/thegrumpyape/umbrellasync/pkg/comparison"
	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

func NewBlockFile(path string) (BlockFile, error) {
	var lines []string

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	// Get file info
	fileinfo, err := file.Stat()
	if err != nil {
		return BlockFile{}, err
	}

	// Get file data as slice of lines
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return BlockFile{}, err
	}

	return BlockFile{Path: file.Name(), Name: fileinfo.Name(), Data: lines}, nil
}

func (f *BlockFile) Sync(umbrellaService umbrella.UmbrellaService, destinationList umbrella.DestinationList) error {
	chunkSize := 500

	log.Println("Syncing file:", f.Name)

	destinations, err := umbrellaService.GetDestinations(destinationList.ID, 100)
	if err != nil {
		log.Fatal(err)
	}

	var destinationData []string
	for _, destination := range destinations {
		destinationData = append(destinationData, destination.Destination)
	}

	destinationsToAdd, destinationsToRemove := comparison.Compare(f.Data, destinationData)

	if len(destinationsToAdd) != 0 {
		umbrellaService.AddDestinations(destinationList, destinationsToAdd, chunkSize)
	}

	if len(destinationsToRemove) != 0 {
		umbrellaService.DeleteDestinations(destinationList, destinationsToRemove, destinations, chunkSize)
	}

	return nil
}
