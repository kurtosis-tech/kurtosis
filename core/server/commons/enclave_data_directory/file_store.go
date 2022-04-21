/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"sync"
	"github.com/google/uuid"
	"strings"
	"fmt"
)

type FileStore struct {
	uuid							string
	absolutePath 					string
	dirpathRelativeToDataDirRoot 	string
	mutex 							*sync.Mutex
}

func newFileStore(uuid string, absolutePath string, dirpathRelativeToDataDirRoot string) *FileStore {
	return &FileStore {
		uuid: 							uuid,
		absolutePath: 					absolutePath,
		dirpathRelativeToDataDirRoot: 	dirpathRelativeToDataDirRoot,
		mutex: 							&sync.Mutex{},
	}
}

// StoreFile: Saves file to disk and returns a FileStore struct.
func StoreFile(fileName string, filedata []byte) (*FileStore, error) {
	generatedUUID, _ := getUniversallyUniqueID()
	uuidString := strings.Replace(generatedUUID.String(), "-", "",-1)
	uuidFileName := strings.Join([]string{uuidString, fileName}, "_")
	relativeFolder := "" //Don't need a folder?

	handler, err := createFileHandler()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	absolutePath, saveErr := handler.SaveBytesToPath(uuidFileName, relativeFolder, filedata)

	if saveErr != nil {
		fmt.Println(saveErr.Error())
		return nil, saveErr
	}
	return newFileStore(uuidString, absolutePath, relativeFolder), nil
}

//There are some suggestions that go's implementation of uuid is not RFC compliant
//If we can verify it is, it would be better to use ipv6 as nodeID and interface name where the data came in.
//Just generating a random one for now.
func getUniversallyUniqueID() (uuid.UUID, error) {
	return uuid.NewRandom()
}
