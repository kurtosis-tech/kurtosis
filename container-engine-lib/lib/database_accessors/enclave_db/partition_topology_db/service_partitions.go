package partition_topology_db

import (
	"encoding/json"
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

func (ps *ServicePartitionsBucket) DoesServiceExist(serviceName service.ServiceName) (bool, error) {
	partitionExists := false
	err := ps.db.View(func(tx *bolt.Tx) error {
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

func (ps *ServicePartitionsBucket) GetPartitionForService(service service.ServiceName) (map[service.ServiceName]bool, error) {
	result := map[partition.PartitionID]bool{}
	err := ps.db.View(func(tx *bolt.Tx) error {
		values := tx.Bucket(partitionServicesBucketName).Get([]byte(partitionId))
		if values == nil {
			return nil
		}
		err := json.Unmarshal(values, &result)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting the services stored '%s' against partition '%s' in bolt to a usable Go type; This is a bug in Kurtosis", values, partitionId)
		}
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all services for partition '%v'", partitionId)
	}
	return result, nil
}

func (ps *ServicePartitionsBucket) GetAllServices() (map[service.ServiceName]map[partition.PartitionID]bool, error) {
	result := map[service.ServiceName]map[partition.PartitionID]bool{}
	err := ps.db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(servicePartitionsBucketName).ForEach(func(k, v []byte) error {
			partitionForServices := map[partition.PartitionID]bool{}
			err := json.Unmarshal(v, &partitionForServices)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting the partitions stored '%s' against service '%s' in bolt to a usable Go type; This is a bug in Kurtosis", v, k)
			}
			result[service.ServiceName(k)] = partitionForServices
			return nil
		})
		return err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all partitions & associated services")
	}
	return result, nil
}

func GetOrCreateServicePartitionsBucket(db *enclave_db.EnclaveDB) (*ServicePartitionsBucket, error) {
	bucketExists := false
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(servicePartitionsBucketName)
		if err != nil {
			bucketExists = true
			return stacktrace.Propagate(err, "An error occurred while creating partition services database bucket")
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
