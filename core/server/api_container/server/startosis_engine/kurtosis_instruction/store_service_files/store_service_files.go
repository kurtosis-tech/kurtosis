package store_service_files

import (
	"context"
	"fmt"
	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	"go.starlark.net/starlark"
)

const (
	StoreServiceFilesBuiltinName = "store_service_files"

	ServiceNameArgName  = "service_name"
	SrcArgName          = "src"
	ArtifactNameArgName = "name"
)

func NewStoreServiceFiles(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: StoreServiceFilesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameArgName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator:         nil,
				},
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
			return &StoreServiceFilesCapabilities{
				serviceNetwork: serviceNetwork,

				serviceName:  "", // populated at interpretation time
				src:          "", // populated at interpretation time
				artifactName: "", // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ServiceNameArgName:  true,
			SrcArgName:          true,
			ArtifactNameArgName: true,
		},
	}
}

type StoreServiceFilesCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	serviceName  kurtosis_backend_service.ServiceName
	src          string
	artifactName string
}

func (builtin *StoreServiceFilesCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
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

	serviceName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ServiceNameArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", ServiceNameArgName)
	}

	src, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, SrcArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", SrcArgName)
	}

	builtin.serviceName = kurtosis_backend_service.ServiceName(serviceName.GoString())
	builtin.src = src.GoString()
	return starlark.String(builtin.artifactName), nil
}

func (builtin *StoreServiceFilesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesServiceNameExist(builtin.serviceName) == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("There was an error validating '%v' with service name '%v' that does not exist", StoreServiceFilesBuiltinName, builtin.serviceName)
	}
	if validatorEnvironment.DoesArtifactNameExist(builtin.artifactName) == startosis_validator.ComponentCreatedOrUpdatedDuringPackageRun {
		return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' already exists", StoreServiceFilesBuiltinName, builtin.artifactName)
	}
	validatorEnvironment.AddArtifactName(builtin.artifactName)
	return nil
}

func (builtin *StoreServiceFilesCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	artifactUuid, err := builtin.serviceNetwork.CopyFilesFromService(ctx, string(builtin.serviceName), builtin.src, builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to copy file '%v' from service '%v", builtin.src, builtin.serviceName)
	}
	instructionResult := fmt.Sprintf("Files with artifact name '%s' uploaded with artifact UUID '%s'", builtin.artifactName, artifactUuid)
	return instructionResult, nil
}

func (builtin *StoreServiceFilesCapabilities) TryResolveWith(instructionsAreEqual bool, other kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if other == nil {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}
	otherStoreServiceFilesCapabilities, ok := other.(*StoreServiceFilesCapabilities)
	if !ok {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	if otherStoreServiceFilesCapabilities.artifactName != builtin.artifactName {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	if !instructionsAreEqual {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	if enclaveComponents.HasServiceBeenUpdated(builtin.serviceName) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentWasLeftIntact)
	return enclave_structure.InstructionIsEqual
}
