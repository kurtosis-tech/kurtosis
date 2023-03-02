package partition_topology_db

import (
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/partition"
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
	lexicalFirst  partition.PartitionID
	lexicalSecond partition.PartitionID
}

func (pc *PartitionConnectionID) String() string {
	return string(pc.lexicalFirst + pc.lexicalSecond)
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
