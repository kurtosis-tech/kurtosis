package render_templates

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network/render_templates"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_plan_persistence"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/enclave_structure"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/runtime_value_store"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"reflect"
)

const (
	RenderTemplatesBuiltinName = "render_templates"

	TemplateAndDataByDestinationRelFilepathArg = "config"
	ArtifactNameArgName                        = "name"

	templatesAndDataArgName = "config"
	templateFieldKey        = "template"
	templateDataFieldKey    = "data"
	jsonParsingThreadName   = "Unused thread name"
	jsonParsingModuleId     = "Unused module id"
	descriptionFormatStr    = "Rendering a template to a files artifact with name '%v'"
)

func NewRenderTemplatesInstruction(serviceNetwork service_network.ServiceNetwork, runtimeValueStore *runtime_value_store.RuntimeValueStore) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RenderTemplatesBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              TemplateAndDataByDestinationRelFilepathArg,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
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
			return &RenderTemplatesCapabilities{
				serviceNetwork:    serviceNetwork,
				runtimeValueStore: runtimeValueStore,

				artifactName:                      "",  // will be populated at interpretation time
				templatesAndDataByDestRelFilepath: nil, // will be populated at interpretation time
				description:                       "",  // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			ArtifactNameArgName: true,
		},
	}
}

type RenderTemplatesCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	artifactName                      string
	templatesAndDataByDestRelFilepath map[string]*render_templates.TemplateData

	runtimeValueStore *runtime_value_store.RuntimeValueStore

	description string
}

func (builtin *RenderTemplatesCapabilities) Interpret(_ string, arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	if !arguments.IsSet(ArtifactNameArgName) {
		natureThemeName, err := builtin.serviceNetwork.GetUniqueNameForFileArtifact()
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to auto generate name '%s' argument", ArtifactNameArgName)
		}
		builtin.artifactName = natureThemeName
	} else {
		artifactName, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, ArtifactNameArgName)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%s'", ArtifactNameArgName)
		}
		builtin.artifactName = artifactName.GoString()
	}

	config, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, TemplateAndDataByDestinationRelFilepathArg)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to parse '%s'", TemplateAndDataByDestinationRelFilepathArg)
	}
	templatesAndDataByDestRelFilepath, interpretationErr := parseTemplatesAndData(config)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	builtin.templatesAndDataByDestRelFilepath = templatesAndDataByDestRelFilepath
	builtin.description = builtin_argument.GetDescriptionOrFallBack(arguments, fmt.Sprintf(descriptionFormatStr, builtin.artifactName))
	return starlark.String(builtin.artifactName), nil
}

func (builtin *RenderTemplatesCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	if validatorEnvironment.DoesArtifactNameExist(builtin.artifactName) == startosis_validator.ComponentCreatedOrUpdatedDuringPackageRun {
		return startosis_errors.NewValidationError("There was an error validating '%v' as artifact name '%v' already exists", RenderTemplatesBuiltinName, builtin.artifactName)
	}
	validatorEnvironment.AddArtifactName(builtin.artifactName)
	return nil
}

func (builtin *RenderTemplatesCapabilities) Execute(_ context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	for _, templateData := range builtin.templatesAndDataByDestRelFilepath {
		if err := templateData.ReplaceRuntimeValues(builtin.runtimeValueStore); err != nil {
			return "", stacktrace.Propagate(err, "An error occurred replacing runtime values for render_template instruction")
		}
	}

	artifactUUID, err := builtin.serviceNetwork.RenderTemplates(builtin.templatesAndDataByDestRelFilepath, builtin.artifactName)
	if err != nil {
		return "", stacktrace.Propagate(err, "Failed to render templates '%v'", builtin.templatesAndDataByDestRelFilepath)
	}
	instructionResult := fmt.Sprintf("Templates artifact name '%s' rendered with artifact UUID '%s'", builtin.artifactName, artifactUUID)
	return instructionResult, nil
}

func (builtin *RenderTemplatesCapabilities) TryResolveWith(instructionsAreEqual bool, other *enclave_plan_persistence.EnclavePlanInstruction, enclaveComponents *enclave_structure.EnclaveComponents) enclave_structure.InstructionResolutionStatus {
	// if other instruction is nil or other instruction is not an add_service instruction, status is unknown
	if other == nil {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}
	if other.Type != RenderTemplatesBuiltinName {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// if artifact names don't match, status is unknown, instructions can't be resolved together
	if !other.HasOnlyFilesArtifactName(builtin.artifactName) {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsNew)
		return enclave_structure.InstructionIsUnknown
	}

	// just check for instruction equality. If it's not equal it needs to be rerun
	if !instructionsAreEqual {
		enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentIsUpdated)
		return enclave_structure.InstructionIsUpdate
	}
	enclaveComponents.AddFilesArtifact(builtin.artifactName, enclave_structure.ComponentWasLeftIntact)
	return enclave_structure.InstructionIsEqual
}

func (builtin *RenderTemplatesCapabilities) FillPersistableAttributes(builder *enclave_plan_persistence.EnclavePlanInstructionBuilder) {
	// technically, we need the MD5 of the files artifact here but because the template is passed in plaintext to the
	// instruction the check for instruction equality also checks that the content of the artifact is identical.
	// So we can safely ignore the MD5
	builder.SetType(
		RenderTemplatesBuiltinName,
	).AddFilesArtifact(
		builtin.artifactName, nil,
	)
}

func (builtin *RenderTemplatesCapabilities) Description() string {
	return builtin.description
}

func parseTemplatesAndData(templatesAndData *starlark.Dict) (map[string]*render_templates.TemplateData, *startosis_errors.InterpretationError) {
	templateAndDataByDestRelFilepath := make(map[string]*render_templates.TemplateData)
	for _, relPathInFilesArtifactKey := range templatesAndData.Keys() {
		relPathInFilesArtifactStr, castErr := kurtosis_types.SafeCastToString(relPathInFilesArtifactKey, fmt.Sprintf("%v.key:%v", templatesAndDataArgName, relPathInFilesArtifactKey))
		if castErr != nil {
			return nil, castErr
		}
		value, found, dictErr := templatesAndData.Get(relPathInFilesArtifactKey)
		if !found || dictErr != nil {
			return nil, startosis_errors.NewInterpretationError("'%s' key in dict '%s' doesn't have a value we could retrieve. This is a Kurtosis bug.", relPathInFilesArtifactKey.String(), templatesAndDataArgName)
		}
		structValue, ok := value.(*starlarkstruct.Struct)
		if !ok {
			return nil, startosis_errors.NewInterpretationError("Expected %v[\"%v\"] to be a dict. Got '%s'", templatesAndData, relPathInFilesArtifactStr, reflect.TypeOf(value))
		}
		template, err := structValue.Attr(templateFieldKey)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateFieldKey)
		}
		templateStr, castErr := kurtosis_types.SafeCastToString(template, fmt.Sprintf("%v[\"%v\"][\"%v\"]", templatesAndDataArgName, relPathInFilesArtifactStr, templateFieldKey))
		if castErr != nil {
			return nil, castErr
		}
		templateDataStarlarkValue, err := structValue.Attr(templateDataFieldKey)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Expected values in '%v' to have a '%v' field", templatesAndDataArgName, templateDataFieldKey)
		}

		templateDataJSONStrValue, encodingError := encodeStarlarkObjectAsJSON(templateDataStarlarkValue, templateDataFieldKey)
		if encodingError != nil {
			return nil, encodingError
		}
		// Massive Hack
		// We do this for a couple of reasons,
		// 1. Unmarshalling followed by Marshalling, allows for the non-scientific notation of floats to be preserved
		// 2. Don't have to write a custom way to jsonify Starlark
		// 3. This behaves as close to marshalling primitives in Golang as possible
		// 4. Allows us to validate that string input is valid JSON
		var temporaryUnmarshalledValue interface{}
		err = json.Unmarshal([]byte(templateDataJSONStrValue), &temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Template data for file '%v', '%v' isn't valid JSON", relPathInFilesArtifactStr, templateDataJSONStrValue)
		}
		templateDataJson, err := json.Marshal(temporaryUnmarshalledValue)
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("Template data for file '%v', '%v' isn't valid JSON", relPathInFilesArtifactStr, templateDataJSONStrValue)
		}
		// end Massive Hack
		templateAndData, err := render_templates.CreateTemplateData(templateStr, string(templateDataJson))
		if err != nil {
			return nil, startosis_errors.NewInterpretationError("An error occurred creating the template. Make sure the it is a valid template string.")
		}
		templateAndDataByDestRelFilepath[relPathInFilesArtifactStr] = templateAndData
	}
	return templateAndDataByDestRelFilepath, nil
}

func encodeStarlarkObjectAsJSON(object starlark.Value, argNameForLogging string) (string, *startosis_errors.InterpretationError) {
	jsonifiedVersion := ""
	thread := &starlark.Thread{
		Name:       jsonParsingThreadName,
		OnMaxSteps: nil,
		Print: func(_ *starlark.Thread, msg string) {
			jsonifiedVersion = msg
		},
		Load:  nil,
		Steps: 0,
	}

	predeclared := &starlark.StringDict{
		// go-starlark add-ons
		starlarkjson.Module.Name:          starlarkjson.Module,
		starlarkstruct.Default.GoString(): starlark.NewBuiltin(starlarkstruct.Default.GoString(), starlarkstruct.Make), // extension to build struct in starlark
	}

	// We do a print here as if we return the encoded variable we get extra quotes and slashes
	// {"fizz": "buzz"} becomes "{\"fizz": \"buzz"\}"
	scriptToRun := fmt.Sprintf(`encoded_json = json.encode(%v)
print(encoded_json)`, object.String())

	_, err := starlark.ExecFile(thread, jsonParsingModuleId, scriptToRun, *predeclared)

	if err != nil {
		return "", startosis_errors.NewInterpretationError("Error converting '%v' with string value '%v' to JSON", argNameForLogging, object.String())
	}

	return jsonifiedVersion, nil
}
