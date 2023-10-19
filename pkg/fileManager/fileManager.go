package fileManager

import (
	"bytes"
	"io/fs"
	"os"
	"path"
	"strings"
)

func GetHomedir() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		homedir = "."
	}

	return homedir
}

func WriteToFile(filepath string, content []byte) error {
	filedir := path.Dir(filepath)
	os.MkdirAll(filedir, 0755)
	err := os.WriteFile(filepath, content, 0644)
	return err
}

func ReadFile(filepath string) ([]byte, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func FileInfo(filepath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}

func ReadFolder(folderPath string) ([]fs.FileInfo, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, err
	}

	var fileInfos []fs.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}
		fileInfos = append(fileInfos, info)
	}

	return fileInfos, nil
}

func IsExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ToLines(content []byte) []string {

	lines := bytes.Split(content, []byte("\n"))
	var stringLines []string
	for _, line := range lines {
		linestring := string(line)
		linestring = strings.TrimSuffix(linestring, "\r") // For Windows
		stringLines = append(stringLines, linestring)
	}

	return stringLines
}
