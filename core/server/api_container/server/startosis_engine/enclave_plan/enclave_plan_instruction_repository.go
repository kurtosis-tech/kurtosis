package enclave_plan

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

const (
	instructionsSequenceKeyStr = "instructions-sequence"
)

var (
	enclavePlanInstructionBucketName = []byte("enclave-plan-instruction-repository")
	instructionsSequenceKey          = []byte(instructionsSequenceKeyStr)
)

type enclavePlanInstructionRepository struct {
	enclaveDb *enclave_db.EnclaveDB
}

func getOrCreateNewEnclavePlanInstructionRepository(enclaveDb *enclave_db.EnclaveDB) (*enclavePlanInstructionRepository, error) {
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

	repository := &enclavePlanInstructionRepository{
		enclaveDb: enclaveDb,
	}

	return repository, nil
}

func (repository *enclavePlanInstructionRepository) Save(
	uuid instructions_plan.ScheduledInstructionUuid,
	instruction *EnclavePlanInstruction,
) error {

	currentInstructionSequence, err := repository.getInstructionsSequence()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the curren enclave plan instruction sequence")
	}

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

		if err := repository.addNewInstructionUuidInTheSequence(tx, currentInstructionSequence, uuid); err != nil {
			return stacktrace.Propagate(err, "An error occurred adding instruction UUID '%v' in the enclave plan instruction sequence", uuid)
		}

		return nil
	}); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction '%+v' with UUID '%s' into the enclave db", instruction, uuid)
	}
	return nil
}

func (repository *enclavePlanInstructionRepository) Get(
	uuid instructions_plan.ScheduledInstructionUuid,
) (*EnclavePlanInstruction, error) {
	instruction := &EnclavePlanInstruction{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		bucket.Sequence()

		uuidKey := getUuidKey(uuid)

		// first get the bytes
		jsonBytes := bucket.Get(uuidKey)

		// check for existence
		if jsonBytes == nil {
			return stacktrace.NewError("Enclave plan instruction with key UUID '%s' does not exist on the enclave plan instruction repository", uuid)
		}

		if err := json.Unmarshal(jsonBytes, &instruction); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling the enclave plan instruction with UUID '%s' from the repository", uuid)
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the enclave plan instruction with UUID '%s' from the enclave db", uuid)
	}

	return instruction, nil
}

// GetAll returns all the enclave plan instructions stored in the order that these were stored, from the first one to the last one
func (repository *enclavePlanInstructionRepository) GetAll() ([]*EnclavePlanInstruction, error) {

	instructionSequence, err := repository.getInstructionsSequence()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave plan instruction sequence from the repository")
	}

	allEnclavePlanInstructionsMap := map[instructions_plan.ScheduledInstructionUuid]*EnclavePlanInstruction{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		if err := bucket.ForEach(func(keyBytes, instructionBytes []byte) error {
			keyStr := string(keyBytes)

			// skip the instructions-sequence key because it's not an enclave plan instruction value
			if keyStr == instructionsSequenceKeyStr {
				return nil
			}

			if instructionBytes == nil {
				return stacktrace.NewError("Expected to get a non nil enclave plan instruction with UUID '%s' from the repository but a nil value was returned instead", keyStr)
			}

			uuid := instructions_plan.ScheduledInstructionUuid(keyStr)

			instruction := &EnclavePlanInstruction{}
			if err := json.Unmarshal(instructionBytes, &instruction); err != nil {
				return stacktrace.Propagate(err, "An error occurred unmarshalling the enclave plan instruction with UUID '%s' from the repository", keyStr)
			}
			if _, found := allEnclavePlanInstructionsMap[uuid]; found {
				return stacktrace.NewError("The instruction '%v' with UUID '%s' was already set in the all enclave plan instructions map '%+v', this is a bug in Kurtosis", instruction, keyStr, allEnclavePlanInstructionsMap)
			}
			allEnclavePlanInstructionsMap[uuid] = instruction

			return nil
		}); err != nil {
			return stacktrace.Propagate(err, "An error occurred while iterating the service registration repository to get all registrations")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting all service registrations from the service registration repository")
	}

	if len(instructionSequence) != len(allEnclavePlanInstructionsMap) {
		return nil, stacktrace.NewError(
			"The enclave plan repository is corrupted, there should be the same number of values in the "+
				"instruction sequence that the enclave plan instructions stored, but these are not equal,"+
				"there are '%v' values in the instructions sequence and '%v' enclave plan instructions stored;"+
				" this is a bug in Kurtosis", len(instructionSequence), len(allEnclavePlanInstructionsMap))
	}

	allEnclavePlanInstructions := []*EnclavePlanInstruction{}

	for _, uuid := range instructionSequence {
		allEnclavePlanInstructions = append(allEnclavePlanInstructions, allEnclavePlanInstructionsMap[uuid])
	}

	return allEnclavePlanInstructions, nil
}

func (repository *enclavePlanInstructionRepository) Size() (int, error) {
	var size int

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		stats := bucket.Stats()

		size = stats.KeyN

		return nil
	}); err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred while getting the enclave plan instruction size from the repository")
	}

	return size, nil
}

func (repository *enclavePlanInstructionRepository) addNewInstructionUuidInTheSequence(
	tx *bolt.Tx,
	currentInstructionSequence []instructions_plan.ScheduledInstructionUuid,
	uuid instructions_plan.ScheduledInstructionUuid,
) error {

	newInstructionSequence := append(currentInstructionSequence, uuid)

	bucket := tx.Bucket(enclavePlanInstructionBucketName)

	jsonBytes, err := json.Marshal(newInstructionSequence)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred marshalling enclave plan instruction sequence '%+v' in the enclave plan instruction repository", newInstructionSequence)
	}

	// save it to disk
	if err := bucket.Put(instructionsSequenceKey, jsonBytes); err != nil {
		return stacktrace.Propagate(err, "An error occurred while saving enclave plan instruction sequence '%+v' with key '%s' into the enclave db bucket", newInstructionSequence, instructionsSequenceKey)
	}

	return nil

}

func (repository *enclavePlanInstructionRepository) getInstructionsSequence() ([]instructions_plan.ScheduledInstructionUuid, error) {
	instructionsSequence := []instructions_plan.ScheduledInstructionUuid{}

	if err := repository.enclaveDb.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(enclavePlanInstructionBucketName)

		// first get the bytes
		jsonBytes := bucket.Get(instructionsSequenceKey)

		// if nil means that no instruction has been stored yet
		if jsonBytes == nil {
			return nil
		}

		if err := json.Unmarshal(jsonBytes, &instructionsSequence); err != nil {
			return stacktrace.Propagate(err, "An error occurred unmarshalling the enclave plan instruction sequence from the repository")
		}

		return nil
	}); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting the enclave plan instruction sequence from the enclave db")
	}
	return instructionsSequence, nil
}

func getUuidKey(uuid instructions_plan.ScheduledInstructionUuid) []byte {
	return []byte(uuid)
}
