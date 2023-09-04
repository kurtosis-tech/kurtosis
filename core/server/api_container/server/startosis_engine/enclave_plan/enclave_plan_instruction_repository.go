package enclave_plan

import "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"

type EnclavePlanInstructionRepository struct {
	//TODO include db here
}

func NewPersistableInstructionDataRepository() *EnclavePlanInstructionRepository {
	return &EnclavePlanInstructionRepository{}
}

func (repository *EnclavePlanInstructionRepository) Save(
	uuid instructions_plan.ScheduledInstructionUuid,
	data *EnclavePlanInstruction,
) error {
	//TODO implement
	return nil
}

func (repository *EnclavePlanInstructionRepository) Get(
	uuid instructions_plan.ScheduledInstructionUuid,
) (*EnclavePlanInstruction, error) {
	//TODO implement
	return &EnclavePlanInstruction{}, nil
}
