package blockfile

import (
	"bufio"
	"os"
)

func New(path string) (BlockFile, error) {
	var lines []string

	file, err := os.Open(path)
	if err != nil {
		return BlockFile{}, err
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
