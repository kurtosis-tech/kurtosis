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
)

//FileHandler: A class that is responsible for saving data to disk and cleaning up temporary files.
//currentWorkingDirectory - A directory that the FileHandler will append all other relative paths to.
type FileHandler struct {
	currentWorkingDirectory string
}

func newFileHandler(currentWorkingDirectory string) *FileHandler {
	return &FileHandler{
		currentWorkingDirectory: currentWorkingDirectory,
	}
}

//ChangeDirectory: Changes the working directory so new files can have cleaner relative save paths in code.
//directoryToChangeTo: A relative or absolute directory we wish to change to. Accepts "../" "./" "/" and no prefixes.
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
// fileName - The name of the file we wish to save the bytes to.
// relativeFolder - The folder name, relative to the currentWorkingDirectory, we want to save in.
// content - The data, in the form of a byte array, we wish to save.
func (handler *FileHandler) SaveBytesToPath(fileName string, relativeFolder string, content []byte) error{
	absolutePath := filepath.Join(handler.currentWorkingDirectory, relativeFolder, fileName)
	file, err := os.Create(absolutePath)

	if err != nil {
		fmt.Printf("Error creating %s.\n %s", fileName, err.Error())
		return err
	}

	defer file.Close()

	_, writeErr := file.Write(content)
	if writeErr != nil {
		fmt.Printf("Error writing %s.\n %s", fileName, writeErr.Error())
		return writeErr
	}

	return nil
}