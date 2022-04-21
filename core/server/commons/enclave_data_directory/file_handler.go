/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"os"
	"path/filepath"
	"fmt"
	"strings"
	"errors"
)

//FileHandler: A class that is responsible for saving data to disk and cleaning up temporary files.
type FileHandler struct {
	currentWorkingDirectory string
}

func newFileHandler(currentWorkingDirectory string) *FileHandler {
	return &FileHandler{
		currentWorkingDirectory: currentWorkingDirectory,
	}
}

func createFileHandler() (*FileHandler, error) {
	currentDirectory, err := os.Getwd()
	if err != nil {
		message := "Couldn't get working directory."
		return nil, errors.New(message)
	}
	return newFileHandler(currentDirectory), nil
}

//ChangeDirectory: Changes the working directory so new files can have cleaner relative save paths in code.
//TODO: Check to make sure that the directory is even a valid location or not.
func (handler *FileHandler) ChangeDirectory(directoryToChangeTo string){
	isRelative := strings.HasPrefix(directoryToChangeTo, "../")
	isRelative = strings.HasPrefix(directoryToChangeTo, "./") || isRelative
	isRelative = !strings.HasPrefix(directoryToChangeTo, "/") || isRelative //Just assume user means "./"

	if isRelative {
		handler.currentWorkingDirectory = filepath.Join(handler.currentWorkingDirectory, directoryToChangeTo)
	} else {
		handler.currentWorkingDirectory = directoryToChangeTo
	}
}

// SaveBytesToPath: Save bytes directly to the disk.
// Replace content with reader
func (handler *FileHandler) SaveBytesToPath(fileName string, relativeFolder string, content []byte) (string, error) {
	absolutePath := filepath.Join(handler.currentWorkingDirectory, relativeFolder, fileName)
	file, err := os.Create(absolutePath)

	if err != nil {
		fmt.Printf("Error creating %s.\n %s", fileName, err.Error())
		return absolutePath, err
	}

	defer file.Close()

	_, writeErr := file.Write(content)
	if writeErr != nil {
		fmt.Printf("Error writing %s.\n %s", fileName, writeErr.Error())
		return absolutePath, writeErr
	}

	return absolutePath, nil
}