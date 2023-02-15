package upload_files

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	SrcArgName = "src"

	ArtifactNameArgName = "name"

	ensureCompressedFileIsLesserThanGRPCLimit = false
)

func NewUploadFiles(serviceNetwork service_network.ServiceNetwork, packageContentProvider startosis_packages.PackageContentProvider) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UploadFilesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SrcArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
				{
					Name:              ArtifactNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &UploadFilesCapabilities{
				serviceNetwork:         serviceNetwork,
				packageContentProvider: packageContentProvider,

				src:          "", // populated at interpretation time
				artifactName: "", // populated at interpretation time
				pathOnDisk:   "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			SrcArgName:          true,
			ArtifactNameArgName: true,
		},
	}
}

type UploadFilesCapabilities struct {
	serviceNetwork         service_network.ServiceNetwork
	packageContentProvider startosis_packages.PackageContentProvider

	src          string
	artifactName string
	pathOnDisk   string
}

func (builtin *UploadFilesCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	src, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, SrcArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SrcArgName)
	}

	artifactName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ArtifactNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ArtifactNameArgName)
	}

	pathOnDisk, interpretationErr := builtin.packageContentProvider.GetOnDiskAbsoluteFilePath(src.GoString())
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	builtin.artifactName = artifactName.GoString()
	builtin.src = src.GoString()
	builtin.pathOnDisk = pathOnDisk
	return starlark.String(builtin.artifactName), nil
}

func (builtin *UploadFilesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesArtifactNameExist(builtin.artifactName) {
		return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' already exists", UploadFilesBuiltinName, builtin.artifactName)
	}
	validatorEnvironment.AddArtifactName(builtin.artifactName)
	return nil
}

func (builtin *UploadFilesCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	compressedData, err := shared_utils.CompressPath(builtin.pathOnDisk, ensureCompressedFileIsLesserThanGRPCLimit)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while compressing the files '%v'", builtin.pathOnDisk)
	}
	filesArtifactUuid, err := builtin.serviceNetwork.UploadFilesArtifact(compressedData, builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred while uploading the compressed contents\n'%v'", compressedData)
	}
	instructionResult := fmt.Sprintf("Files  with artifact name '%s' uploaded with artifact UUID '%s'", builtin.artifactName, filesArtifactUuid)
	return instructionResult, nil
}
