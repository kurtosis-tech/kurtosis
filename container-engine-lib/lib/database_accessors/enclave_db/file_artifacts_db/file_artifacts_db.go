package file_artifacts_db

import (
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

type FileArtifactPersisted struct {
	enclaveDb *enclave_db.EnclaveDB
	data      *fileArtifactData
}

type fileArtifactData struct {
	ArtifactNameToArtifactUuid map[string]string
	ShortenedUuidToFullUuid    map[string][]string
	ArtifactContentMd5         map[string][]byte
}

var (
	fileArtifactBucketName    = []byte("file-artifact")
	fileArtifactDataStructKey = []byte("file-artifact-data-struct")
)

func (fileArtifactDb *FileArtifactPersisted) ListFiles() map[string]bool {
	artifactNameSet := make(map[string]bool)
	for artifactName := range fileArtifactDb.data.ArtifactNameToArtifactUuid {
		artifactNameSet[artifactName] = true
	}
	return artifactNameSet
}

func (fileArtifactDb *FileArtifactPersisted) Persist() error {
	err := fileArtifactDb.enclaveDb.Update(func(tx *bolt.Tx) error {
		jsonData, err := json.Marshal(fileArtifactDb.data)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling data '%v'", fileArtifactDb.data)
		}
		return tx.Bucket(fileArtifactBucketName).Put(fileArtifactDataStructKey, jsonData)
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred during file artifact db update")
	}
	return nil
}

func (fileArtifactDb *FileArtifactPersisted) SetArtifactUuid(artifactName string, artifactUuid string) {
	fileArtifactDb.data.ArtifactNameToArtifactUuid[artifactName] = artifactUuid
}

func (fileArtifactDb *FileArtifactPersisted) GetArtifactUuid(artifactName string) (string, bool) {
	value, found := fileArtifactDb.data.ArtifactNameToArtifactUuid[artifactName]
	return value, found
}

func (fileArtifactDb *FileArtifactPersisted) GetArtifactUuidMap() map[string]string {
	return fileArtifactDb.data.ArtifactNameToArtifactUuid
}

func (fileArtifactDb *FileArtifactPersisted) DeleteArtifactUuid(artifactName string) {
	delete(fileArtifactDb.data.ArtifactNameToArtifactUuid, artifactName)
}

func (fileArtifactDb *FileArtifactPersisted) SetFullUuid(shortenedUuid string, fullUuid []string) {
	fileArtifactDb.data.ShortenedUuidToFullUuid[shortenedUuid] = fullUuid
}

func (fileArtifactDb *FileArtifactPersisted) GetFullUuid(shortenedUuid string) ([]string, bool) {
	value, found := fileArtifactDb.data.ShortenedUuidToFullUuid[shortenedUuid]
	return value, found
}

func (fileArtifactDb *FileArtifactPersisted) GetFullUuidMap() map[string][]string {
	return fileArtifactDb.data.ShortenedUuidToFullUuid
}

func (fileArtifactDb *FileArtifactPersisted) DeleteFullUuid(shortenedUuid string) {
	delete(fileArtifactDb.data.ShortenedUuidToFullUuid, shortenedUuid)
}

func (fileArtifactDb *FileArtifactPersisted) DeleteFullUuidIndex(shortenedUuid string, targetArtifactIdx int) {
	artifactUuids := fileArtifactDb.data.ShortenedUuidToFullUuid[shortenedUuid]
	fileArtifactDb.data.ShortenedUuidToFullUuid[shortenedUuid] = append(artifactUuids[0:targetArtifactIdx], artifactUuids[targetArtifactIdx+1:]...)
}

func (fileArtifactDb *FileArtifactPersisted) SetContentMd5(artifactName string, md5 []byte) {
	fileArtifactDb.data.ArtifactContentMd5[artifactName] = md5
}

func (fileArtifactDb *FileArtifactPersisted) GetContentMd5(artifactName string) ([]byte, bool) {
	value, found := fileArtifactDb.data.ArtifactContentMd5[artifactName]
	return value, found
}

func (fileArtifactDb *FileArtifactPersisted) GetContentMd5Map() map[string][]byte {
	return fileArtifactDb.data.ArtifactContentMd5
}

func (fileArtifactDb *FileArtifactPersisted) DeleteContentMd5(artifactName string) {
	delete(fileArtifactDb.data.ArtifactContentMd5, artifactName)
}

func GetOrCreateNewFileArtifactsDb() (*FileArtifactPersisted, error) {
	data := fileArtifactData{
		map[string]string{},
		map[string][]string{},
		map[string][]byte{},
	}
	// using the noEnclaveDatabaseDirpath because at this point we know that the enclave database has been created, so we are getting it from this call
	noEnclaveDatabaseDirpath := ""
	db, err := enclave_db.GetOrCreateEnclaveDatabase(noEnclaveDatabaseDirpath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to get enclave database")
	}
	fileArtifactPersisted, err := getFileArtifactsDbFromEnclaveDb(db, &data)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to hydrate pre-existing file artifacts")
	}
	return fileArtifactPersisted, nil
}

func getFileArtifactsDbFromEnclaveDb(db *enclave_db.EnclaveDB, data *fileArtifactData) (*FileArtifactPersisted, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, bucketErr := tx.CreateBucket(fileArtifactBucketName)
		// New bucket was created
		if errors.Is(bucketErr, bolt.ErrBucketExists) {
			content := tx.Bucket(fileArtifactBucketName).Get(fileArtifactDataStructKey)
			if err := json.Unmarshal(content, data); err != nil {
				return stacktrace.Propagate(err, "An error occurred restoring previous file artifact db state from '%v'", content)
			}
			return nil
		}
		// No need to create new bucket
		if bucketErr == nil {
			if err := bucket.Put(fileArtifactDataStructKey, consts.EmptyValueForJsonSet); err != nil {
				return stacktrace.Propagate(err, "An error occurred while creating updating artifact bucket o '%v'", consts.EmptyValueForJsonSet)
			}
			return nil
		}
		// Unexpected error
		return stacktrace.Propagate(bucketErr, "An error occurred while creating file artifact bucket")
	})
	if err == nil {
		return &FileArtifactPersisted{
			db,
			data,
		}, nil
	}
	return nil, stacktrace.Propagate(err, "An error occurred while getting file artifact db")
}

func GetFileArtifactsDbForTesting(db *enclave_db.EnclaveDB, nameToUuid map[string]string) (*FileArtifactPersisted, error) {
	fileArtifactPersisted, err := getFileArtifactsDbFromEnclaveDb(db, &fileArtifactData{
		nameToUuid,
		map[string][]string{},
		map[string][]byte{},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to hydrate pre-existing file artifacts")
	}
	return fileArtifactPersisted, nil
}
