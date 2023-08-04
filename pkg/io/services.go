package io

import (
	"bufio"
	"log"
	"os"

	"github.com/thegrumpyape/umbrellasync/pkg/api"
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
