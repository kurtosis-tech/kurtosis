package file_system_path_arg

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/args"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_framework/lowlevel/flags"
	"github.com/kurtosis-tech/stacktrace"
	"os"
	"strings"
)

const (
	isNotGreedyArg = false
	defaultValue   = ""
)

// This function type can be used to create custom argument value conditions where the validation will be omitted
// To omit the argument value the function has to return true when the condition is accomplished
type fileSystemArgumentValidationExceptionFunc func(argumentValue string) (isValid bool)

// Prebuilt file path arg which has tab-completion and validation ready out-of-the-box
func NewFilepathArg(
	argKey string,
	isOptional bool,
	validationExceptionFunc fileSystemArgumentValidationExceptionFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		FileSystemPathType_Filepath,
		validationExceptionFunc,
	)
}

// Prebuilt dir path arg which has tab-completion and validation ready out-of-the-box
func NewDirpathArg(
	argKey string,
	isOptional bool,
	validationExceptionFunc fileSystemArgumentValidationExceptionFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		FileSystemPathType_Dirpath,
		validationExceptionFunc,
	)
}

// Prebuilt file path or dir path arg which has tab-completion and validation ready out-of-the-box
func NewFilepathOrDirpathArg(
	argKey string,
	isOptional bool,
	validationExceptionFunc fileSystemArgumentValidationExceptionFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		FileSystemPathType_FilepathOrDirpath,
		validationExceptionFunc,
	)
}

func newFileSystemPathArg(
	// The arg key where this file system path argument will be stored
	argKey string,
	isOptional bool,
	pathType FileSystemPathType,
	validationExceptionFunc fileSystemArgumentValidationExceptionFunc,
) *args.ArgConfig {

	validate := getValidationFunc(argKey, pathType, validationExceptionFunc)

	return &args.ArgConfig{
		Key:            argKey,
		IsOptional:     isOptional,
		DefaultValue:   defaultValue,
		IsGreedy:       isNotGreedyArg,
		ValidationFunc: validate,
		//No custom completion because we are enabling default shell's file completion with ShouldShellProvideDefaultFileCompletion
		ArgCompletionProvider: args.NewDefaultShellFileCompletionProvider(),
	}
}

// Create a validation function using the previously-created
func getValidationFunc(
	argKey string,
	pathType FileSystemPathType,
	validationExceptionFunc fileSystemArgumentValidationExceptionFunc,
) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {

		filePathOrDirpath, err := args.GetNonGreedyArg(argKey)
		if err != nil {
			return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
		}

		filePathOrDirpath = strings.TrimSpace(filePathOrDirpath)
		if filePathOrDirpath == "" {
			return stacktrace.NewError("Received an empty '%v'. It should be a non empty string.", argKey)
		}

		isExceptionalValue := validationExceptionFunc(filePathOrDirpath)
		if isExceptionalValue {
			return nil
		}

		fileInfo, err := os.Stat(filePathOrDirpath)
		if err != nil {
			return stacktrace.Propagate(err, "Error reading %v '%s'", pathType.String(), filePathOrDirpath)
		}

		isRegularFile := fileInfo.Mode().IsRegular()
		isDirectory := fileInfo.Mode().IsDir()

		switch pathType {
		case FileSystemPathType_Filepath:
			if !isRegularFile {
				return stacktrace.Propagate(err, "The file system path '%v' does not point to a file on disk", filePathOrDirpath)
			}
		case FileSystemPathType_Dirpath:
			if !isDirectory {
				return stacktrace.Propagate(err, "The file system path '%v' does not point to a directory on disk", filePathOrDirpath)
			}
		case FileSystemPathType_FilepathOrDirpath:
			if !isRegularFile && !isDirectory {
				return stacktrace.Propagate(err, "The file system path '%v' does not point to a file or to a directory on disk", filePathOrDirpath)
			}
		}

		return nil
	}
}
