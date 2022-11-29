package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisValidator struct {
	dockerImagesValidator *startosis_validator.DockerImagesValidator

	serviceNetwork service_network.ServiceNetwork
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend, serviceNetwork service_network.ServiceNetwork) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		dockerImagesValidator,
		serviceNetwork,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	starlarkExecutionResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer close(starlarkExecutionResponseLineStream)
		environment := startosis_validator.NewValidatorEnvironment(validator.serviceNetwork.GetServiceIDs())
		for _, instruction := range instructions {
			err := instruction.ValidateAndUpdateEnvironment(environment)
			if err != nil {
				propagatedError := stacktrace.Propagate(err, "Error while validating instruction %v. The instruction can be found at %v", instruction.String(), instruction.GetPositionInOriginalScript().String())
				serializedError := binding_constructors.NewStarlarkValidationError(propagatedError.Error())
				starlarkExecutionResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(serializedError)
			}
		}
		errors := validator.dockerImagesValidator.Validate(ctx, environment)
		for _, err := range errors {
			propagatedError := stacktrace.Propagate(err, "Error while validating final environment of script")
			serializedError := binding_constructors.NewStarlarkValidationError(propagatedError.Error())
			starlarkExecutionResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(serializedError)
		}
	}()
	return starlarkExecutionResponseLineStream
}
