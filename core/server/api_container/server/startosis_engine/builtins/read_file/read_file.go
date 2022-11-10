package read_file

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"go.starlark.net/starlark"
)

const (
	ReadFileBuiltinName = "read_file"

	srcPathArgName = "src_path"
)

func GenerateReadFileBuiltin(provider startosis_modules.ModuleContentProvider) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		srcPath, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		fileContents, interpretationError := provider.GetModuleContents(srcPath)
		if interpretationError != nil {
			return nil, interpretationError
		}
		return starlark.String(fileContents), nil
	}
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, *startosis_errors.InterpretationError) {
	var srcPathArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcPathArgName, &srcPathArg); err != nil {
		return "", explicitInterpretationError(err)
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseFilePath(srcPathArgName, srcPathArg)
	if interpretationErr != nil {
		return "", explicitInterpretationError(interpretationErr)
	}

	return srcPath, nil
}

func explicitInterpretationError(err error) *startosis_errors.InterpretationError {
	return startosis_errors.WrapWithInterpretationError(
		err,
		"Unable to parse arguments of command %s. It should be a non empty string argument pointing to a file inside the module (i.e. \"github.com/kurtosis/module/file.txt\")",
		ReadFileBuiltinName)
}
