package service_partitions

import (
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

var (
	servicePartitionsBucketName = []byte("service-partitions")
)

type ServicePartitionsBucket struct {
	db *enclave_db.EnclaveDB
}

func newServicePartitions(db *enclave_db.EnclaveDB) *ServicePartitionsBucket {
	return &ServicePartitionsBucket{
		db,
	}
}

func (sp *ServicePartitionsBucket) AddPartitionToService(serviceName service.ServiceName, partitionId partition.PartitionID) error {
	addPartitionToServiceFunc := func(tx *bolt.Tx) error {
		return tx.Bucket(servicePartitionsBucketName).Put([]byte(serviceName), []byte(partitionId))
	}
	if err := sp.db.Update(addPartitionToServiceFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding partition '%v' for service '%v'", partitionId, serviceName)
	}
	return nil
}

func (sp *ServicePartitionsBucket) DoesServiceExist(serviceName service.ServiceName) (bool, error) {
	partitionExists := false
	doesPartitionExistFunc := func(tx *bolt.Tx) error {
		values := tx.Bucket(servicePartitionsBucketName).Get([]byte(serviceName))
		if values != nil {
			partitionExists = true
		}
		return nil
	}
	if err := sp.db.View(doesPartitionExistFunc); err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while verifying whether the '%s' service exists in the '%v' bucket", serviceName, servicePartitionsBucketName)
	}
	return partitionExists, nil
}

func (sp *ServicePartitionsBucket) GetPartitionForService(service service.ServiceName) (partition.PartitionID, error) {
	var partitionForService partition.PartitionID
	getPartitionForServiceFunc := func(tx *bolt.Tx) error {
		values := tx.Bucket(servicePartitionsBucketName).Get([]byte(service))
		if values == nil {
			return nil
		}
		partitionForService = partition.PartitionID(values)
		return nil
	}
	if err := sp.db.View(getPartitionForServiceFunc); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting all partition for service '%v'", service)
	}
	return partitionForService, nil
}

func (sp *ServicePartitionsBucket) ReplaceBucketContents(newPartitioning map[service.ServiceName]partition.PartitionID) error {
	deleteAndReplaceBucketFunc := func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(servicePartitionsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting the bucket")
		}
		bucket, err := tx.CreateBucket(servicePartitionsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred recreating the bucket")
		}
		for serviceName, partitionForService := range newPartitioning {
			err = bucket.Put([]byte(serviceName), []byte(partitionForService))
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while storing partition '%v' for service '%v'", serviceName, partitionForService)
			}
		}
		return nil
	}
	if err := sp.db.Update(deleteAndReplaceBucketFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while replacing the existing service partition configuration with new configuration")
	}
	return nil
}

func (sp *ServicePartitionsBucket) RemoveService(serviceName service.ServiceName) error {
	removeServiceFromBucketFunc := func(tx *bolt.Tx) error {
		return tx.Bucket(servicePartitionsBucketName).Delete([]byte(serviceName))
	}
	if err := sp.db.Update(removeServiceFromBucketFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing '%v' from bucket", serviceName)
	}
	return nil
}

func (sp *ServicePartitionsBucket) GetAllServicePartitions() (map[service.ServiceName]partition.PartitionID, error) {
	result := map[service.ServiceName]partition.PartitionID{}
	getAllServicePartitionsFunc := func(tx *bolt.Tx) error {
		iterateThroughBucketAndPopulateResult := func(serviceName, partitionId []byte) error {
			partitionForService := partition.PartitionID(partitionId)
			if oldPartitionForService, found := result[service.ServiceName(serviceName)]; found {
				return stacktrace.NewError("The service '%s' has more than one mappings, found mapping for partition '%v' & '%v'; This should never happen this is a bug in Kurtosis", serviceName, oldPartitionForService, partitionForService)
			}
			result[service.ServiceName(serviceName)] = partitionForService
			return nil
		}
		return tx.Bucket(servicePartitionsBucketName).ForEach(iterateThroughBucketAndPopulateResult)
	}
	if err := sp.db.View(getAllServicePartitionsFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all services & associated partitions")
	}
	return result, nil
}

func GetOrCreateServicePartitionsBucket(db *enclave_db.EnclaveDB) (*ServicePartitionsBucket, error) {
	createOrReplaceBucketFunc := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(servicePartitionsBucketName)
		if err != nil && !errors.Is(err, bolt.ErrBucketExists) {
			return stacktrace.Propagate(err, "An error occurred while creating services partitions database bucket")
		}
		return nil
	}
	if err := db.Update(createOrReplaceBucketFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building service partitions")
	}
	// Bucket does exist, skipping population step
	return newServicePartitions(db), nil
}
