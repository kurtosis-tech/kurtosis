package services_files_artifacts

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	servicesFilesArtifactsBucketName = []byte("services-files-artifacts-repository")
	servicesFilesArtifactsSliceKey   = []byte("service-files-artifacts-slice")
)

type ServicesFilesArtifactsRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func GetOrCreateNewServicesFilesArtifactsRepository(enclaveDb *enclave_db.EnclaveDB) (*ServicesFilesArtifactsRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(servicesFilesArtifactsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the services files artifacts identifiers database bucket")
		}
		logrus.Debugf("Services files artifacts bucket identifier: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the services files artifacts repository")
	}

	servicesFilesArtifactsRepository := &ServicesFilesArtifactsRepository{
		enclaveDb: enclaveDb,
	}

	return servicesFilesArtifactsRepository, nil
}

func (repository *ServicesFilesArtifactsRepository) AddServicesFilesArtifacts(
	servicesFilesArtifactsObj *servicesFilesArtifacts,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(servicesFilesArtifactsBucketName)

		// retrieve the list from the bucket
		servicesFilesArtifacts, err := getServicesFilesArtifactsFromBucket(bucket)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service identifiers from bucket with name '%s'", servicesFilesArtifactsBucketName)
		}

		// add the new element
		servicesFilesArtifacts = append(servicesFilesArtifacts, servicesFilesArtifactsObj)
		servicesFilesArtifactsPtr := &servicesFilesArtifacts
		jsonBytes, err := json.Marshal(servicesFilesArtifactsPtr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling service identifiers '%+v'", servicesFilesArtifacts)
		}

		// save it to disk
		if err := bucket.Put(servicesFilesArtifactsSliceKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving service identifiers '%+v' into the enclave db", servicesFilesArtifacts)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding service identifier '%v' into the enclave db", servicesFilesArtifactsObj)
	}
	return nil
}

func getServicesFilesArtifactsFromBucket(bucket *bolt.Bucket) ([]*servicesFilesArtifacts, error) {

	// first get the list
	servicesFilesArtifactsBytes := bucket.Get(servicesFilesArtifactsSliceKey)
	// for empty list case
	if servicesFilesArtifactsBytes == nil {
		servicesFilesArtifactsBytes = consts.EmptyValueForJsonList
	}
	servicesFilesArtifacts := []*servicesFilesArtifacts{}
	servicesFilesArtifactsPtr := &servicesFilesArtifacts

	if err := json.Unmarshal(servicesFilesArtifactsBytes, servicesFilesArtifactsPtr); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling service identifiers")
	}

	return servicesFilesArtifacts, nil
}
