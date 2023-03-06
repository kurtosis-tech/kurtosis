package partition_connection_overrides

import (
	"encoding/json"
	"errors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	bolt "go.etcd.io/bbolt"
)

type PartitionConnectionOverridesBucket struct {
	db *enclave_db.EnclaveDB
}

var (
	partitionConnectionOverridesBucketName = []byte("partition-connection-overrides")
)

func newPartitionConnectionOverridesBucket(db *enclave_db.EnclaveDB) *PartitionConnectionOverridesBucket {
	return &PartitionConnectionOverridesBucket{
		db: db,
	}
}

func (pc *PartitionConnectionOverridesBucket) GetPartitionConnectionOverride(connectionId PartitionConnectionID) (PartitionConnection, error) {
	var connection PartitionConnection
	getPartitionConnectionOverride := func(tx *bolt.Tx) error {
		jsonifiedPartitionConnectionId, err := json.Marshal(connectionId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection ID '%v' to json", connectionId)
		}
		values := tx.Bucket(partitionConnectionOverridesBucketName).Get(jsonifiedPartitionConnectionId)
		if values == nil {
			return nil
		}
		if err = json.Unmarshal(values, &connection); err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection '%v' from json bytes to Golang type; This is a bug in Kurtosis", values)
		}
		return nil
	}
	if err := pc.db.View(getPartitionConnectionOverride); err != nil {
		return EmptyPartitionConnection, stacktrace.Propagate(err, "An error occurred while fetching partition connection override for connection with ID '%v'", connectionId)
	}
	return connection, nil
}

func (pc *PartitionConnectionOverridesBucket) DoesPartitionConnectionOverrideExist(connectionId PartitionConnectionID) (bool, error) {
	var exists bool
	getPartitionConnection := func(tx *bolt.Tx) error {
		jsonifiedPartitionConnectionId, err := json.Marshal(connectionId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection with ID '%v' to json", connectionId)
		}
		if values := tx.Bucket(partitionConnectionOverridesBucketName).Get(jsonifiedPartitionConnectionId); values == nil {
			return nil
		}
		exists = true
		return nil
	}
	if err := pc.db.View(getPartitionConnection); err != nil {
		return exists, stacktrace.Propagate(err, "An error occurred while verifying whether connection override with ID exists '%v'", connectionId)
	}
	return exists, nil
}

func (pc *PartitionConnectionOverridesBucket) AddPartitionConnectionOverride(connectionId PartitionConnectionID, connection PartitionConnection) error {
	addPartitionToServiceFunc := func(tx *bolt.Tx) error {
		jsonifiedConnectionId, err := json.Marshal(connectionId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection ID '%v' to json", connectionId)
		}
		jsonifiedPartitionConnection, err := json.Marshal(connection)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection '%v' to json", connection)
		}
		return tx.Bucket(partitionConnectionOverridesBucketName).Put(jsonifiedConnectionId, jsonifiedPartitionConnection)
	}
	if err := pc.db.Update(addPartitionToServiceFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while adding partition connection override '%v' with ID '%v' to bucket", connection, connectionId)
	}
	return nil
}

func (pc *PartitionConnectionOverridesBucket) GetAllPartitionConnectionOverrides() (map[PartitionConnectionID]PartitionConnection, error) {
	result := map[PartitionConnectionID]PartitionConnection{}
	getAllServicePartitionsFunc := func(tx *bolt.Tx) error {
		iterateThroughBucketAndPopulateResult := func(connectionIdBytes, connectionBytes []byte) error {
			var connectionIdUnmarshalled PartitionConnectionID
			if err := json.Unmarshal(connectionIdBytes, &connectionIdUnmarshalled); err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting partition connection ID in bucket '%v' to Golang Type; this is a bug in Kurtosis", connectionIdBytes)
			}
			var connectionUnmarshalled PartitionConnection
			if err := json.Unmarshal(connectionBytes, &connectionUnmarshalled); err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting connection override in bucket '%v' to Golang Type' this is a bug in Kurtosis", connectionBytes)
			}
			result[connectionIdUnmarshalled] = connectionUnmarshalled
			return nil
		}
		return tx.Bucket(partitionConnectionOverridesBucketName).ForEach(iterateThroughBucketAndPopulateResult)
	}
	if err := pc.db.View(getAllServicePartitionsFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all services & associated partitions")
	}
	return result, nil
}

func (pc *PartitionConnectionOverridesBucket) ReplaceBucketContents(newConnections map[PartitionConnectionID]PartitionConnection) error {
	deleteAndReplaceBucketFunc := func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket(partitionConnectionOverridesBucketName); err != nil {
			return stacktrace.Propagate(err, "An error occurred deleting the bucket")
		}
		bucket, err := tx.CreateBucket(partitionConnectionOverridesBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred recreating the bucket")
		}
		for partitionConnectionId, connection := range newConnections {
			jsonifiedConnectionId, err := json.Marshal(partitionConnectionId)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting partition connection ID '%v' to json", partitionConnectionId)
			}
			jsonifiedPartitionConnection, err := json.Marshal(connection)
			if err != nil {
				return stacktrace.Propagate(err, "An error occurred while converting partition connection '%v' to json", connection)
			}
			if err = bucket.Put(jsonifiedConnectionId, jsonifiedPartitionConnection); err != nil {
				return stacktrace.Propagate(err, "An error occurred while storing connection override with connection ID '%v' and values '%v' to bucket", jsonifiedConnectionId, jsonifiedPartitionConnection)
			}
		}
		return nil
	}
	if err := pc.db.Update(deleteAndReplaceBucketFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while replacing the existing partition connection bucket with new contents")
	}
	return nil
}

func (pc *PartitionConnectionOverridesBucket) RemovePartitionConnectionOverride(connectionId PartitionConnectionID) error {
	removeServiceFromBucketFunc := func(tx *bolt.Tx) error {
		jsonifiedConnectionId, err := json.Marshal(connectionId)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while converting partition connection ID '%v' to json", connectionId)
		}
		return tx.Bucket(partitionConnectionOverridesBucketName).Delete(jsonifiedConnectionId)
	}
	if err := pc.db.Update(removeServiceFromBucketFunc); err != nil {
		return stacktrace.Propagate(err, "An error occurred while removing partition connection override with ID '%v' from bucket", connectionId)
	}
	return nil
}

func GetOrCreatePartitionConnectionOverrideBucket(db *enclave_db.EnclaveDB) (*PartitionConnectionOverridesBucket, error) {
	createOrReplaceBucketFunc := func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket(partitionConnectionOverridesBucketName)
		if err != nil && !errors.Is(err, bolt.ErrBucketExists) {
			return stacktrace.Propagate(err, "An error occurred while creating partition connection override bucket")
		}
		return nil
	}
	if err := db.Update(createOrReplaceBucketFunc); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building partition connection override")
	}

	return newPartitionConnectionOverridesBucket(
		db,
	), nil
}
