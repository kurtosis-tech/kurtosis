package kurtosis_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/enclave_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
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

	enclavePlanInstructionRepository *enclave_plan_instruction.EnclavePlanInstructionRepository
}

func NewKurtosisPlanInstructionWrapper(
	instruction *KurtosisPlanInstruction,
	enclaveComponents *enclave_structure.EnclaveComponents,
	instructionPlanMask *resolver.InstructionsPlanMask,
	instructionsPlan *instructions_plan.InstructionsPlan,
	enclavePlanInstructionRepository *enclave_plan_instruction.EnclavePlanInstructionRepository,
) *KurtosisPlanInstructionWrapper {
	return &KurtosisPlanInstructionWrapper{
		KurtosisPlanInstruction:          instruction,
		enclaveComponents:                enclaveComponents,
		instructionPlanMask:              instructionPlanMask,
		instructionsPlan:                 instructionsPlan,
		enclavePlanInstructionRepository: enclavePlanInstructionRepository,
	}
}

func (builtin *KurtosisPlanInstructionWrapper) CreateBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		wrappedBuiltin, interpretationErr := kurtosis_starlark_framework.WrapKurtosisBaseBuiltin(builtin.KurtosisBaseBuiltin, thread, args, kwargs)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		instructionWrapper := newKurtosisPlanInstructionInternal(wrappedBuiltin, builtin.Capabilities(), builtin.DefaultDisplayArguments)
		locatorOfModuleInWhichInstructionIsBeingInterpreted := thread.Name
		returnedFutureValue, interpretationErr := instructionWrapper.interpret(locatorOfModuleInWhichInstructionIsBeingInterpreted)
		if interpretationErr != nil {
			return nil, interpretationErr
		}

		var scheduledInstructionPulledFromMaskMaybe *instructions_plan.ScheduledInstruction
		var instructionResolutionStatus enclave_structure.InstructionResolutionStatus
		if builtin.instructionPlanMask.HasNext() {
			_, scheduledInstructionPulledFromMaskMaybe = builtin.instructionPlanMask.Next()
			if scheduledInstructionPulledFromMaskMaybe != nil {

				scheduledInstructionUuid := scheduledInstructionPulledFromMaskMaybe.GetUuid()
				enclavePlanInstruction, err := builtin.enclavePlanInstructionRepository.Get(scheduledInstructionUuid)
				if err != nil {
					return nil, stacktrace.Propagate(err, "An error occurred getting the enclave plan instruction with scheduled instruction UUID '%v'", scheduledInstructionUuid)
				}

				instructionResolutionStatus = instructionWrapper.TryResolveWith(enclavePlanInstruction, builtin.enclaveComponents)
			} else {
				instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
			}
		} else {
			instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
		}

		switch instructionResolutionStatus {
		case enclave_structure.InstructionIsEqual:
			// add instruction from the mask and mark it as executed but not imported from the current enclave plan
			builtin.instructionsPlan.AddScheduledInstruction(scheduledInstructionPulledFromMaskMaybe).Executed(true)

			if err := builtin.enclavePlanInstructionRepository.Executed(scheduledInstructionPulledFromMaskMaybe.GetUuid(), true); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred setting enclave instruction plan with UUID '%v' has executed", scheduledInstructionPulledFromMaskMaybe.GetUuid())
			}
			return scheduledInstructionPulledFromMaskMaybe.GetReturnedValue(), nil
		case enclave_structure.InstructionIsUpdate:
			// otherwise add the instruction as a new one to the plan and return its own returned value
			if _, err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to add Kurtosis instruction '%s' at position '%s' to the plan currently being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			return returnedFutureValue, nil
		case enclave_structure.InstructionIsUnknown:
			scheduledInstruction, err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue)
			if err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err,
					"Unable to add Kurtosis instruction '%s' at position '%s' to the plan currently being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			if scheduledInstructionPulledFromMaskMaybe != nil {
				builtin.instructionPlanMask.MarkAsInvalid()
				logrus.Debugf("Marking the plan as invalid as instruction '%s' differs from '%s'",
					instructionWrapper.String(), scheduledInstructionPulledFromMaskMaybe.GetInstruction().String())
			}
			kurtosisInstructionStr := scheduledInstruction.GetInstruction().String()
			enclaveCapabilities := scheduledInstruction.GetInstruction().GetCapabilites().GetEnclavePlanCapabilities()
			enclavePlanInstruction := enclave_plan_instruction.NewEnclavePlanInstructionImpl(kurtosisInstructionStr, enclaveCapabilities)
			if err := builtin.enclavePlanInstructionRepository.Save(scheduledInstruction.GetUuid(), enclavePlanInstruction); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred saving enclave instruction plan '%+v' with scheduled instruction with UUID '%s'", enclavePlanInstruction, scheduledInstruction.GetUuid())
			}

			return returnedFutureValue, nil
		case enclave_structure.InstructionIsNotResolvableAbort:
			// if the instructions differs, then the mask is invalid
			builtin.instructionPlanMask.MarkAsInvalid()
			logrus.Debugf("Marking the plan as invalid as instruction '%s' had the following resolution status: '%s'",
				instructionWrapper.String(), instructionResolutionStatus)
			if _, err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
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
