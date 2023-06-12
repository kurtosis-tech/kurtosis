package run_sh

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

const (
	RunShBuiltinName = "run_sh"

	ImageNameArgName = "image"
	WorkDirArgName   = "workdir"
	RunArgName       = "run"

	defaultWorkDir = "task"
	FilesAttr      = "files"
)

func NewRunShService(serviceNetwork service_network.ServiceNetwork) *kurtosis_plan_instruction.KurtosisPlanInstruction {
	return &kurtosis_plan_instruction.KurtosisPlanInstruction{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: RunShBuiltinName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              RunArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              ImageNameArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              WorkDirArgName,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
				},
				{
					Name:              FilesAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
				},
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork: serviceNetwork,

				image:   "badouralix/curl-jq", // populated at interpretation time
				run:     "",                   // populated at interpretation time
				workdir: "",                   // populated at interpretation time
				files:   nil,
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RunArgName:       true,
			ImageNameArgName: true,
			WorkDirArgName:   true,
			FilesAttr:        true,
		},
	}
}

type RunShCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	run     string
	image   string
	workdir string
	files   map[string]string
}

func (builtin *RunShCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	runCommand, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}

	builtin.run = runCommand.GoString()

	if arguments.IsSet(FilesAttr) {
		filesStarlark, err := builtin_argument.ExtractArgumentValue[*starlark.Dict](arguments, FilesAttr)
		if err != nil {
			return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", FilesAttr)
		}
		if filesStarlark.Len() > 0 {
			filesArtifactMountDirpaths, interpretationErr := kurtosis_types.SafeCastToMapStringString(filesStarlark, FilesAttr)
			if interpretationErr != nil {
				return nil, interpretationErr
			}
			builtin.files = filesArtifactMountDirpaths
		}
	}

	return starlark.None, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// cmd string for default workdir
	createAndSwitchTheDirectory := fmt.Sprintf("mkdir -p %v && cd %v", defaultWorkDir, defaultWorkDir)
	completeRunCommand := fmt.Sprintf("%v && %v", createAndSwitchTheDirectory, builtin.run)

	createDefaultDirectory := []string{"/bin/sh", "-c", completeRunCommand}
	serviceConfigBuilder := services.NewServiceConfigBuilder(builtin.image)
	serviceConfigBuilder.WithFilesArtifactMountDirpaths(builtin.files)

	serviceConfig := serviceConfigBuilder.Build()
	_, err := builtin.serviceNetwork.AddService(ctx, "task", serviceConfig)

	if err != nil {
		logrus.Errorf("some error happened %v", err)
	}

	//starlark_warning.PrintOnceAtTheEndOfExecutionf("Cmd %+v", createDefaultDirectory)
	code, output, err := builtin.serviceNetwork.ExecCommand(ctx, "task", createDefaultDirectory)
	if err != nil {
		logrus.Errorf("some error happened %v", err)
	}

	starlark_warning.PrintOnceAtTheEndOfExecutionf("Ouput Code: %v , Output: %v", code, output)
	return "", err
}
