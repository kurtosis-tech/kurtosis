package partition_topology_db

import (
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

func (sp *ServicePartitionsBucket) AddPartitionToService(serviceName service.ServiceName, partitionId partition.PartitionID) error {
	err := sp.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(servicePartitionsBucketName).Put([]byte(serviceName), []byte(partitionId))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while adding partition '%v' for service '%v'", partitionId, serviceName)
		}
		return nil
	})
	return err
}

func (sp *ServicePartitionsBucket) DoesServiceExist(serviceName service.ServiceName) (bool, error) {
	partitionExists := false
	err := sp.db.View(func(tx *bolt.Tx) error {
		values := tx.Bucket(servicePartitionsBucketName).Get([]byte(serviceName))
		if values != nil {
			partitionExists = true
		}
		return nil
	})
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred while fetching information from the underlying bucket")
	}
	return partitionExists, nil
}

func (sp *ServicePartitionsBucket) GetPartitionForService(service service.ServiceName) (partition.PartitionID, error) {
	var partitionForService partition.PartitionID
	err := sp.db.View(func(tx *bolt.Tx) error {
		values := tx.Bucket(servicePartitionsBucketName).Get([]byte(service))
		if values == nil {
			return nil
		}
		partitionForService = partition.PartitionID(values)
		return nil
	})
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while getting all partition for service '%v'", service)
	}
	return partitionForService, nil
}

func (sp *ServicePartitionsBucket) RepartitionBucket(newPartitioning map[service.ServiceName]partition.PartitionID) error {
	err := sp.db.Update(func(tx *bolt.Tx) error {
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
				return stacktrace.Propagate(err, "An error occurred while storing partition for service '%v'", serviceName)
			}
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while repartitioning the bucket")
	}
	return nil
}

func (sp *ServicePartitionsBucket) RemoveService(serviceName service.ServiceName) error {
	err := sp.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket(servicePartitionsBucketName).Delete([]byte(serviceName))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while removing '%v' from store", serviceName)
		}
		return nil
	})
	return err
}

func (sp *ServicePartitionsBucket) GetAllServicePartitions() (map[service.ServiceName]partition.PartitionID, error) {
	result := map[service.ServiceName]partition.PartitionID{}
	err := sp.db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(servicePartitionsBucketName).ForEach(func(k, v []byte) error {
			partitionForService := partition.PartitionID(v)
			result[service.ServiceName(k)] = partitionForService
			return nil
		})
		return err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all services & associated partitions")
	}
	return result, nil
}

func GetOrCreateServicePartitionsBucket(db *enclave_db.EnclaveDB) (*ServicePartitionsBucket, error) {
	bucketExists := false
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(servicePartitionsBucketName)
		if err != nil {
			bucketExists = true
			return stacktrace.Propagate(err, "An error occurred while creating services partitions database bucket")
		}
		return nil
	})
	if err != nil && !bucketExists {
		return nil, stacktrace.Propagate(err, "An error occurred while building service partitions")
	}
	// Bucket does exist, skipping population step
	return &ServicePartitionsBucket{
		db,
	}, nil
}
