package partition_services

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

var (
	partitionServicesBucketName = []byte("partition-services")
)

type PartitionServicesBucket struct {
	db *enclave_db.EnclaveDB
}

func (ps *PartitionServicesBucket) DoesPartitionExist(partitionId partition.PartitionID) (bool, error) {
	partitionExists := false
	err := ps.db.View(func(tx *bolt.Tx) error {
		values := tx.Bucket(partitionServicesBucketName).Get([]byte(partitionId))
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

func (ps *PartitionServicesBucket) GetServicesForPartition(partitionId partition.PartitionID) (map[service.ServiceName]bool, error) {
	result := map[service.ServiceName]bool{}
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

func (ps *PartitionServicesBucket) GetAllPartitions() (map[partition.PartitionID]map[service.ServiceName]bool, error) {
	result := map[partition.PartitionID]map[service.ServiceName]bool{}
	err := ps.db.View(func(tx *bolt.Tx) error {
		err := tx.Bucket(partitionServicesBucketName).ForEach(func(k, v []byte) error {
			servicesForPartition := map[service.ServiceName]bool{}
			err := json.Unmarshal(v, &servicesForPartition)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting the services stored '%s' against partition '%s' in bolt to a usable Go type; This is a bug in Kurtosis", v, k)
			}
			result[partition.PartitionID(k)] = servicesForPartition
			return nil
		})
		return err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all partitions & associated services")
	}
	return result, nil
}

func (ps *PartitionServicesBucket) BulkAddPartitionServices(partitionServices map[partition.PartitionID]map[service.ServiceName]bool) error {
	err := ps.db.Update(func(tx *bolt.Tx) error {
		for partitionId, servicesForPartition := range partitionServices {
			servicesForPartitionBytes, err := json.Marshal(servicesForPartition)
			if err != nil {
				return stacktrace.Propagate(err, "An unexpected error occurred converting a Go type to bytes services for partition '%s'; This is a bug in Kurtosis", partitionId)
			}
			return tx.Bucket(partitionServicesBucketName).Put([]byte(partitionId), servicesForPartitionBytes)
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while bulk adding partition services '%v'", partitionServices)
	}
	return nil
}

func (ps *PartitionServicesBucket) AddServicesToPartition(partitionId partition.PartitionID, services map[service.ServiceName]bool) error {
	partitionServices := map[partition.PartitionID]map[service.ServiceName]bool{
		partitionId: services,
	}
	err := ps.BulkAddPartitionServices(partitionServices)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding services for partition '%s'", partitionId)
	}
	return nil
}

func (ps *PartitionServicesBucket) DeletePartition(partitionId partition.PartitionID) error {
	err := ps.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(partitionServicesBucketName).Delete([]byte(partitionId))
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while deleting services for partition '%s'", partitionId)
	}
	return nil
}

func (ps *PartitionServicesBucket) AddServiceToPartition(partitionId partition.PartitionID, serviceName service.ServiceName) error {
	err := ps.db.Update(func(tx *bolt.Tx) error {
		servicesForPartition := map[service.ServiceName]bool{}
		values := tx.Bucket(partitionServicesBucketName).Get([]byte(partitionId))
		err := json.Unmarshal(values, &servicesForPartition)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting services stored against partition '%v' to a Go type; This is a bug in Kurtosis.", partitionId)
		}
		servicesForPartition[serviceName] = true
		servicesForPartitionAsBytes, err := json.Marshal(servicesForPartition)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting servicesForPartition to json. This is a bug in Kurtosis.")
		}
		return tx.Bucket(partitionServicesBucketName).Put([]byte(partitionId), servicesForPartitionAsBytes)
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding service '%s' to partition '%s'", serviceName, partitionId)
	}
	return nil
}

func (ps *PartitionServicesBucket) RemoveServiceFromPartition(serviceName service.ServiceName, partitionId partition.PartitionID) error {
	err := ps.db.Update(func(tx *bolt.Tx) error {
		servicesForPartition := map[service.ServiceName]bool{}
		values := tx.Bucket(partitionServicesBucketName).Get([]byte(partitionId))
		err := json.Unmarshal(values, &servicesForPartition)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting services stored against partition '%v' to a Go type; This is a bug in Kurtosis", partitionId)
		}
		delete(servicesForPartition, serviceName)
		servicesForPartitionAsBytes, err := json.Marshal(servicesForPartition)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting servicesForPartition to json. This is a bug in Kurtosis.")
		}
		return tx.Bucket(partitionServicesBucketName).Put([]byte(partitionId), servicesForPartitionAsBytes)
	})
	return err
}

func (ps *PartitionServicesBucket) RepartitionBucket(newPartitioning map[partition.PartitionID]map[service.ServiceName]bool) error {
	err := ps.db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(partitionServicesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting the bucket")
		}
		bucket, err := tx.CreateBucket(partitionServicesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred recreating the bucket")
		}
		for partitionId, services := range newPartitioning {
			jsonifiedServices, err := json.Marshal(services)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while serializing services for partition '%v' to store in db", partitionId)
			}
			err = bucket.Put([]byte(partitionId), jsonifiedServices)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while storing services for partition '%v'", partitionId)
			}
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while repartitioning the bucket")
	}
	return nil
}

func GetOrCreatePartitionServicesBucket(db *enclave_db.EnclaveDB, defaultPartitionId partition.PartitionID) (*PartitionServicesBucket, error) {
	bucketExists := false
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(partitionServicesBucketName)
		if err != nil {
			bucketExists = true
			return stacktrace.Propagate(err, "An error occurred while creating partition services database bucket")
		}
		return tx.Bucket(partitionServicesBucketName).Put([]byte(defaultPartitionId), consts.EmptyValueForJsonSet)
	})
	if err != nil && !bucketExists {
		return nil, stacktrace.Propagate(err, "An error occurred while building partition services")
	}
	// Bucket does exist, skipping population step
	return &PartitionServicesBucket{
		db,
	}, nil
}
