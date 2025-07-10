package store_service_files

import (
	"context"
	"fmt"

	kurtosis_backend_service "github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
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
	DependsOnArgName    = "depends_on"
)

const (
	descriptionFormatStr = "Storing files from service '%v' at path '%v' to files artifact with name '%v'"
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
				{
					Name:              DependsOnArgName,
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
				dependsOn:    "", // populated at interpretation time
				description:  "", // populated at interpretation time
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
	description  string
	dependsOn    string
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

	var dependsOn starlark.String
	if arguments.IsSet(ArtifactNameArgName) {
		dependsOn, err = builtin_argument.ExtractArgumentValue[starlark.String](arguments, DependsOnArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", DependsOnArgName)
		}
	}

	builtin.serviceName = kurtosis_backend_service.ServiceName(serviceName.GoString())
	builtin.src = src.GoString()
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.serviceName, builtin.src, builtin.artifactName))
	builtin.dependsOn = dependsOn.GoString()
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

func (builtin *StoreServiceFilesCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	// if other instruction is nil or other instruction is not an add_service instruction, status is unknown
	if other == nil {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}
	if other.Type != StoreServiceFilesBuiltinName {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// if artifact names don't match, status is unknown, instructions can't be resolved together
	if !other.HasOnlyFilesArtifactName(builtin.artifactName) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// If artifact names are the same but instructions are not equal, it needs to be re-run anyway
	if !instructionsAreEqual {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	// From here the instructions are equal
	// If the service has been updated, the instruction needs to be re-run since the file content might have changed
	if enclaveComponents.HasServiceBeenUpdated(builtin.serviceName) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}

	enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentWasLeftIntact)
	return enclave_structure.InstructionIsEqual
}

func (builtin *StoreServiceFilesCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	// No need for the MD5 here because "store_service_files" is an atomic operation at the service_network level.
	// Here we just consider that if the service has been updated, we store the file again (b/c it the content might
	// have changed), otherwise we don't
	builder.SetType(
		StoreServiceFilesBuiltinName,
	).AddFilesArtifact(
		builtin.artifactName, nil,
	)
}

func (builtin *StoreServiceFilesCapabilities) UpdatePlan(plan *plan_yaml.PlanYamlGenerator) error {
	err := plan.AddStoreServiceFiles(builtin.artifactName, builtin.src)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating plan with store service files")
	}
	return nil
}

func (builtin *StoreServiceFilesCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *StoreServiceFilesCapabilities) UpdateDependencyGraph(instructionUuid dependency_graph.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionsDependencyGraph) error {
	dependencyGraph.StoreOutput(instructionUuid, string(builtin.artifactName))

	dependencyGraph.DependsOnOutput(instructionUuid, string(builtin.serviceName))
	dependencyGraph.AddInstructionShortDescriptor(instructionUuid, fmt.Sprintf("store_service_files(%s, %s)", builtin.serviceName, builtin.description))

	if builtin.dependsOn != "" {
		dependencyGraph.DependsOnOutput(instructionUuid, builtin.dependsOn)
	}

	return nil
}
