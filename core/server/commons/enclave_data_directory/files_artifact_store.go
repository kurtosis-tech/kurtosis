/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/google/uuid"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"strings"
)

const (
	artifactExtension = "tgz"
)

type FilesArtifactStore struct {
	fileCache 	*FileCache
}

func newFilesArtifactStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactStore {
	return &FilesArtifactStore {
		fileCache: 	newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader) (service.FilesArtifactID, error) {
	newFileUuid, err := getUniversallyUniqueID()
	if err != nil{
		return "", stacktrace.Propagate(err, "Could not generate Universally Unique ID.")
	}
	filename := strings.Join(
		[]string{newFileUuid,artifactExtension},
		".",
	)
	_, err = store.fileCache.AddFile(filename, reader)
	if err != nil{
		return "", stacktrace.Propagate(
			err,
			"Could not store file '%s' to the file cache",
			filename,
		)
	}
	return service.FilesArtifactID(newFileUuid), nil
}

// Get the file by uuid
func (store FilesArtifactStore) GetFile(filesArtifactId service.FilesArtifactID) (*EnclaveDataDirFile, error) {
	filename := strings.Join(
		[]string{string(filesArtifactId), artifactExtension},
		".",
	)
	enclaveDataDirFile, err := store.fileCache.GetFile(filename)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"Could not retrieve file with filename '%s' from the file cache",
			filename,
		)
	}
	return enclaveDataDirFile, nil
}


//There are some suggestions that go's implementation of uuid is not RFC compliant.
//If we can verify it is compliant, it would be better to use ipv6 as nodeID and interface name where the data came in.
//Just generating a random one for now.
func getUniversallyUniqueID() (string, error) {
	generatedUUID, err := uuid.NewRandom()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred generating a UUID")
	}
	return generatedUUID.String(), nil
}