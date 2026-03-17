package kurtosis_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

type KurtosisPlanInstruction struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltin

	Capabilities func() KurtosisPlanInstructionCapabilities

	DefaultDisplayArguments map[string]bool
}

// KurtosisPlanInstructionWrapper is a convenience wrapper to store the instructionQueue necessary in the
// CreateBuiltin next to the KurtosisPlanInstruction, without polluting its declaration
type KurtosisPlanInstructionWrapper struct {
	*KurtosisPlanInstruction

	enclaveComponents   *enclave_structure.EnclaveComponents
	instructionPlanMask *resolver.InstructionsPlanMask

	// TODO: This can be changed to KurtosisPlanInstructionInternal when we get rid of KurtosisInstruction
	instructionsPlan *instructions_plan.InstructionsPlan

	starlarkValueSerde *kurtosis_types.StarlarkValueSerde
}

func NewKurtosisPlanInstructionWrapper(instruction *KurtosisPlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents, starlarkValueSerde *kurtosis_types.StarlarkValueSerde, instructionPlanMask *resolver.InstructionsPlanMask, instructionsPlan *instructions_plan.InstructionsPlan) *KurtosisPlanInstructionWrapper {
	return &KurtosisPlanInstructionWrapper{
		KurtosisPlanInstruction: instruction,
		enclaveComponents:       enclaveComponents,
		instructionPlanMask:     instructionPlanMask,
		instructionsPlan:        instructionsPlan,
		starlarkValueSerde:      starlarkValueSerde,
	}
}

func (builtin *KurtosisPlanInstructionWrapper) CreateBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		wrappedBuiltin, interpretationErr := kurtosis_starlark_framework.WrapKurtosisBaseBuiltin(builtin.KurtosisBaseBuiltin, thread, args, kwargs)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		instructionWrapper := newKurtosisPlanInstructionInternal(wrappedBuiltin, builtin.Capabilities(), builtin.DefaultDisplayArguments)

		returnedFutureValue, interpretationErr := instructionWrapper.interpret()
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		var enclavePlanInstructionPulledFromMaskMaybe *enclave_plan_persistence.EnclavePlanInstruction
		var instructionResolutionStatus enclave_structure.InstructionResolutionStatus
		if builtin.instructionPlanMask.HasNext() {
			_, enclavePlanInstructionPulledFromMaskMaybe = builtin.instructionPlanMask.Next()
			if enclavePlanInstructionPulledFromMaskMaybe != nil {
				instructionResolutionStatus = instructionWrapper.TryResolveWith(enclavePlanInstructionPulledFromMaskMaybe, builtin.enclaveComponents)
			} else {
				instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
			}
		} else {
			instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
		}

		switch instructionResolutionStatus {
		case enclave_structure.InstructionIsEqual:
			// Build a scheduled instruction from a mix of the instruction and the EnclavePlanInstruction from the mask
			// The important thing is to keep the returnedValue from the enclavePlan instruction such that runtime values
			// IDs are kept. We also re-use the same UUID to make sure it's stable, but right now enclave plan instruction
			// UUIDs are not used so it's not critical here
			// Lastly, this scheduled instruction is marked as "EXECUTED" such that it will later be skipped by the
			// executor
			returnedValue, err := builtin.starlarkValueSerde.Deserialize(enclavePlanInstructionPulledFromMaskMaybe.ReturnedValue)
			if err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "The instruction was resolved with "+
					"an instruction from the enclave plan, but the result of this instruction could not be"+
					"deserialized. The instruction was: '%s' and its result was: '%s'",
					enclavePlanInstructionPulledFromMaskMaybe.StarlarkCode,
					enclavePlanInstructionPulledFromMaskMaybe.ReturnedValue)
			}
			scheduledInstruction := instructions_plan.NewScheduledInstruction(
				types.ScheduledInstructionUuid(enclavePlanInstructionPulledFromMaskMaybe.Uuid),
				instructionWrapper,
				returnedValue,
			).Executed(true)
			builtin.instructionsPlan.AddScheduledInstruction(scheduledInstruction).Executed(true)
			return returnedValue, nil
		case enclave_structure.InstructionIsUpdate:
			// otherwise add the instruction as a new one to the plan and return its own returned value
			if err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to add Kurtosis instruction '%s' at position '%s' to the plan currently being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			return returnedFutureValue, nil
		case enclave_structure.InstructionIsUnknown:
			if err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err,
					"Unable to add Kurtosis instruction '%s' at position '%s' to the plan currently being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			if enclavePlanInstructionPulledFromMaskMaybe != nil { // why is it that the mask is invalid if this is the case? and why not make this check before adding the instruction to the plan?
				builtin.instructionPlanMask.MarkAsInvalid()
				logrus.Debugf("Marking the plan as invalid as instruction '%s' differs from '%s'",
					instructionWrapper.String(), enclavePlanInstructionPulledFromMaskMaybe.StarlarkCode)
			}
			return returnedFutureValue, nil
		case enclave_structure.InstructionIsNotResolvableAbort:
			// if the instructions differs, then the mask is invalid
			builtin.instructionPlanMask.MarkAsInvalid()
			logrus.Debugf("Marking the plan as invalid as instruction '%s' had the following resolution status: '%s'",
				instructionWrapper.String(), instructionResolutionStatus)
			if err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err,
					"Unable to add Kurtosis instruction '%s' at position '%s' to the plan currently being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			return returnedFutureValue, nil
		}
		return nil, stacktrace.NewError("Unexpected error, resolution status of instruction '%s' did not match any of the covered case.", instructionResolutionStatus)
	}
}
