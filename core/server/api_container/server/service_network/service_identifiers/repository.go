package service_identifiers

import (
	"encoding/json"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/consts"
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	serviceIdentifiersBucketName = []byte("service-identifiers-repository")
	serviceIdentifiersSliceKey   = []byte("service-identifiers-slice")
)

type ServiceIdentifiersRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func GetOrCreateNewServiceIdentifiersRepository(enclaveDb *enclave_db.EnclaveDB) (*ServiceIdentifiersRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(serviceIdentifiersBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the service identifiers database bucket")
		}
		logrus.Debugf("Service identifiers bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the service identifiers repository")
	}

	serviceIdentifiersRepository := &ServiceIdentifiersRepository{
		enclaveDb: enclaveDb,
	}

	return serviceIdentifiersRepository, nil
}

func (repository *ServiceIdentifiersRepository) AddServiceIdentifier(
	serviceIdentifierObj *serviceIdentifier,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceIdentifiersBucketName)

		// retrieve the list from the bucket
		serviceIdentifiers, err := getServiceIdentifiersFromBucket(bucket)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service identifiers from bucket with name '%s'", serviceIdentifiersBucketName)
		}

		// add the new element
		serviceIdentifiers = append(serviceIdentifiers, serviceIdentifierObj)
		serviceIdentifiersPtr := &serviceIdentifiers
		jsonBytes, err := json.Marshal(serviceIdentifiersPtr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling service identifiers '%+v'", serviceIdentifiers)
		}

		// save it to disk
		if err := bucket.Put(serviceIdentifiersSliceKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving service identifiers '%+v' into the enclave db", serviceIdentifiers)
		}
		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding service identifier '%v' into the enclave db", serviceIdentifierObj)
	}
	return nil
}

func (repository *ServiceIdentifiersRepository) GetServiceIdentifiers() (ServiceIdentifiers, error) {
	var (
		serviceIdentifiers = []*serviceIdentifier{}
		err                error
	)

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(serviceIdentifiersBucketName)

		serviceIdentifiers, err = getServiceIdentifiersFromBucket(bucket)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting service identifiers from bucket with name '%s'", serviceIdentifiersBucketName)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting service identifiers from the enclave db")
	}
	return serviceIdentifiers, nil
}

func getServiceIdentifiersFromBucket(bucket *bolt.Bucket) ([]*serviceIdentifier, error) {

	// first get the list
	serviceIdentifiersBytes := bucket.Get(serviceIdentifiersSliceKey)
	// for empty list case
	if serviceIdentifiersBytes == nil {
		serviceIdentifiersBytes = consts.EmptyValueForJsonList
	}
	serviceIdentifiers := []*serviceIdentifier{}
	serviceIdentifiersPtr := &serviceIdentifiers

	if err := json.Unmarshal(serviceIdentifiersBytes, serviceIdentifiersPtr); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling service identifiers")
	}

	return serviceIdentifiers, nil
}
