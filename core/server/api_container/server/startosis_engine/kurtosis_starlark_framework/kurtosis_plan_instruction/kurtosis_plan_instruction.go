package kurtosis_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan/resolver"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
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

	instructionPlanMask *resolver.InstructionsPlanMask

	// TODO: This can be changed to KurtosisPlanInstructionInternal when we get rid of KurtosisInstruction
	instructionsPlan *instructions_plan.InstructionsPlan
}

func NewKurtosisPlanInstructionWrapper(instruction *KurtosisPlanInstruction, instructionPlanMask *resolver.InstructionsPlanMask, instructionsPlan *instructions_plan.InstructionsPlan) *KurtosisPlanInstructionWrapper {
	return &KurtosisPlanInstructionWrapper{
		KurtosisPlanInstruction: instruction,
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

		var instructionPulledFromMaskIdx int
		var instructionPulledFromMaskMaybe *instructions_plan.ScheduledInstruction
		if builtin.instructionPlanMask.HasNext() {
			instructionPulledFromMaskIdx, instructionPulledFromMaskMaybe = builtin.instructionPlanMask.Next()
			if instructionPulledFromMaskMaybe != nil && instructionPulledFromMaskMaybe.GetInstruction().String() != instructionWrapper.String() {
				// if the instructions differs, then the mask is invalid
				builtin.instructionPlanMask.MarkAsInvalid()
				logrus.Debugf("The instruction number %d in the plan mask did not match the newly interpreter "+
					"instruction and therefore the plan mask was marked as invalid:\nInstruction from mask - '%s'"+
					"\nInstruction from interpretation: '%s'",
					instructionPulledFromMaskIdx,
					instructionPulledFromMaskMaybe.GetInstruction().String(),
					instructionWrapper.String())
				// TODO: we could interrupt the interpretation here, because with an invalid mask the list of
				//  instruction generated will be invalid anyway. Though we currently don't have a nive way to
				//  interrupt an interpretation in progress (other than by throwing an error, which would be
				//  misleading here)
				//  To properly solve that, I think we should switch to an interactive interpretation where we
				//  interpret each instruction one after the other, and evaluating the state after each step
			}
		}

		if instructionPulledFromMaskMaybe != nil {
			// If there's a mask for this instruction, add the mask the plan and returned the mask's returned value
			builtin.instructionsPlan.AddScheduledInstruction(instructionPulledFromMaskMaybe).Executed(true).ImportedFromCurrentEnclavePlan(false)
			return instructionPulledFromMaskMaybe.GetReturnedValue(), nil
		} else {
			// otherwise add the instruction as a new one to the plan and return its own returned value
			if err := builtin.instructionsPlan.AddInstruction(instructionWrapper, returnedFutureValue); err != nil {
				return nil, startosis_errors.WrapWithInterpretationError(err,
					"Unable to add Kurtosis instruction '%s' at position '%s' to the current plan being assembled. This is a Kurtosis internal bug",
					instructionWrapper.String(),
					instructionWrapper.GetPositionInOriginalScript().String())
			}
			return returnedFutureValue, nil
		}
	}
}
