package partition_topology_db

import (
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

type PartitionConnectionBucket struct {
	db *enclave_db.EnclaveDB
}

var (
	partitionConnectionsBucketName = []byte("partition-connections")
)

func newPartitionConnectionBucket(db *enclave_db.EnclaveDB) *PartitionConnectionBucket {
	return &PartitionConnectionBucket{
		db: db,
	}
}

type PartitionConnectionID struct {
	LexicalFirst  partition.PartitionID `json:"lexical_first"`
	LexicalSecond partition.PartitionID `json:"lexical_second"`
}

type delayDistribution struct {
	AvgDelayMs  uint32  `json:"avg_delay"`
	Jitter      uint32  `json:"jitter"`
	Correlation float32 `json:"correlation"`
}

type PartitionConnection struct {
	PacketLoss              float32           `json:"packet_loss"`
	PacketDelayDistribution delayDistribution `json:"delay_distribution"`
}

// get all
// remove
// add
// get

func (pc *PartitionConnectionBucket) GetAllPartitionConnections() (map[PartitionConnectionID]PartitionConnection, error) {
	result := map[PartitionConnectionID]PartitionConnection{}
	getAllServicePartitionsFunc := func(tx *bolt.Tx) error {
		iterateThroughBucketAndPopulateResult := func(connectionId, connection []byte) error {
			partitionForService := partition.PartitionID(partitionId)
			if oldPartitionForService, found := result[service.ServiceName(serviceName)]; found {
				return stacktrace.NewError("The service '%s' has more than one mappings, found mapping for partition '%v' & '%v'; This should never happen this is a bug in Kurtosis", serviceName, oldPartitionForService, partitionForService)
			}
			result[service.ServiceName(serviceName)] = partitionForService
			return nil
		}
		return tx.Bucket(servicePartitionsBucketName).ForEach(iterateThroughBucketAndPopulateResult)
	}
	if err := sp.db.Update(getAllServicePartitionsFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all services & associated partitions")
	}
	return result, nil
}

func (pc *PartitionConnectionBucket) ReplaceBucketContents(newConnections map[PartitionConnectionID]PartitionConnection) error {
	deleteAndReplaceBucketFunc := func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(partitionConnectionsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting the bucket")
		}
		bucket, err := tx.CreateBucket(partitionConnectionsBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred recreating the bucket")
		}
		for partitionConnectionId, connection := range newConnections {
			jsonifiedPartitionConnection, err := json.Marshal(connection)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting partition connection '%v' to json", connection)
			}
			if err = bucket.Put([]byte(partitionConnectionId.String()), jsonifiedPartitionConnection); err != nil {
				return stacktrace.Propagate(err, "An error occurred while storing connection with id '%v' and values '%v'", partitionConnectionId, jsonifiedPartitionConnection)
			}
		}
		return nil
	}
	if err := pc.db.Update(deleteAndReplaceBucketFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while replacing the existing service partition configuration with new configuration")
	}
	return nil
}

func GetPartitionConnectionBucket(db *enclave_db.EnclaveDB) (*PartitionConnectionBucket, error) {
	createOrReplaceBucketFunc := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(partitionConnectionsBucketName)
		if err != nil && !errors.Is(err, bolt.ErrBucketExists) {
			return stacktrace.Propagate(err, "An error occurred while creating services partitions database bucket")
		}
		return nil
	}
	if err := db.Update(createOrReplaceBucketFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building service partitions")
	}

	return newPartitionConnectionBucket(
		db,
	), nil
}
