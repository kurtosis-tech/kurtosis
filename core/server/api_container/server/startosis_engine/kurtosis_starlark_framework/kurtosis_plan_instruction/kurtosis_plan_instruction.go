package kurtosis_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
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
}

func NewKurtosisPlanInstructionWrapper(instruction *KurtosisPlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents, instructionPlanMask *resolver.InstructionsPlanMask, instructionsPlan *instructions_plan.InstructionsPlan) *KurtosisPlanInstructionWrapper {
	return &KurtosisPlanInstructionWrapper{
		KurtosisPlanInstruction: instruction,
		enclaveComponents:       enclaveComponents,
		instructionPlanMask:     instructionPlanMask,
		instructionsPlan:        instructionsPlan,
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

		var enclavePlanInstructionFromMaskMaybe *enclave_plan.EnclavePlanInstruction
		var instructionResolutionStatus enclave_structure.InstructionResolutionStatus
		if builtin.instructionPlanMask.HasNext() {
			_, enclavePlanInstructionFromMaskMaybe = builtin.instructionPlanMask.Next()
			if enclavePlanInstructionFromMaskMaybe != nil {
				instructionResolutionStatus = instructionWrapper.TryResolveWith(enclavePlanInstructionFromMaskMaybe, builtin.enclaveComponents)
			} else {
				instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
			}
		} else {
			instructionResolutionStatus = instructionWrapper.TryResolveWith(nil, builtin.enclaveComponents)
		}

		switch instructionResolutionStatus {
		case enclave_structure.InstructionIsEqual:
			// add instruction from the mask and mark it as executed but not imported from the current enclave plan
			//TODO I don't know if we have to store this in instructionsPlan or EnclavePlan
			//builtin.instructionsPlan.AddScheduledInstruction(scheduledInstructionPulledFromMaskMaybe).Executed(true).ImportedFromCurrentEnclavePlan(false)
			//return scheduledInstructionPulledFromMaskMaybe.GetReturnedValue(), nil //TODO implement returned value
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
			if enclavePlanInstructionFromMaskMaybe != nil {
				builtin.instructionPlanMask.MarkAsInvalid()
				logrus.Debugf("Marking the plan as invalid as instruction '%s' differs from '%s'",
					instructionWrapper.String(), enclavePlanInstructionFromMaskMaybe.GetKurtosisInstructionStr())
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
