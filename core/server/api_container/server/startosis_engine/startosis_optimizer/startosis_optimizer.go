package startosis_optimizer

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_optimizer/graph"
)

type StartosisExecutionOptimizer struct {
	instructionsGraph *graph.InstructionGraph
}

func NewStartosisExecutionOptimizer() *StartosisExecutionOptimizer {
	return &StartosisExecutionOptimizer{
		instructionsGraph: graph.NewInstructionGraph(),
	}
}

func OptimizePlan(enclaveState *graph.InstructionGraph, plan *graph.InstructionGraph) ([]PlannedInstruction, *startosis_errors.ValidationError) {
	enclaveStateInstructionList, err := enclaveState.Traverse()
	if err != nil {
		return nil, startosis_errors.WrapWithValidationError(err, "An error occurred analysing current enclave state")
	}

	planInstructionList, err := plan.Traverse()
	if err != nil {
		return nil, startosis_errors.WrapWithValidationError(err, "Instructions list is invalid")
	}

	var plannedInstructions []PlannedInstruction
	currentIndexOnThePlan := 0
	for _, enclaveInstruction := range enclaveStateInstructionList {
		if planInstructionList[currentIndexOnThePlan].GetHash() != enclaveInstruction.GetHash() {
			if currentIndexOnThePlan > 0 {
				return nil, startosis_errors.NewValidationError("The set of instructions submitted is invalid " +
					"inside this enclave. This Starlark script should be executed in a new enclave")
			}
			plannedInstruction := NewPlannedInstruction(
				enclaveInstruction.GetUuid(),
				enclaveInstruction.GetHash(),
				enclaveInstruction.GetInstruction(),
				true,
				true,
			)
			plannedInstructions = append(plannedInstructions, *plannedInstruction)
		} else {
			currentIndexOnThePlan += 1
			plannedInstruction := NewPlannedInstruction(
				enclaveInstruction.GetUuid(),
				enclaveInstruction.GetHash(),
				enclaveInstruction.GetInstruction(),
				false,
				true,
			)
			plannedInstructions = append(plannedInstructions, *plannedInstruction)
		}
	}

	for i := currentIndexOnThePlan; i < len(planInstructionList); i++ {
		plannedInstruction := NewPlannedInstruction(
			planInstructionList[i].GetUuid(),
			planInstructionList[i].GetHash(),
			planInstructionList[i].GetInstruction(),
			false,
			false,
		)
		plannedInstructions = append(plannedInstructions, *plannedInstruction)
	}
	return plannedInstructions, nil

}
