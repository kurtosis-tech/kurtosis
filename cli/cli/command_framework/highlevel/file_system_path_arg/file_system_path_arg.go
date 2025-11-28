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
	isNotGreedyArg                     = false
	ContinueWithDefaultValidation      = true
	DoNotContinueWithDefaultValidation = false
)

// The File System Path Arg comes with a default validation function (see below getValidationFunc)
// A custom validation function can be used to customize the validation. This validation
// function can be called along the default validation function or alone.
// The custom validation function is called before the default validation function.
// The custom validation function takes the argument value and returns two values:
//   - Validation error
//   - Should we also call the default validation function getValidationFunc if the custom validation succeeded?
//
// If you want to just call the default validation function, set validationFunc to DefaultValidationFunc
// If you want to bypass the default validation function, set validationFunc to BypassDefaultValidationFunc
// If you want to call a custom validation function first, set validationFunc to your validation function and
// then return ContinueWithDefaultValidation or DoNotContinueWithDefaultValidation depending on if you want to
// also call the default validation function or not.
type fileSystemArgumentValidationFunc func(argumentValue string) (error, bool)

var (
	// Use this function to call the default validation function getValidationFunc
	DefaultValidationFunc = func(argumentValue string) (error, bool) {
		return nil, ContinueWithDefaultValidation
	}
	// Use this function to bypass the default validation function getValidationFunc
	BypassDefaultValidationFunc = func(argumentValue string) (error, bool) {
		return nil, DoNotContinueWithDefaultValidation
	}
)

// Prebuilt file path arg which has tab-completion and validation ready out-of-the-box
func NewFilepathArg(
	argKey string,
	isOptional bool,
	defaultValue string,
	validationFunc fileSystemArgumentValidationFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		defaultValue,
		FileSystemPathType_Filepath,
		validationFunc,
	)
}

// Prebuilt dir path arg which has tab-completion and validation ready out-of-the-box
func NewDirpathArg(
	argKey string,
	isOptional bool,
	defaultValue string,
	validationFunc fileSystemArgumentValidationFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		defaultValue,
		FileSystemPathType_Dirpath,
		validationFunc,
	)
}

// Prebuilt file path or dir path arg which has tab-completion and validation ready out-of-the-box
func NewFilepathOrDirpathArg(
	argKey string,
	isOptional bool,
	defaultValue string,
	validationFunc fileSystemArgumentValidationFunc,
) *args.ArgConfig {
	return newFileSystemPathArg(
		argKey,
		isOptional,
		defaultValue,
		FileSystemPathType_FilepathOrDirpath,
		validationFunc,
	)
}

func newFileSystemPathArg(
	// The arg key where this file system path argument will be stored
	argKey string,
	isOptional bool,
	defaultValue string,
	pathType FileSystemPathType,
	validationFunc fileSystemArgumentValidationFunc,
) *args.ArgConfig {

	validate := getValidationFunc(argKey, defaultValue, pathType, validationFunc)

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
	defaultValue string,
	pathType FileSystemPathType,
	validationFunc fileSystemArgumentValidationFunc,
) func(context.Context, *flags.ParsedFlags, *args.ParsedArgs) error {
	return func(ctx context.Context, flags *flags.ParsedFlags, args *args.ParsedArgs) error {

		filePathOrDirpath, err := args.GetNonGreedyArg(argKey)
		if err != nil {
			return stacktrace.Propagate(err, "Expected a value for greedy arg '%v' but didn't find one", argKey)
		}

		if filePathOrDirpath == defaultValue {
			return nil
		}
		filePathOrDirpath = strings.TrimSpace(filePathOrDirpath)
		if filePathOrDirpath == "" {
			return stacktrace.NewError("Received an empty '%v'. It should be a non empty string.", argKey)
		}

		if validationFunc != nil {
			err, continueWithDefaultValidation := validationFunc(filePathOrDirpath)
			if err != nil {
				return stacktrace.Propagate(err, "Error validating %v '%s'", pathType.String(), filePathOrDirpath)
			}
			if !continueWithDefaultValidation {
				return nil
			}
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
