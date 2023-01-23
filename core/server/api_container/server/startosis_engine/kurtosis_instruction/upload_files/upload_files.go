package upload_files

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction/shared_helpers"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	srcArgName = "src"

	artifactNameArgName = "name"

	ensureCompressedFileIsLesserThanGRPCLimit = false
)

type UploadFilesInstruction struct {
	serviceNetwork service_network.ServiceNetwork
	provider       startosis_packages.PackageContentProvider

	position       *kurtosis_instruction.InstructionPosition
	starlarkKwargs starlark.StringDict

	src          string
	artifactName string

	pathOnDisk string
}

func GenerateUploadFilesBuiltin(instructionsQueue *[]kurtosis_instruction.KurtosisInstruction, provider startosis_packages.PackageContentProvider, serviceNetwork service_network.ServiceNetwork) func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: Force returning an InterpretationError rather than a normal error
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		instructionPosition := shared_helpers.GetCallerPositionFromThread(thread)
		uploadInstruction := newEmptyUploadFilesInstruction(instructionPosition, serviceNetwork, provider)
		if interpretationError := uploadInstruction.parseStartosisArgs(b, args, kwargs); interpretationError != nil {
			return nil, interpretationError
		}
		*instructionsQueue = append(*instructionsQueue, uploadInstruction)
		return starlark.String(uploadInstruction.artifactName), nil
	}
}

func NewUploadFilesInstruction(position *kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_packages.PackageContentProvider, src string, pathOnDisk string, artifactId string, starlarkKwargs starlark.StringDict) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		src:            src,
		provider:       provider,
		pathOnDisk:     pathOnDisk,
		artifactName:   artifactId,
		starlarkKwargs: starlarkKwargs,
	}
}

func newEmptyUploadFilesInstruction(position *kurtosis_instruction.InstructionPosition, serviceNetwork service_network.ServiceNetwork, provider startosis_packages.PackageContentProvider) *UploadFilesInstruction {
	return &UploadFilesInstruction{
		position:       position,
		serviceNetwork: serviceNetwork,
		provider:       provider,
		src:            "",
		pathOnDisk:     "",
		artifactName:   "",
		starlarkKwargs: starlark.StringDict{},
	}
}

func (instruction *UploadFilesInstruction) GetPositionInOriginalScript() *kurtosis_instruction.InstructionPosition {
	return instruction.position
}

func (instruction *UploadFilesInstruction) GetCanonicalInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	args := []*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[srcArgName]), srcArgName, kurtosis_instruction.Representative),
		binding_constructors.NewStarlarkInstructionKwarg(shared_helpers.CanonicalizeArgValue(instruction.starlarkKwargs[artifactNameArgName]), artifactNameArgName, kurtosis_instruction.Representative),
	}
	return binding_constructors.NewStarlarkInstruction(instruction.position.ToAPIType(), UploadFilesBuiltinName, instruction.String(), args)
}

func (instruction *UploadFilesInstruction) Execute(_ context.Context) (*string, error) {
	compressedData, err := shared_utils.CompressPath(instruction.pathOnDisk, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while compressing the files '%v'", instruction.pathOnDisk)
	}
	filesArtifactUuid, err := instruction.serviceNetwork.UploadFilesArtifact(compressedData, instruction.artifactName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while uploading the compressed contents\n'%v'", compressedData)
	}
	instructionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", instruction.artifactName, filesArtifactUuid)
	return &instructionResult, nil
}

func (instruction *UploadFilesInstruction) String() string {
	return shared_helpers.CanonicalizeInstruction(UploadFilesBuiltinName, kurtosis_instruction.NoArgs, instruction.starlarkKwargs)
}

func (instruction *UploadFilesInstruction) ValidateAndUpdateEnvironment(environment *startosis_validator.ValidatorEnvironment) error {
	if environment.DoesArtifactNameExist(instruction.artifactName) {
		return stacktrace.NewError("There was an error validating '%v' as artifact name '%v' already exists", UploadFilesBuiltinName, instruction.artifactName)
	}
	environment.AddArtifactName(instruction.artifactName)
	return nil
}

func (instruction *UploadFilesInstruction) parseStartosisArgs(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) *startosis_errors.InterpretationError {
	var srcPathArg starlark.String
	var artifactNameArg starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, srcArgName, &srcPathArg, artifactNameArgName, &artifactNameArg); err != nil {
		return startosis_errors.WrapWithInterpretationError(err, "Failed parsing arguments for function '%s' (unparsed arguments were: '%v' '%v')", UploadFilesBuiltinName, args, kwargs)
	}

	instruction.starlarkKwargs[srcArgName] = srcPathArg
	instruction.starlarkKwargs[artifactNameArgName] = artifactNameArg
	instruction.starlarkKwargs.Freeze()

	srcPath, interpretationErr := kurtosis_instruction.ParseNonEmptyString(srcArgName, srcPathArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	artifactName, interpretationErr := kurtosis_instruction.ParseNonEmptyString(artifactNameArgName, artifactNameArg)
	if interpretationErr != nil {
		return interpretationErr
	}

	pathOnDisk, interpretationError := instruction.provider.GetOnDiskAbsoluteFilePath(srcPath)
	if interpretationError != nil {
		return interpretationError
	}

	instruction.src = srcPath
	instruction.artifactName = artifactName
	instruction.pathOnDisk = pathOnDisk
	return nil
}
