/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/google/uuid"
	"strings"
	"io"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	artifactExtension = "tgz"
)

type FilesArtifactStore struct {
	fileCache 	*FileCache
}

func newFileStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) (*FilesArtifactStore, error) {
	return &FilesArtifactStore {
		fileCache: 	newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}, nil
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader) (string, error) {
	uuid, err := getUniversallyUniqueID()
	if err != nil{
		return "", stacktrace.Propagate(err, "Could not generate Universally Unique ID.")
	}
	filename := strings.Join([]string{uuid,artifactExtension}, ".")
	_, err = store.fileCache.AddFile(filename, reader)
	if err != nil{
		return "", stacktrace.Propagate(err, "Could not add file with UUID %s at %s.", uuid,
			       store.fileCache.absoluteDirpath)
	}
	return uuid, nil
}

// Get the file by uuid
func (store FilesArtifactStore) GetFilepathByUUID(uuid string) (string, error) {
	filename := strings.Join([]string{uuid, artifactExtension}, ".")
	enclaveDataDirFile, err := store.fileCache.GetFile(filename)
	if err != nil {
		return "", stacktrace.Propagate(err, "Could not retrieve file with UUID %s from %s.",
					uuid, store.fileCache.absoluteDirpath)
	}
	return enclaveDataDirFile.absoluteFilepath, nil
}


//There are some suggestions that go's implementation of uuid is not RFC compliant.
//If we can verify it is compliant, it would be better to use ipv6 as nodeID and interface name where the data came in.
//Just generating a random one for now.
func getUniversallyUniqueID() (string, error) {
	generatedUUID, err := uuid.NewRandom()
	uuidString := strings.Replace(generatedUUID.String(), "-", "",-1)
	return uuidString, err
}