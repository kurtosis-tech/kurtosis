package get_files_artifact

import (
	"context"
	"fmt"

	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/dependency_graph"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/plan_yaml"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/types"
	"go.starlark.net/starlark"
)

const (
	GetFilesArtifactBuiltinName = "get_files_artifact"
	FilesArtifactName           = "name"

	descriptionFormatStr = "Fetching files artifact '%v'"
)

func NewGetFilesArtifact() *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: GetFilesArtifactBuiltinName,
			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              FilesArtifactName,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, FilesArtifactName)
					},
				},
			},
			Deprecation: nil,
		},
		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &GetFilesArtifactCapabilities{
				artifactName: "", // populated at interpretation time
				description:  "", // populated at interpretation time
			}
		},
		DefaultDisplayArguments: map[string]bool{
			FilesArtifactName: true,
		},
	}
}

type GetFilesArtifactCapabilities struct {
	artifactName string
	description  string
}

func (builtin *GetFilesArtifactCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	artifactNameArgumentValue, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, FilesArtifactName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", FilesArtifactName)
	}
	artifactName := artifactNameArgumentValue.GoString()
	builtin.artifactName = artifactName
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.artifactName))

	// while this instruction simply returns what the input argument was, the returned starlark value can be used to set the files artifact elsewhere
	return starlark.String(artifactName), nil
}

func (builtin *GetFilesArtifactCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	// as long as the files artifact exists in the environment, this instruction will evaluate to the files artifact
	if exists := validatorEnvironment.DoesArtifactNameExist(builtin.artifactName); exists == startosis_validator.ComponentNotFound {
		return startosis_errors.NewValidationError("Files artifact '%v' required by '%v' instruction doesn't exist", builtin.artifactName, GetFilesArtifactBuiltinName)
	}
	return nil
}

func (builtin *GetFilesArtifactCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// Note this is a no-op
	return fmt.Sprintf("Fetched files artifact '%v'", builtin.artifactName), nil
}

func (builtin *GetFilesArtifactCapabilities) TryResolveWith(instructionsAreEqual bool, _ *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	if instructionsAreEqual && enclaveComponents.HasFilesArtifactBeenUpdated(builtin.artifactName) {
		return enclave_structure.InstructionIsUpdate
	} else if instructionsAreEqual {
		return enclave_structure.InstructionIsEqual
	}
	return enclave_structure.InstructionIsUnknown
}

func (builtin *GetFilesArtifactCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	builder.SetType(GetFilesArtifactBuiltinName).AddFilesArtifact(builtin.artifactName, nil)
}

func (builtin *GetFilesArtifactCapabilities) UpdatePlan(planYaml *plan_yaml.PlanYamlGenerator) error {
	// get files artifact does not affect the planYaml
	return nil
}

func (builtin *GetFilesArtifactCapabilities) Description() string {
	return builtin.description
}

// UpdateDependencyGraph updates the dependency graph with the effects of running this instruction.
func (builtin *GetFilesArtifactCapabilities) UpdateDependencyGraph(instructionUuid types.ScheduledInstructionUuid, dependencyGraph *dependency_graph.InstructionDependencyGraph) error {
	shortDescriptor := fmt.Sprintf("get_files_artifact(%s)", builtin.artifactName)
	dependencyGraph.UpdateInstructionShortDescriptor(instructionUuid, shortDescriptor)

	dependencyGraph.ConsumesFilesArtifact(instructionUuid, string(builtin.artifactName))
	dependencyGraph.ProducesFilesArtifact(instructionUuid, string(builtin.artifactName))
	return nil
}
