/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
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
func (store FilesArtifactStore) StoreFile(reader io.Reader) (FilesArtifactUUID, error) {
	newFilesArtifactUuid, err := newFilesArtifactUUID()
	if err != nil{
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}
	filename := strings.Join(
		[]string{string(newFilesArtifactUuid),artifactExtension},
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
	return newFilesArtifactUuid, nil
}

// Get the file by uuid
func (store FilesArtifactStore) GetFile(filesArtifactUuid FilesArtifactUUID) (*EnclaveDataDirFile, error) {
	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
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
