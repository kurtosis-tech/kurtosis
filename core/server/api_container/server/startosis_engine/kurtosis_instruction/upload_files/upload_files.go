package upload_files

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/utils"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
	"os"
)

const (
	UploadFilesBuiltinName = "upload_files"

	SrcArgName = "src"

	ArtifactNameArgName = "name"

	enforceMaxFileSizeLimit = false
	readOnlyFilePerm        = 0400
	descriptionFormatStr    = "Uploading file '%v' to files artifact '%v'"
)

func NewUploadFiles(
	packageId string,
	serviceNetwork service_network.ServiceNetwork,
	packageContentProvider startosis_packages.PackageContentProvider,
	packageReplaceOptions map[string]string,
) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: UploadFilesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              SrcArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, SrcArgName)
					},
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

				src:                   "",  // populated at interpretation time
				artifactName:          "",  // populated at interpretation time
				archivePathOnDisk:     "",  // populated at interpretation time
				filesArtifactMd5:      nil, // populated at interpretation time
				packageReplaceOptions: packageReplaceOptions,
				packageId:             packageId,
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

	src                   string
	artifactName          string
	archivePathOnDisk     string
	filesArtifactMd5      []byte
	packageReplaceOptions map[string]string
	packageId             string
	description           string
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

	absoluteLocator, interpretationErr := builtin.packageContentProvider.GetAbsoluteLocator(builtin.packageId, locatorOfModuleInWhichThisBuiltInIsBeingCalled, src.GoString(), builtin.packageReplaceOptions)
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Tried to convert locator '%v' into absolute locator but failed", src.GoString())
	}

	pathOnDisk, interpretationErr := builtin.packageContentProvider.GetOnDiskAbsolutePath(absoluteLocator)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	compressedDataPath, _, compressedDataMd5, err := utils.CompressPathToFile(pathOnDisk, enforceMaxFileSizeLimit)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "An error occurred while compressing the files at '%s'", pathOnDisk)
	}

	builtin.src = src.GoString()
	builtin.archivePathOnDisk = compressedDataPath
	builtin.filesArtifactMd5 = compressedDataMd5
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.src, builtin.artifactName))
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
	filesArtifactContentReader, err := os.OpenFile(builtin.archivePathOnDisk, os.O_RDONLY, readOnlyFilePerm)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred opening the files artifact archive at '%s'", builtin.archivePathOnDisk)
	}
	defer filesArtifactContentReader.Close()

	currentlyStoredFileArtifactUuid, currentlyStoredFileContentMd5, found, err := builtin.serviceNetwork.GetFilesArtifactMd5(builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred checking if the files artifact '%s' is already stored in this enclave", builtin.artifactName)
	}
	var instructionResult string
	if found && (len(builtin.filesArtifactMd5) > 0 && bytes.Equal(currentlyStoredFileContentMd5, builtin.filesArtifactMd5)) {
		instructionResult = fmt.Sprintf("Files with artifact name '%s' resolved with artifact UUID '%s' as content were matching", builtin.artifactName, currentlyStoredFileArtifactUuid)
	} else if found {
		err = builtin.serviceNetwork.UpdateFilesArtifact(currentlyStoredFileArtifactUuid, filesArtifactContentReader, builtin.filesArtifactMd5)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while updating the compressed contents with md5 '%s' to artifact '%s' (UUID: '%s')",
				builtin.filesArtifactMd5, builtin.artifactName, currentlyStoredFileArtifactUuid)
		}
		instructionResult = fmt.Sprintf("Files with artifact name '%s' with artifact UUID '%s' updated", builtin.artifactName, currentlyStoredFileArtifactUuid)
	} else {
		filesArtifactUuid, err := builtin.serviceNetwork.UploadFilesArtifact(filesArtifactContentReader, builtin.filesArtifactMd5, builtin.artifactName)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred while uploading the compressed contents with md5 '%s' to create new artifact with name '%s'",
				builtin.filesArtifactMd5, builtin.artifactName)
		}
		instructionResult = fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", builtin.artifactName, filesArtifactUuid)
	}
	return instructionResult, nil
}

func (builtin *UploadFilesCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	// if other instruction is nil or other instruction is not an add_service instruction, status is unknown
	if other == nil {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	if other.Type != UploadFilesBuiltinName {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// if artifact names don't match, status is unknown, instructions can't be resolved together
	if !other.HasOnlyFilesArtifactName(builtin.artifactName) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// If the artifact names are equal but the instructions are not equal, it's an update
	if !instructionsAreEqual {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	// From here the instructions are equal
	// If the hash of the files don't match, the instruction needs to be re-run
	if !other.HasOnlyFilesArtifactMd5(builtin.filesArtifactMd5) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}
	enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentWasLeftIntact)
	return enclave_structure.InstructionIsEqual
}

func (builtin *UploadFilesCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(
		UploadFilesBuiltinName,
	).AddFilesArtifact(
		builtin.artifactName, builtin.filesArtifactMd5,
	)
}

func (builtin *UploadFilesCapabilities) Description() string {
	return builtin.description
}
