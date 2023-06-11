package run_sh

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_plan_instruction"
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
			},
		},

		Capabilities: func() kurtosis_plan_instruction.KurtosisPlanInstructionCapabilities {
			return &RunShCapabilities{
				serviceNetwork: serviceNetwork,

				image:   "badouralix/curl-jq", // populated at interpretation time
				run:     "",                   // populated at interpretation time
				workdir: "",                   // populated at interpretation time
			}
		},

		DefaultDisplayArguments: map[string]bool{
			RunArgName:       true,
			ImageNameArgName: true,
			WorkDirArgName:   true,
		},
	}
}

type RunShCapabilities struct {
	serviceNetwork service_network.ServiceNetwork

	run     string
	image   string
	workdir string
}

func (builtin *RunShCapabilities) Interpret(arguments *builtin_argument.ArgumentValuesSet) (starlark.Value, *startosis_errors.InterpretationError) {
	runCommand, err := builtin_argument.ExtractArgumentValue[starlark.String](arguments, RunArgName)
	if err != nil {
		return nil, startosis_errors.WrapWithInterpretationError(err, "Unable to extract value for '%s' argument", RunArgName)
	}

	builtin.run = runCommand.GoString()
	return starlark.None, nil
}

func (builtin *RunShCapabilities) Validate(_ *builtin_argument.ArgumentValuesSet, validatorEnvironment *startosis_validator.ValidatorEnvironment) *startosis_errors.ValidationError {
	return nil
}

func (builtin *RunShCapabilities) Execute(ctx context.Context, _ *builtin_argument.ArgumentValuesSet) (string, error) {
	// create default workdir
	createDefaultDirectory := []string{"/bin/sh", "-c", "mkdir -p task && cd task && echo $(pwd)"}
	serviceConfigBuilder := services.NewServiceConfigBuilder(builtin.image)
	// serviceConfigBuilder.WithEntryPointArgs(entryPoint)
	// serviceConfigBuilder.WithCmdArgs(entryPoint)

	serviceConfig := serviceConfigBuilder.Build()
	_, err := builtin.serviceNetwork.AddService(ctx, "task", serviceConfig)

	if err != nil {
		logrus.Errorf("some error happened %v", err)
	}

	code, output, err := builtin.serviceNetwork.ExecCommand(ctx, "task", createDefaultDirectory)
	if err != nil {
		logrus.Errorf("some error happened %v", err)
	}

	starlark_warning.PrintOnceAtTheEndOfExecutionf("Ouput Code: %v , Output: %v", code, output)
	return "", err
}
