package enclave_plan

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

type EnclavePlan struct {
	enclavePlanInstructionRepository *enclavePlanInstructionRepository
}

func NewEnclavePlan(enclavePlanInstructionRepository *enclavePlanInstructionRepository) *EnclavePlan {
	return &EnclavePlan{
		enclavePlanInstructionRepository: enclavePlanInstructionRepository,
	}
}

func (plan *EnclavePlan) Size() (int, error) {
	size, err := plan.enclavePlanInstructionRepository.Size()
	if err != nil {
		return 0, stacktrace.Propagate(err, "An error occurred getting the enclave plan size from the repository")
	}
	return size, nil
}

// TODO add a test in order to checkt if the returnes list is in order.
func (plan *EnclavePlan) GeneratePlan() ([]*EnclavePlanInstruction, *startosis_errors.InterpretationError) {
	generatedPlan, err := plan.enclavePlanInstructionRepository.GetAll()
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("Unexpected error generating the Kurtosis Enclave Instructions plan.")
	}
	return generatedPlan, nil
}

func (plan *EnclavePlan) AddInstruction(
	instruction kurtosis_instruction.KurtosisInstruction,
	returnedValue starlark.Value,
	instructionCapabilities kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities,
) error {

	//TODO I'm not pretty sure if we should generated here or take it from the InstructionPlan object because
	//TODO both have to match ???
	generatedUuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return stacktrace.Propagate(err, "Unable to generate a random UUID for instruction '%s' to add it to the plan", instruction.String())
	}

	scheduledInstructionUuid := instructions_plan.ScheduledInstructionUuid(generatedUuid)

	instructionStr := instruction.String()
	capabilities := instructionCapabilities.GetEnclavePlanCapabilities()
	returnedValueStr := returnedValue.String()

	enclavePlanInstruction := NewEnclavePlanInstruction(instructionStr, capabilities, returnedValueStr)

	if err := plan.enclavePlanInstructionRepository.Save(scheduledInstructionUuid, enclavePlanInstruction); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving enclave plan instruction '%+v' with UUID '%s'", enclavePlanInstruction, scheduledInstructionUuid)
	}
	return nil
}
