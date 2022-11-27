package upload_files

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_modules"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	srcArgName = "src"

	artifactIdArgName            = "artifact_id?"
	nonOptionalArtifactIdArgName = "artifact_id"

	ensureCompressedFileIsLesserThanGRPCLimit = false

	emptyStarlarkString = starlark.String("")
)

type UploadFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork
	provider       startosis_modules.ModuleContentProvider

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	src        string
	artifactId enclave_data_directory.FilesArtifactUUID

	pathOnDisk string
}

func GenerateUploadFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, provider startosis_modules.ModuleContentProvider, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		uploadInstruction := newEmptyUploadFilesInstruction(instructionPosition, serviceNetwork, provider)
		if interpretationError := uploadInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, uploadInstruction)
		return starlark.String(uploadInstruction.artifactId), nil
	}
}

func NewUploadFilesInstruction(position *kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_modules.ModuleContentProvider, src string, pathOnDisk string, artifactId enclave_data_directory.FilesArtifactUUID, starlarkKwargs starlark.StringDict) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		src:            src,
		provider:       provider,
		pathOnDisk:     pathOnDisk,
		artifactId:     artifactId,
		starlarkKwargs: starlarkKwargs,
	}
}

func newEmptyUploadFilesInstruction(position *kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_modules.ModuleContentProvider) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		provider:       provider,
		src:            "",
		pathOnDisk:     "",
		artifactId:     "",
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *UploadFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *UploadFilesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.KurtosisInstruction {
	args := []*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[srcArgName]), srcArgName, kurtosis_instruction.Representative),
		binding_constructors.NewKurtosisInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[nonOptionalArtifactIdArgName]), nonOptionalArtifactIdArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewKurtosisInstruction(instruction.position.ToAPIType(), UploadFilesBuiltinName, instruction.String(), args)
}

func (instruction *UploadFilesInstruction) Execute(_ context.Context) (*string, error) {
	compressedData, err := shared_utils.CompressPath(instruction.pathOnDisk, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while compressing the files '%v'", instruction.pathOnDisk)
	}
	err = instruction.serviceNetwork.UploadFilesArtifactToTargetArtifactUUID(compressedData, instruction.artifactId)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while uploading the compressed contents\n'%v'", compressedData)
	}
	logrus.Infof("Succesfully uploaded files from instruction '%v' to '%v'", instruction.position.String(), instruction.artifactId)
	return nil, nil
}

func (instruction *UploadFilesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(UploadFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *UploadFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	// this doesn't do anything but can't return an error as the validator runs this regardless
	// this is a no-op
	return nil
}

func (instruction *UploadFilesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var srcPathArg starlark.String
	var artifactIdArg = emptyStarlarkString
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcArgName, &srcPathArg, artifactIdArgName, &artifactIdArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", UploadFilesBuiltinName, args, kwargs)
	}

	if artifactIdArg == emptyStarlarkString {
		placeHolderArtifactId, err := enclave_data_directory.NewFilesArtifactUUID()
		if err != nil {
			return startosis_errors.NewInterpretationError("An empty or no artifact_uuid was passed, we tried creating one but failed")
		}
		artifactIdArg = starlark.String(placeHolderArtifactId)
	}

	instruction.starlarkKwargs[srcArgName] = srcPathArg
	instruction.starlarkKwargs[nonOptionalArtifactIdArgName] = artifactIdArg
	instruction.starlarkKwargs.Freeze()

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	artifactId, interpretationErr := kurtosis_instruction.ParseArtifactId(nonOptionalArtifactIdArgName, artifactIdArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	pathOnDisk, interpretationError := instruction.provider.GetOnDiskAbsoluteFilePath(srcPath)
	if interpretationError != nil {
		return interpretationError
	}

	instruction.src = srcPath
	instruction.artifactId = artifactId
	instruction.pathOnDisk = pathOnDisk
	return nil
}
