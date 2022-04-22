/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/google/uuid"
	"strings"
	"sync"
	"io"
	"github.com/kurtosis-tech/stacktrace"
	"path/filepath"
)

const (
	artifactExtension = "tgz"
)

type FileStore struct {
	uuid		string
	fileCache 	*FileCache
	mutex 		*sync.Mutex
}

func newFileStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) (*FileStore, error) {
	uuid, err := getUniversallyUniqueID()

	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not create Universally Unique ID for FileStore")
	}

	return &FileStore {
		uuid: 	 	uuid,
		fileCache: 	newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}, nil
}

// StoreFile: Saves file to disk.
func (store FileStore) StoreFile(reader io.Reader) (*EnclaveDataDirFile, error) {
	//Can't mutex lock in this case because underlying filecache will be trying to do that.
	return store.fileCache.AddFile(store.GetFileName(), reader)
}

//Get the absolute path to the file.
func (store FileStore) GetFilePath() string {
	return filepath.Join(store.fileCache.absoluteDirpath, store.GetFileName())
}

//Returns a file name created from the uuid and the artifactExtension.
func (store FileStore) GetFileName() string {
	return strings.Join([]string{store.uuid, artifactExtension}, ".")
}

//There are some suggestions that go's implementation of uuid is not RFC compliant.
//If we can verify it is compliant, it would be better to use ipv6 as nodeID and interface name where the data came in.
//Just generating a random one for now.
func getUniversallyUniqueID() (string, error) {
	generatedUUID, err := uuid.NewRandom()
	uuidString := strings.Replace(generatedUUID.String(), "-", "",-1)
	return uuidString, err
}