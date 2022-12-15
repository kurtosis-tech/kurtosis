/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"strings"
	"sync"
)

const (
	artifactExtension = "tgz"
)

type FilesArtifactStore struct {
	fileCache *FileCache
	mutex *sync.RWMutex
}

func newFilesArtifactStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache: newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex: &sync.RWMutex{},
	}
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader) (FilesArtifactID, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	newFilesArtifactId, err := NewFilesArtifactID()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}
	err = store.storeFilesToArtifactUuidUnlocked(reader, newFilesArtifactId)
	if err != nil {
		return "", stacktrace.Propagate(err, "There was an error in storing data to files artifact with uuid '%v'", newFilesArtifactId)
	}
	return newFilesArtifactId, nil
}

// StoreFile: Saves file to disk.
func (store FilesArtifactStore) StoreFileToArtifactUUID(reader io.Reader, targetArtifactUUID FilesArtifactID) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return store.storeFilesToArtifactUuidUnlocked(reader, targetArtifactUUID)
}

// Get the file by uuid
func (store FilesArtifactStore) GetFile(filesArtifactUuid FilesArtifactID) (*EnclaveDataDirFile, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

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
	store.mutex.Lock()
	defer store.mutex.Unlock()

	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)

	if err := store.fileCache.RemoveFile(filename); err != nil {
		return stacktrace.Propagate(err, "There was an error in removing '%v' from the file store", filename)
	}

	return nil
}

// storeFilesToArtifactUuidUnlocked this is an non thread method to be used from thread safe contexts
func (store FilesArtifactStore) storeFilesToArtifactUuidUnlocked(reader io.Reader, targetArtifactUUID FilesArtifactID) error {
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
