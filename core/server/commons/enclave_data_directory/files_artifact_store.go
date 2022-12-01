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
	fileCache *FileCache
}

func newFilesArtifactStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache: newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
	}
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader) (FilesArtifactID, error) {
	newFilesArtifactUuid, err := NewFilesArtifactID()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}
	err = store.StoreFileToArtifactUUID(reader, newFilesArtifactUuid)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in storing data to files artifact with uuid '%v'", newFilesArtifactUuid)
	}
	return newFilesArtifactUuid, nil
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFileToArtifactUUID(reader io.Reader, targetArtifactUUID FilesArtifactID) error {
	filename := strings.Join(
		[]string{string(targetArtifactUUID), artifactExtension},
		".",
	)
	_, err := store.fileCache.AddFile(filename, reader)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"Could not store file '%s' to the file cache",
			filename,
		)
	}
	return nil
}

// Get the file by uuid
func (store FilesArtifactStore) GetFile(filesArtifactUuid FilesArtifactID) (*EnclaveDataDirFile, error) {
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

// RemoveFile: Remove the file by uuid
func (store FilesArtifactStore) RemoveFile(filesArtifactUuid FilesArtifactID) error {
	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)

	if err := store.fileCache.RemoveFile(filename); err != nil {
		return stacktrace.Propagate(err, "There was an error in removing '%v' from the file store", filename)
	}

	return nil
}
