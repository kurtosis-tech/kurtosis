package startosis_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisInterpreter struct {
}

type SerializedInterpretationOutput struct {
	Output string
}
type InterpretationError struct {
	Error string
}

func NewStartosisInterpreter() *StartosisInterpreter {
	// TODO(gb): build the bindings to populate an instruction list on interpret
	return &StartosisInterpreter{}
}

// Interpret interprets the Startosis script and produce different outputs:
//   - The serialized output of the interpretation (what the Startosis script printed)
//   - A potential interpretation error that the writer of the script should be aware of (syntax error in the Startosis
//     code, inconsistent). Can be nil if the script was successfully interpreted
//   - The list of Kurtosis instructions that was generated based on the interpretation of the script. It can be empty
//     if the interpretation of the script failed
//   - An error if something unexpected happens (crash independent of the Startosis script). This should be as rare as
//     possible
func (interpreter *StartosisInterpreter) Interpret(serializedScript string) (*SerializedInterpretationOutput, *InterpretationError, []kurtosis_instruction.KurtosisInstruction, error) {
	// TODO(gb): implement
	return nil, nil, nil, stacktrace.NewError("not implemented")
}
