/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
)

const (
	artifactExtension                     = "tgz"
	maxAllowedMatchesAgainstShortenedUuid = 1
	// TODO: this is something we can take a look in detail
	// but we with random numbers as suffix, we should always be able to have some unique name available
	maxFileArtifactNameRetriesDefault = 5
)

type FilesArtifactStore struct {
	fileCache                       *FileCache
	mutex                           *sync.RWMutex
	artifactNameToArtifactUuid      map[string]FilesArtifactUUID
	shortenedUuidToFullUuid         map[string][]FilesArtifactUUID
	maxRetriesToGetFileArtifactName int
	generateNatureThemeName         func() string
}

func newFilesArtifactStore(absoluteDirpath string, dirpathRelativeToDataDirRoot string) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache:                       newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex:                           &sync.RWMutex{},
		artifactNameToArtifactUuid:      make(map[string]FilesArtifactUUID),
		shortenedUuidToFullUuid:         make(map[string][]FilesArtifactUUID),
		maxRetriesToGetFileArtifactName: maxFileArtifactNameRetriesDefault,
		generateNatureThemeName:         name_generator.GenerateNatureThemeNameForFileArtifacts,
	}
}

// method needed for testing
func NewFilesArtifactStoreForTesting(
	absoluteDirpath string,
	dirpathRelativeToDataDirRoot string,
	artifactNameToArtifactUuid map[string]FilesArtifactUUID,
	shortenedUuidToFullUuid map[string][]FilesArtifactUUID,
	maxRetry int,
	nameGeneratorMock func() string,
) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache:                       newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex:                           &sync.RWMutex{},
		artifactNameToArtifactUuid:      artifactNameToArtifactUuid,
		shortenedUuidToFullUuid:         shortenedUuidToFullUuid,
		maxRetriesToGetFileArtifactName: maxRetry,
		generateNatureThemeName:         nameGeneratorMock,
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
func (store FilesArtifactStore) GetFile(artifactIdentifier string) (*EnclaveDataDirFile, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	filesArtifactUuid := FilesArtifactUUID(artifactIdentifier)
	file, err := store.getFileUnlocked(filesArtifactUuid)
	if err == nil {
		return file, nil
	}

	filesArtifactUuids, found := store.shortenedUuidToFullUuid[artifactIdentifier]
	if found {
		if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
			return nil, stacktrace.NewError("Tried using the shortened uuid '%v' to get file but found multiple matches '%v'. Use a complete uuid to be specific about what to get.", artifactIdentifier, filesArtifactUuids)
		}
		filesArtifactUuid := filesArtifactUuids[0]
		return store.getFileUnlocked(filesArtifactUuid)
	}

	filesArtifactUuid, found = store.artifactNameToArtifactUuid[artifactIdentifier]
	if found {
		return store.getFileUnlocked(filesArtifactUuid)
	}

	return nil, stacktrace.NewError("Couldn't find file for identifier '%v' tried, tried looking up UUID, shortened UUID and by name", artifactIdentifier)
}

// RemoveFile Remove the file by uuid, then by shortened uuid and then by name
func (store FilesArtifactStore) RemoveFile(artifactIdentifier string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	var filesArtifactUuid FilesArtifactUUID

	filesArtifactUuid = FilesArtifactUUID(artifactIdentifier)
	err := store.removeFileUnlocked(filesArtifactUuid)
	if err == nil {
		return nil
	}

	filesArtifactUuids, found := store.shortenedUuidToFullUuid[artifactIdentifier]
	if found {
		if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
			return stacktrace.NewError("Tried using the shortened uuid '%v' to remove file but found multiple matches '%v'. Use a complete uuid to be specific about what to delete.", artifactIdentifier, filesArtifactUuids)
		}
		filesArtifactUuid = filesArtifactUuids[0]
		return store.removeFileUnlocked(filesArtifactUuid)
	}

	filesArtifactUuid, found = store.artifactNameToArtifactUuid[artifactIdentifier]
	if found {
		return store.removeFileUnlocked(filesArtifactUuid)
	}

	return stacktrace.NewError("Couldn't find file for identifier '%v' tried, tried looking up UUID, shortened UUID and by name", artifactIdentifier)
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

// CheckIfArtifactNameExists - It checks whether the FileArtifact with a name exists or not
func (store FilesArtifactStore) CheckIfArtifactNameExists(artifactName string) bool {

	_, found := store.artifactNameToArtifactUuid[artifactName]
	return found
}

func (store FilesArtifactStore) GenerateUniqueNameForFileArtifact() string {
	var maybeUniqueName string

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// try to find unique nature theme random generator
	for i := 0; i <= store.maxRetriesToGetFileArtifactName; i++ {
		maybeUniqueName = store.generateNatureThemeName()
		_, found := store.artifactNameToArtifactUuid[maybeUniqueName]
		if !found {
			return maybeUniqueName
		}
	}

	// if unique name not found, append a random number after the last found random name
	additionalSuffix := 1
	maybeUniqueNameWithRandomNumber := fmt.Sprintf("%v-%v", maybeUniqueName, additionalSuffix)

	_, found := store.artifactNameToArtifactUuid[maybeUniqueNameWithRandomNumber]
	for found {
		additionalSuffix = additionalSuffix + 1
		maybeUniqueNameWithRandomNumber = fmt.Sprintf("%v-%v", maybeUniqueName, additionalSuffix)
		_, found = store.artifactNameToArtifactUuid[maybeUniqueNameWithRandomNumber]
	}

	logrus.Warnf("Cannot find unique name generator, therefore using a name with a number %v", maybeUniqueNameWithRandomNumber)
	return maybeUniqueNameWithRandomNumber
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
