/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_data_directory

import (
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db/file_artifacts_db"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/dzobbe/PoTE-kurtosis/name_generator"
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
	fileArtifactDb                  *file_artifacts_db.FileArtifactPersisted
	maxRetriesToGetFileArtifactName int
	generateNatureThemeName         func() string
}

func newFilesArtifactStoreFromDb(absoluteDirpath string, dirpathRelativeToDataDirRoot string, db *file_artifacts_db.FileArtifactPersisted) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache:                       newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex:                           &sync.RWMutex{},
		maxRetriesToGetFileArtifactName: maxFileArtifactNameRetriesDefault,
		generateNatureThemeName:         name_generator.GenerateNatureThemeNameForFileArtifacts,
		fileArtifactDb:                  db,
	}
}

// method needed for testing
func NewFilesArtifactStoreForTesting(
	absoluteDirpath string,
	dirpathRelativeToDataDirRoot string,
	fileArtifactDb *file_artifacts_db.FileArtifactPersisted,
	maxRetry int,
	nameGeneratorMock func() string,
) *FilesArtifactStore {
	return &FilesArtifactStore{
		fileCache:                       newFileCache(absoluteDirpath, dirpathRelativeToDataDirRoot),
		mutex:                           &sync.RWMutex{},
		fileArtifactDb:                  fileArtifactDb,
		maxRetriesToGetFileArtifactName: maxRetry,
		generateNatureThemeName:         nameGeneratorMock,
	}
}

// StoreFile Saves file to disk.
func (store FilesArtifactStore) StoreFile(reader io.Reader, contentMd5 []byte, artifactName string) (FilesArtifactUUID, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	filesArtifactUuid, err := NewFilesArtifactUUID()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}

	if _, found := store.fileArtifactDb.GetArtifactUuid(artifactName); found {
		return "", stacktrace.NewError("Files artifact name '%v' has already been used", artifactName)
	}

	err = store.storeFilesToArtifactUuidUnlocked(artifactName, filesArtifactUuid, reader, contentMd5)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new files artifact UUID")
	}
	if err := store.fileArtifactDb.Persist(); err != nil {
		return "", stacktrace.Propagate(err, "Failed persisting data on file artifacts db")
	}
	return filesArtifactUuid, nil
}

func (store FilesArtifactStore) UpdateFile(filesArtifactUuid FilesArtifactUUID, reader io.Reader, contentMd5 []byte) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	artifactName, err := store.removeFileUnlocked(filesArtifactUuid)
	if err != nil {
		return stacktrace.Propagate(err, "Error removing the files artifact '%s' prior to persisting its new content", filesArtifactUuid)
	}
	if err = store.storeFilesToArtifactUuidUnlocked(artifactName, filesArtifactUuid, reader, contentMd5); err != nil {
		return stacktrace.Propagate(err, "Error persisting updated content for files artifact '%s'", filesArtifactUuid)
	}
	return nil
}

// GetFile Get the file by uuid, then by shortened uuid and finally by name
func (store FilesArtifactStore) GetFile(artifactIdentifier string) (FilesArtifactUUID, *EnclaveDataDirFile, []byte, bool, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	filesArtifactUuid := FilesArtifactUUID(artifactIdentifier)
	fileArtifactUuid, file, contentMd5, found, err := store.getFileUnlocked(filesArtifactUuid)
	if err == nil {
		return fileArtifactUuid, file, contentMd5, found, nil
	}

	filesArtifactUuids, found := store.fileArtifactDb.GetFullUuid(artifactIdentifier)
	if found {
		if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
			return "", nil, nil, false, stacktrace.NewError("Tried using the shortened uuid '%v' to get file but found multiple matches '%v'. Use a complete uuid to be specific about what to get.", artifactIdentifier, filesArtifactUuids)
		}
		filesArtifactUuid = FilesArtifactUUID(filesArtifactUuids[0])
		return store.getFileUnlocked(filesArtifactUuid)
	}

	filesArtifactUuidGet, found := store.fileArtifactDb.GetArtifactUuid(artifactIdentifier)
	if found {
		return store.getFileUnlocked(FilesArtifactUUID(filesArtifactUuidGet))
	}

	if err := store.fileArtifactDb.Persist(); err != nil {
		return "", nil, nil, false, stacktrace.Propagate(err, "Failed persisting data on file artifacts db")
	}
	return "", nil, nil, false, nil
}

// RemoveFile Remove the file by uuid, then by shortened uuid and then by name
func (store FilesArtifactStore) RemoveFile(artifactIdentifier string) (deferedErr error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	var filesArtifactUuid FilesArtifactUUID

	filesArtifactUuid = FilesArtifactUUID(artifactIdentifier)
	_, err := store.removeFileUnlocked(filesArtifactUuid)
	if err == nil {
		return nil
	}

	filesArtifactUuids, found := store.fileArtifactDb.GetFullUuid(artifactIdentifier)
	if found {
		if len(filesArtifactUuids) > maxAllowedMatchesAgainstShortenedUuid {
			return stacktrace.NewError("Tried using the shortened uuid '%v' to remove file but found multiple matches '%v'. Use a complete uuid to be specific about what to delete.", artifactIdentifier, filesArtifactUuids)
		}
		filesArtifactUuid = FilesArtifactUUID(filesArtifactUuids[0])
		_, err = store.removeFileUnlocked(filesArtifactUuid)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred removing files artifact with UUID '%s'", filesArtifactUuid)
		}
		return nil
	}

	filesArtifactUuidGet, found := store.fileArtifactDb.GetArtifactUuid(artifactIdentifier)
	if found {
		_, err = store.removeFileUnlocked(FilesArtifactUUID(filesArtifactUuidGet))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred removing files artifact with UUID '%s'", filesArtifactUuid)
		}
		return nil
	}

	if err := store.fileArtifactDb.Persist(); err != nil {
		return stacktrace.Propagate(err, "Failed persisting data on file artifacts db")
	}
	return stacktrace.NewError("Couldn't find file for identifier '%v', tried looking up UUID, shortened UUID and by name", artifactIdentifier)
}

func (store FilesArtifactStore) ListFiles() map[string]bool {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	return store.fileArtifactDb.ListFiles()
}

func (store FilesArtifactStore) GetFileNamesAndUuids() []FileNameAndUuid {
	store.mutex.RLock()
	defer store.mutex.RUnlock()
	var nameAndUuids []FileNameAndUuid
	for artifactName, artifactUuid := range store.fileArtifactDb.GetArtifactUuidMap() {
		nameAndUuid := FileNameAndUuid{
			uuid: FilesArtifactUUID(artifactUuid),
			name: artifactName,
		}
		nameAndUuids = append(nameAndUuids, nameAndUuid)
	}
	return nameAndUuids
}

// CheckIfArtifactNameExists - It checks whether the FileArtifact with a name exists or not
func (store FilesArtifactStore) CheckIfArtifactNameExists(artifactName string) bool {
	_, found := store.fileArtifactDb.GetArtifactUuid(artifactName)
	return found
}

func (store FilesArtifactStore) GenerateUniqueNameForFileArtifact() string {
	var maybeUniqueName string

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	// try to find unique nature theme random generator
	for i := 0; i <= store.maxRetriesToGetFileArtifactName; i++ {
		maybeUniqueName = store.generateNatureThemeName()
		_, found := store.fileArtifactDb.GetArtifactUuid(maybeUniqueName)
		if !found {
			return maybeUniqueName
		}
	}

	// if unique name not found, append a random number after the last found random name
	additionalSuffix := 1
	maybeUniqueNameWithRandomNumber := fmt.Sprintf("%v-%v", maybeUniqueName, additionalSuffix)

	_, found := store.fileArtifactDb.GetArtifactUuid(maybeUniqueNameWithRandomNumber)
	for found {
		additionalSuffix = additionalSuffix + 1
		maybeUniqueNameWithRandomNumber = fmt.Sprintf("%v-%v", maybeUniqueName, additionalSuffix)
		_, found = store.fileArtifactDb.GetArtifactUuid(maybeUniqueNameWithRandomNumber)
	}

	logrus.Warnf("Cannot find unique name generator, therefore using a name with a number %v", maybeUniqueNameWithRandomNumber)
	return maybeUniqueNameWithRandomNumber
}

// storeFilesToArtifactUuidUnlocked this is an non thread method to be used from thread safe contexts
func (store FilesArtifactStore) storeFilesToArtifactUuidUnlocked(filesArtifactName string, filesArtifactUuid FilesArtifactUUID, reader io.Reader, contentMd5 []byte) error {

	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	_, err := store.fileCache.AddFile(filename, reader)
	if err != nil {
		return stacktrace.Propagate(err, "Could not store file '%s' to the file cache", filename)
	}
	shortenedUuidSlice, _ := store.fileArtifactDb.GetFullUuid(uuid_generator.ShortenedUUIDString(string(filesArtifactUuid)))
	store.fileArtifactDb.SetFullUuid(uuid_generator.ShortenedUUIDString(string(filesArtifactUuid)), append(shortenedUuidSlice, string(filesArtifactUuid)))
	store.fileArtifactDb.SetContentMd5(string(filesArtifactUuid), contentMd5)
	store.fileArtifactDb.SetArtifactUuid(filesArtifactName, string(filesArtifactUuid))
	return nil
}

// getFileUnlocked this is not thread safe, must be used from a thread safe context
func (store FilesArtifactStore) getFileUnlocked(filesArtifactUuid FilesArtifactUUID) (FilesArtifactUUID, *EnclaveDataDirFile, []byte, bool, error) {
	fileContentHash, found := store.fileArtifactDb.GetContentMd5(string(filesArtifactUuid))
	if !found {
		return "", nil, nil, false, stacktrace.NewError("Could not retrieve file with ID '%s' as it doesn't exist in the store", filesArtifactUuid)
	}
	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)
	enclaveDataDirFile, err := store.fileCache.GetFile(filename)
	if err != nil {
		return "", nil, nil, false, stacktrace.Propagate(err, "Could not retrieve file with filename '%s' from the file cache", filename)
	}
	return filesArtifactUuid, enclaveDataDirFile, fileContentHash, true, nil
}

// removeFileUnlocked this is not thread safe, must be used from a thread safe context
func (store FilesArtifactStore) removeFileUnlocked(filesArtifactUuid FilesArtifactUUID) (string, error) {
	filename := strings.Join(
		[]string{string(filesArtifactUuid), artifactExtension},
		".",
	)

	var artifactName string
	if err := store.fileCache.RemoveFile(filename); err != nil {
		return "", stacktrace.Propagate(err, "There was an error in removing '%v' from the file store", filename)
	}
	for name, artifactUuid := range store.fileArtifactDb.GetArtifactUuidMap() {
		if artifactUuid == string(filesArtifactUuid) {
			artifactName = name
			store.fileArtifactDb.DeleteArtifactUuid(name)
			store.fileArtifactDb.DeleteContentMd5(artifactUuid)
		}
	}
	shortenedUuid := uuid_generator.ShortenedUUIDString(string(filesArtifactUuid))
	artifactUuids, found := store.fileArtifactDb.GetFullUuid(shortenedUuid)
	if found {
		var targetArtifactIdx int
		for index, artifactUuid := range artifactUuids {
			if artifactUuid == string(filesArtifactUuid) {
				targetArtifactIdx = index
			}
		}
		// if there's only one matching uuid we delete the shortened uuid
		if len(artifactUuids) == 1 {
			store.fileArtifactDb.DeleteFullUuid(shortenedUuid)
			return artifactName, nil
		}
		// otherwise we just delete the target artifact uuid
		store.fileArtifactDb.DeleteFullUuidIndex(shortenedUuid, targetArtifactIdx)
	}

	return artifactName, nil
}
