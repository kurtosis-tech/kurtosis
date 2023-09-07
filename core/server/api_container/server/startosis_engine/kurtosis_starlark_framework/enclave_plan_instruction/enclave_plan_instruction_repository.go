package enclave_plan_instruction

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var (
	enclavePlanInstructionBucketName = []byte("enclave-plan-instruction-repository")
)

type EnclavePlanInstructionRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func GetOrCreateNewEnclavePlanInstructionRepository(enclaveDb *enclave_db.EnclaveDB) (*EnclavePlanInstructionRepository, error) {
	if err := enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(enclavePlanInstructionBucketName)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred while creating the enclave plan instruction database bucket")
		}
		logrus.Debugf("Enclave plan instruction bucket: '%+v'", bucket)

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while building the enclave plan instruction repository")
	}

	repository := &EnclavePlanInstructionRepository{
		enclaveDb: enclaveDb,
	}

	return repository, nil
}

func (repository *EnclavePlanInstructionRepository) Save(
	uuid instructions_plan.ScheduledInstructionUuid,
	instruction *EnclavePlanInstructionImpl,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		uuidKey := getUuidKey(uuid)

		jsonBytes, err := json.Marshal(instruction)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling enclave plan instruction '%+v' in the enclave plan instruction repository", instruction)
		}

		// save it to disk
		if err := bucket.Put(uuidKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' with UUID '%s' into the enclave db bucket", instruction, uuid)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' with UUID '%s' into the enclave db", instruction, uuid)
	}
	return nil
}

func (repository *EnclavePlanInstructionRepository) Executed(
	uuid instructions_plan.ScheduledInstructionUuid,
	isExecuted bool,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		instruction, err := get(tx, uuid)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclave instruction plan with UUID '%v'", uuid)
		}

		if instruction == nil {
			return stacktrace.Propagate(err, "Imposible to set if the enclave instruction plan with UUID '%v' was executed because it doesn't exist in the repository", uuid)
		}

		instruction.Executed(isExecuted)

		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		uuidKey := getUuidKey(uuid)

		jsonBytes, err := json.Marshal(instruction)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling enclave plan instruction '%+v' in the enclave plan instruction repository", instruction)
		}

		// save it to disk
		if err := bucket.Put(uuidKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' with UUID '%s' into the enclave db bucket", instruction, uuid)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while setting executed field to '%v' for enclave plan instruction with UUID '%s' into the enclave db", isExecuted, uuid)
	}
	return nil
}

// Get returns an instruction with UUID if exist or nil if it doesn't
func (repository *EnclavePlanInstructionRepository) Get(
	uuid instructions_plan.ScheduledInstructionUuid,
) (*EnclavePlanInstructionImpl, error) {
	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	instruction := &EnclavePlanInstructionImpl{}
	var err error

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {

		instruction, err = get(tx, uuid)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclave instruction plan with UUID '%v'", uuid)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the enclave plan instruction with UUID '%s' from the enclave db", uuid)
	}

	return instruction, nil
}

func get(
	tx *bolt.Tx,
	uuid instructions_plan.ScheduledInstructionUuid,
) (*EnclavePlanInstructionImpl, error) {
	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	instruction := &EnclavePlanInstructionImpl{}

	bucket := tx.Bucket(enclavePlanInstructionBucketName)

	bucket.Sequence()

	uuidKey := getUuidKey(uuid)

	// first get the bytes
	jsonBytes := bucket.Get(uuidKey)

	if jsonBytes == nil {
		return nil, nil
	}

	if err := json.Unmarshal(jsonBytes, &instruction); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling the enclave plan instruction with UUID '%s' from the repository", uuid)
	}

	return instruction, nil
}

func getUuidKey(uuid instructions_plan.ScheduledInstructionUuid) []byte {
	return []byte(uuid)
}
