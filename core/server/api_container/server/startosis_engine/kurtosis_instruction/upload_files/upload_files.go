package upload_files

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/shared_utils"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	UploadFilesBuiltinName = "upload_files"

	SrcArgName = "src"

	ArtifactNameArgName = "name"

	enforceMaxFileSizeLimit = false
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
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &UploadFilesCapabilities{
				serviceNetwork:         serviceNetwork,
				packageContentProvider: packageContentProvider,

				src:                     "",  // populated at interpretation time
				artifactName:            "",  // populated at interpretation time
				filesArtifactContent:    nil, // populated at interpretation time
				filesArtifactContentMd5: nil, // populated at interpretation time
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

	src                     string
	artifactName            string
	filesArtifactContent    []byte // WTF why aren't we using a ready here?
	filesArtifactContentMd5 []byte
}

func (builtin *UploadFilesCapabilities) Interpret(locatorOfModuleInWhichThisBuiltInIsBeingCalled string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	if !arguments.IsSet(ArtifactNameArgName) {
		natureThemeName, err := builtin.serviceNetwork.GetUniqueNameForFileArtifact()
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to auto generate name '%s' argument", ArtifactNameArgName)
		}
		builtin.artifactName = natureThemeName
	} else {
		artifactName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ArtifactNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ArtifactNameArgName)
		}
		builtin.artifactName = artifactName.GoString()
	}

	src, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, SrcArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SrcArgName)
	}

	absoluteLocator, interpretationErr := builtin.packageContentProvider.GetAbsoluteLocatorForRelativeModuleLocator(locatorOfModuleInWhichThisBuiltInIsBeingCalled, src.GoString())
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Tried to convert locator '%v' into absolute locator but failed", src.GoString())
	}

	pathOnDisk, interpretationErr := builtin.packageContentProvider.GetOnDiskAbsoluteFilePath(absoluteLocator)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	compressedData, contentMd5, err := shared_utils.CompressPath(pathOnDisk, enforceMaxFileSizeLimit)
	if err != nil {
		return nil, startosis_errors.NewInterpretationError("An error occurred while compressing the files '%v'", pathOnDisk)
	}

	builtin.src = src.GoString()
	builtin.filesArtifactContent = compressedData
	builtin.filesArtifactContentMd5 = contentMd5

	return starlark.String(builtin.artifactName), nil
}

func (builtin *UploadFilesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesArtifactNameExist(builtin.artifactName) == startosis_validator.ComponentCreatedOrUpdatedDuringPackageRun {
		return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' already exists", UploadFilesBuiltinName, builtin.artifactName)
	}
	validatorEnvironment.AddArtifactName(builtin.artifactName)
	return nil
}

func (builtin *UploadFilesCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	currentlyStoredFileArtifactUuid, currentlyStoredFileContentMd5, found, err := builtin.serviceNetwork.GetFilesArtifactMd5(builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking if a file already exists for '%s' in the store", builtin.artifactName)
	}

	logrus.Debugf("Currently stored file md5: '%s', new file md5: '%s'", string(currentlyStoredFileContentMd5), string(builtin.filesArtifactContentMd5))

	var instructionResult string
	if found && bytes.Equal(currentlyStoredFileContentMd5, builtin.filesArtifactContentMd5) {
		instructionResult = fmt.Sprintf("Files with artifact name '%s' resolved with artifact UUID '%s' as content were matching", builtin.artifactName, currentlyStoredFileArtifactUuid)
	} else if found {
		filesArtifactUuid, err := builtin.serviceNetwork.UpdateFilesArtifact(builtin.filesArtifactContent, builtin.filesArtifactContentMd5, builtin.artifactName)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while uploading the compressed contents with md5 '%s' to artifact '%s' (UUID: '%s')",
				builtin.filesArtifactContentMd5, builtin.artifactName, filesArtifactUuid)
		}
		instructionResult = fmt.Sprintf("Files with artifact name '%s' with artifact UUID '%s' updated", builtin.artifactName, filesArtifactUuid)
	} else {
		filesArtifactUuid, err := builtin.serviceNetwork.UploadFilesArtifact(builtin.filesArtifactContent, builtin.filesArtifactContentMd5, builtin.artifactName)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while uploading the compressed contents with md5 '%s' to create new artifact with name '%s'",
				builtin.filesArtifactContentMd5, builtin.artifactName)
		}
		instructionResult = fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", builtin.artifactName, filesArtifactUuid)
	}
	return instructionResult, nil
}

func (builtin *UploadFilesCapabilities) TryResolveWith(instructionsAreEqual bool, other kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if other == nil {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	otherUploadFilesCapabilities, ok := other.(*UploadFilesCapabilities)
	if !ok {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	if !instructionsAreEqual {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	logrus.Debugf("other instruction md5: '%s' - new instruction md5: '%s'", otherUploadFilesCapabilities.filesArtifactContentMd5, builtin.filesArtifactContentMd5)
	if instructionsAreEqual && bytes.Equal(otherUploadFilesCapabilities.filesArtifactContentMd5, builtin.filesArtifactContentMd5) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentWasLeftIntact)
		return enclave_structure.InstructionIsEqual
	} else {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}
}
