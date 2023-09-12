package enclave_plan_persistence

import (
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/database_accessors/enclave_db"
	"github.com/kurtosis-tech/stacktrace"
	"go.etcd.io/bbolt"
)

const (
	enclavePlanBucketName = "EnclavePlan"
	// We persist only one enclave plan per enclave, so we can use a constant key here
	enclavePlanConstantKey = "EnclavePlan"
)

type EnclavePlan struct {
	EnclavePlanInstructions []*EnclavePlanInstruction `json:"enclavePlanInstructions"`
}

func NewEnclavePlan() *EnclavePlan {
	return &EnclavePlan{
		EnclavePlanInstructions: []*EnclavePlanInstruction{},
	}
}

func (enclavePlan *EnclavePlan) PartialDeepClone(indexOfFirstInstructionToNotClone int) *EnclavePlan {
	newPlan := &EnclavePlan{
		EnclavePlanInstructions: []*EnclavePlanInstruction{},
	}
	for idx, instruction := range enclavePlan.EnclavePlanInstructions {
		if idx >= indexOfFirstInstructionToNotClone {
			return newPlan
		}
		newPlan.AppendInstruction(instruction.Clone())
	}
	return newPlan
}

func Load(enclaveDb *enclave_db.EnclaveDB) (*EnclavePlan, error) {
	var persistedEnclavePlanSerializedMaybe []byte
	err := enclaveDb.View(func(tx *bbolt.Tx) error {
		enclavePlanBucket := tx.Bucket([]byte(enclavePlanBucketName))
		if enclavePlanBucket == nil {
			return nil
		}
		persistedEnclavePlanSerializedMaybe = enclavePlanBucket.Get([]byte(enclavePlanConstantKey))
		return nil
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred reading enclave plan value from enclave database")
	}
	if persistedEnclavePlanSerializedMaybe == nil {
		// No enclave plan persisted for this enclave just yet, return an empty plan
		return NewEnclavePlan(), nil
	}
	persistedEnclavePlan := new(EnclavePlan)
	if err := json.Unmarshal(persistedEnclavePlanSerializedMaybe, persistedEnclavePlan); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred deserializing enclave plan from enclave database")
	}
	return persistedEnclavePlan, nil
}

func (enclavePlan *EnclavePlan) Persist(enclaveDb *enclave_db.EnclaveDB) error {
	serializedEnclavePlanBytes, err := json.Marshal(enclavePlan)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred serializing enclave plan to enclave database")
	}
	err = enclaveDb.Update(func(tx *bbolt.Tx) error {
		enclavePlanBucket, err := tx.CreateBucketIfNotExists([]byte(enclavePlanBucketName))
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred getting or creating enclave plan bucket into Bbolt")
		}
		if err := enclavePlanBucket.Put([]byte(enclavePlanConstantKey), serializedEnclavePlanBytes); err != nil {
			return stacktrace.Propagate(err, "An error occurred persisting enclave plan to to Bbolt database")
		}
		return nil
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred persisting enclave plan to enclave database")
	}
	return nil
}

func (enclavePlan *EnclavePlan) AppendInstruction(instruction *EnclavePlanInstruction) {
	enclavePlan.EnclavePlanInstructions = append(enclavePlan.EnclavePlanInstructions, instruction)
}

func (enclavePlan *EnclavePlan) GeneratePlan() []*EnclavePlanInstruction {
	return enclavePlan.EnclavePlanInstructions
}

func (enclavePlan *EnclavePlan) Size() int {
	return len(enclavePlan.EnclavePlanInstructions)
}
