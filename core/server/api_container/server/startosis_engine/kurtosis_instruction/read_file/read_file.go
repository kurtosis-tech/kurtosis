package read_file

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"go.starlark.net/starlark"
	"strings"
)

const (
	ReadFileBuiltinName = "read_file"

	srcPathArgName = "src_path"
)

type ReadFileInstruction struct {
	position kurtosis_instruction.InstructionPosition
	srcPath  string
}

func GenerateReadFileBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, provider startosis_modules.ModuleContentProvider) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		srcPath, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		execInstruction := NewReadFileInstruction(*shared_helpers.GetPositionFromThread(thread), srcPath)
		*instructionsQueue = append(*instructionsQueue, execInstruction)
		fileContents, interpretationError := readFile(shared_helpers.GetFileNameFromThread(thread), srcPath, provider)
		if interpretationError != nil {
			return nil, interpretationError
		}
		return starlark.String(fileContents), nil
	}
}

func NewReadFileInstruction(position kurtosis_instruction.InstructionPosition, srcPath string) *ReadFileInstruction {
	return &ReadFileInstruction{
		position: position,
		srcPath:  srcPath,
	}
}

func (instruction *ReadFileInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *ReadFileInstruction) GetCanonicalInstruction() string {
	buffer := new(strings.Builder)
	buffer.WriteString(ReadFileBuiltinName + "(")
	buffer.WriteString(srcPathArgName + "=\"")
	buffer.WriteString(fmt.Sprintf("%v\")", instruction.srcPath))
	return buffer.String()
}

func (instruction *ReadFileInstruction) Execute(_ context.Context, _ *startosis_executor.ExecutionEnvironment) error {
	// this does nothing as file gets read during interpretation
	return nil
}

func (instruction *ReadFileInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *ReadFileInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func readFile(fileBeingInterpretedName string, fileToRead string, provider startosis_modules.ModuleContentProvider) (string, *startosis_errors.InterpretationError) {
	if provider.IsGithubPath(fileToRead) {
		fileContents, err := provider.GetModuleContents(fileToRead)
		if err != nil {
			return "", startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while reading from '%v' from Github", fileToRead))
		}
		return fileContents, nil
	}

	fileContents, err := provider.GetFileAtRelativePath(fileToRead, fileBeingInterpretedName)
	if err != nil {
		return "", startosis_errors.NewInterpretationError(fmt.Sprintf("An error occurred while reading from '%v' from disk", fileToRead))
	}

	return fileContents, nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, *startosis_errors.InterpretationError) {

	var srcPathArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcPathArgName, &srcPathArg); err != nil {
		return "", startosis_errors.NewInterpretationError(err.Error())
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseSrcPath(srcPathArg)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	return srcPath, nil
}
