package kurtosis_plan_instruction

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
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

	// TODO: This can be changed to KurtosisPlanInstructionInternal when we get rid of KurtosisInstruction
	instructionQueue *[]kurtosis_instruction.KurtosisInstruction
}

func NewKurtosisPlanInstructionWrapper(instruction *KurtosisPlanInstruction, instructionQueue *[]kurtosis_instruction.KurtosisInstruction) *KurtosisPlanInstructionWrapper {
	return &KurtosisPlanInstructionWrapper{
		KurtosisPlanInstruction: instruction,
		instructionQueue:        instructionQueue,
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

		// before returning, automatically add instruction to queue
		*builtin.instructionQueue = append(*builtin.instructionQueue, instructionWrapper)

		return returnedFutureValue, nil
	}
}
