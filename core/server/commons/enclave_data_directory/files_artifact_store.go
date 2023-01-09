/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"io"
	"strings"
	"sync"
)

const (
	artifactExtension                     = "tgz"
	maxAllowedMatchesAgainstShortenedUuid = 1
)

type FilesArtifactStore struct {
	fileCache                  *FileCache
	mutex                      *sync.RWMutex
	artifactNameToArtifactUuid map[string]FilesArtifactUUID
	shortenedUuidToFullUuid    map[string][]FilesArtifactUUID
}

func newFilesArtifactStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache:                  newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex:                      &sync.RWMutex{},
		artifactNameToArtifactUuid: make(map[string]FilesArtifactUUID),
		shortenedUuidToFullUuid:    make(map[string][]FilesArtifactUUID),
	}
}

// StoreFile Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader, artifactName string) (FilesArtifactUUID, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if _, found := store.artifactNameToArtifactUuid[artifactName]; found {
		return "", stacktrace.NewError("Files artifact name '%v' has already been used", artifactName)
	}
	filesArtifactUuid, err := store.storeFilesToArtifactUuidUnlocked(reader)
	if err != nil {
		return "", err
	}
	store.artifactNameToArtifactUuid[artifactName] = filesArtifactUuid
	return filesArtifactUuid, nil
}

// GetFile Get the file by uuid, then by shortened uuid and finally by name
func (store FilesArtifactStore) GetFile(artifactReference string) (*EnclaveDataDirFile, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if uuid_generator.IsUUID(artifactReference) {
		filesArtifactUuid := FilesArtifactUUID(artifactReference)
		file, err := store.getFileUnlocked(filesArtifactUuid)
		if err == nil {
			return file, nil
		}
	}

	if uuid_generator.ISShortenedUUID(artifactReference) {
		filesArtifactUuids, found := store.shortenedUuidToFullUuid[artifactReference]
		if found {
			if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
				return nil, stacktrace.NewError("Tried using the shortened uuid '%v' to get file but found multiple matches '%v'. Use a complete uuid to be specific about what to get.", artifactReference, filesArtifactUuids)
			}
			filesArtifactUuid := filesArtifactUuids[0]
			return store.getFileUnlocked(filesArtifactUuid)
		}
	}

	filesArtifactUuid, found := store.artifactNameToArtifactUuid[artifactReference]
	if found {
		return store.getFileUnlocked(filesArtifactUuid)
	}

	return nil, stacktrace.NewError("Couldn't find file for reference '%v' tried, tried looking up UUID, shortened UUID and by name", artifactReference)
}

// RemoveFile Remove the file by uuid, then by shortened uuid and then by name
func (store FilesArtifactStore) RemoveFile(artifactReference string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	var filesArtifactUuid FilesArtifactUUID

	if uuid_generator.IsUUID(artifactReference) {
		filesArtifactUuid = FilesArtifactUUID(artifactReference)
		err := store.removeFileUnlocked(filesArtifactUuid)
		if err == nil {
			return nil
		}
	}

	if uuid_generator.ISShortenedUUID(artifactReference) {
		filesArtifactUuids, found := store.shortenedUuidToFullUuid[artifactReference]
		if found {
			if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
				return stacktrace.NewError("Tried using the shortened uuid '%v' to remove file but found multiple matches '%v'. Use a complete uuid to be specific about what to delete.", artifactReference, filesArtifactUuids)
			}
			filesArtifactUuid = filesArtifactUuids[0]
			return store.removeFileUnlocked(filesArtifactUuid)
		}
	}

	filesArtifactUuid, found := store.artifactNameToArtifactUuid[artifactReference]
	if found {
		return store.removeFileUnlocked(filesArtifactUuid)
	}

	return stacktrace.NewError("Couldn't find file for reference '%v' tried, tried looking up UUID, shortened UUID and by name", artifactReference)
}

func (store FilesArtifactStore) ListFiles() map[string]bool {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	artifactNameSet := make(map[string]bool)
	for artifactName := range store.artifactNameToArtifactUuid {
		artifactNameSet[artifactName] = true
	}
	return artifactNameSet
}

// storeFilesToArtifactUuidUnlocked this is an non thread method to be used from thread safe contexts
func (store FilesArtifactStore) storeFilesToArtifactUuidUnlocked(reader io.Reader) (FilesArtifactUUID, error) {
	filesArtifactUuid, err := NewFilesArtifactUUID()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}

	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	_, err = store.fileCache.AddFile(filename, reader)
	if err != nil {
		return "", stacktrace.Propagate(
			err,
			"Could not store file '%s' to the file cache",
			filename,
		)
	}
	shortenedUuidSlice := store.shortenedUuidToFullUuid[uuid_generator.ShortenedUUIDString(string(filesArtifactUuid))]
	store.shortenedUuidToFullUuid[uuid_generator.ShortenedUUIDString(string(filesArtifactUuid))] = append(shortenedUuidSlice, filesArtifactUuid)
	return filesArtifactUuid, nil
}

// getFileUnlocked this is not thread safe, must be used from a thread safe context
func (store FilesArtifactStore) getFileUnlocked(filesArtifactUuid FilesArtifactUUID) (*EnclaveDataDirFile, error) {
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

// removeFileUnlocked this is not thread safe, must be used from a thread safe context
func (store FilesArtifactStore) removeFileUnlocked(filesArtifactUuid FilesArtifactUUID) error {
	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)

	if err := store.fileCache.RemoveFile(filename); err != nil {
		return stacktrace.Propagate(err, "There was an error in removing '%v' from the file store", filename)
	}
	for name, artifactUuid := range store.artifactNameToArtifactUuid {
		if artifactUuid == filesArtifactUuid {
			delete(store.artifactNameToArtifactUuid, name)
		}
	}
	shortenedUuid := uuid_generator.ShortenedUUIDString(string(filesArtifactUuid))
	artifactUuids, found := store.shortenedUuidToFullUuid[shortenedUuid]
	if found {
		var targetArtifactIdx int
		for index, artifactUuid := range artifactUuids {
			if artifactUuid == filesArtifactUuid {
				targetArtifactIdx = index
			}
		}
		// if there's only one matching uuid we delete the shortened uuid
		if len(artifactUuids) == 1 {
			delete(store.shortenedUuidToFullUuid, shortenedUuid)
			return nil
		}
		// otherwise we just delete the target artifact uuid
		store.shortenedUuidToFullUuid[shortenedUuid] = append(artifactUuids[0:targetArtifactIdx], artifactUuids[targetArtifactIdx+1:]...)
	}

	return nil
}
