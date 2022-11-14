package upload_files

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_executor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	srcPathArgName = "src_path"

	ensureCompressedFileIsLesserThanGRPCLimit = false
)

type UploadFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork
	provider       startosis_modules.ModuleContentProvider

	position kurtosis_instruction.InstructionPosition
	srcPath  string

	pathOnDisk string
}

func GenerateUploadFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, provider startosis_modules.ModuleContentProvider, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		srcPath, interpretationError := parseStartosisArgs(b, args, kwargs)
		if interpretationError != nil {
			return nil, interpretationError
		}
		pathOnDisk, interpretationError := provider.GetOnDiskAbsoluteFilePath(srcPath)
		if interpretationError != nil {
			return nil, interpretationError
		}
		uploadInstruction := NewUploadFilesInstruction(*shared_helpers.GetCallerPositionFromThread(thread), serviceNetwork, provider, srcPath, pathOnDisk)
		*instructionsQueue = append(*instructionsQueue, uploadInstruction)
		return starlark.String(uploadInstruction.GetPositionInOriginalScript().MagicString(shared_helpers.ArtifactUUIDSuffix)), nil
	}
}

func NewUploadFilesInstruction(position kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_modules.ModuleContentProvider, srcPath string, pathOnDisk string) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		srcPath:        srcPath,
		provider:       provider,
		pathOnDisk:     pathOnDisk,
	}
}

func (instruction *UploadFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return &instruction.position
}

func (instruction *UploadFilesInstruction) GetCanonicalInstruction() string {
	return shared_helpers.CanonicalizeInstruction(UploadFilesBuiltinName, starlark.StringDict{
		srcPathArgName: starlark.String(instruction.srcPath),
	}, &instruction.position)
}

func (instruction *UploadFilesInstruction) Execute(_ context.Context, environment *startosis_executor.ExecutionEnvironment) error {
	compressedData, err := shared_utils.CompressPath(instruction.pathOnDisk, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while compressing the files '%v'", instruction.pathOnDisk)
	}
	filesArtifactUuid, err := instruction.serviceNetwork.UploadFilesArtifact(compressedData)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while uploading the compressed contents\n'%v'", compressedData)
	}
	environment.SetArtifactUuid(instruction.position.MagicString(shared_helpers.ArtifactUUIDSuffix), string(filesArtifactUuid))
	logrus.Infof("Succesfully uploaded files from instruction '%v' to '%v'", instruction.position.String(), filesArtifactUuid)
	return nil
}

func (instruction *UploadFilesInstruction) String() string {
	return instruction.GetCanonicalInstruction()
}

func (instruction *UploadFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, *startosis_errors.InterpretationError) {

	var srcPathArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcPathArgName, &srcPathArg); err != nil {
		return "", startosis_errors.NewInterpretationError(err.Error())
	}

	srcPath, interpretationErr := kurtosis_instruction.ParseFilePath(srcPathArgName, srcPathArg)
	if interpretationErr != nil {
		return "", interpretationErr
	}

	return srcPath, nil
}
