package enclave_plan_instruction

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
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

// SaveIfNotExist will save the instruction if it doesn't exist otherwise it will do nothing
func (repository *EnclavePlanInstructionRepository) SaveIfNotExist(
	instruction *EnclavePlanInstructionImpl,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {

		instructionStr := instruction.GetKurtosisInstructionStr()

		instructionFromDB, err := get(tx, instructionStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclave instruction plan")
		}
		if instructionFromDB != nil {
			return nil
		}

		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		instructionKey := getInstructionKey(instructionStr)

		jsonBytes, err := json.Marshal(instruction)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling enclave plan instruction '%+v' in the enclave plan instruction repository", instruction)
		}

		// save it to disk
		if err := bucket.Put(instructionKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' with key '%s' into the enclave db bucket", instruction, instructionKey)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' into the enclave db", instruction)
	}
	return nil
}

func (repository *EnclavePlanInstructionRepository) Executed(
	scheduledInstructionStr string,
	isExecuted bool,
) error {

	if err := repository.enclaveDb.Update(func(tx *bolt.Tx) error {
		instruction, err := get(tx, scheduledInstructionStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclave instruction plan")
		}

		if instruction == nil {
			return stacktrace.Propagate(err, "Imposible to set if the enclave instruction plan '%s' was executed because it doesn't exist in the repository", instruction.GetKurtosisInstructionStr())
		}

		instruction.Executed(isExecuted)

		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		instructionKey := getInstructionKey(scheduledInstructionStr)

		jsonBytes, err := json.Marshal(instruction)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred marshalling enclave plan instruction '%+v' in the enclave plan instruction repository", instruction)
		}

		// save it to disk
		if err := bucket.Put(instructionKey, jsonBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' into the enclave db bucket", instruction)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while setting executed field to '%v' for enclave plan instruction into the enclave db", isExecuted)
	}
	return nil
}

// Get returns an instruction with UUID if exist or nil if it doesn't
func (repository *EnclavePlanInstructionRepository) Get(
	scheduledInstructionStr string,
) (*EnclavePlanInstructionImpl, error) {
	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	instruction := &EnclavePlanInstructionImpl{}
	var err error

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {

		instruction, err = get(tx, scheduledInstructionStr)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting enclave instruction plan from the repository")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the enclave plan instruction from the enclave db")
	}

	return instruction, nil
}

func get(
	tx *bolt.Tx,
	scheduledInstructionStr string,
) (*EnclavePlanInstructionImpl, error) {
	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	instruction := &EnclavePlanInstructionImpl{}

	bucket := tx.Bucket(enclavePlanInstructionBucketName)

	instructionKey := getInstructionKey(scheduledInstructionStr)

	// first get the bytes
	jsonBytes := bucket.Get(instructionKey)

	if jsonBytes == nil {
		return nil, nil
	}

	if err := json.Unmarshal(jsonBytes, &instruction); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred unmarshalling the enclave plan instruction with hash key '%s' from the repository", string(instructionKey))
	}

	return instruction, nil
}

func getInstructionKey(instructionStr string) []byte {
	return []byte(instructionStr)
}
