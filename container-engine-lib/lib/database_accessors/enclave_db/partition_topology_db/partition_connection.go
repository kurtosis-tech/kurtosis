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

// remove
// get

func (sp *ServicePartitionsBucket) AddPartitionConnection(connectionId PartitionConnectionID, connection PartitionConnection) error {
	addPartitionToServiceFunc := func(tx *bolt.Tx) error {
		jsonifiedPartitionConnection, err := json.Marshal(connection)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection '%v' to json", connection)
		}
		jsonifiedConnectionId, err := json.Marshal(connectionId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection id '%v' to json", connectionId)
		}
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting connection to internal type")
		}
		return tx.Bucket(servicePartitionsBucketName).Put(jsonifiedPartitionConnection, jsonifiedConnectionId)
	}
	if err := sp.db.Update(addPartitionToServiceFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding partition connection '%v' with id '%v'", connection, connectionId)
	}
	return nil
}

func (pc *PartitionConnectionBucket) GetAllPartitionConnections() (map[PartitionConnectionID]PartitionConnection, error) {
	result := map[PartitionConnectionID]PartitionConnection{}
	getAllServicePartitionsFunc := func(tx *bolt.Tx) error {
		iterateThroughBucketAndPopulateResult := func(connectionId, connection []byte) error {
			var connectionIdUnmarshalled PartitionConnectionID
			err := json.Unmarshal(connectionId, &connectionIdUnmarshalled)
			// TODO rework these errors before PR
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting connection id to internal type")
			}
			var connectionUnmarshalled PartitionConnection
			err = json.Unmarshal(connection, &connectionUnmarshalled)
			// TODO rework these errors before PR
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting connection to internal type")
			}
			result[connectionIdUnmarshalled] = connectionUnmarshalled
			return nil
		}
		return tx.Bucket(servicePartitionsBucketName).ForEach(iterateThroughBucketAndPopulateResult)
	}
	if err := pc.db.View(getAllServicePartitionsFunc); err != nil {
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
			jsonifiedConnectionId, err := json.Marshal(partitionConnectionId)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting partition connection id '%v' to json", partitionConnectionId)
			}
			if err = bucket.Put(jsonifiedConnectionId, jsonifiedPartitionConnection); err != nil {
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
