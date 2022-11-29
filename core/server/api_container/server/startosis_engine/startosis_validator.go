package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
)

const (
	causedByPrefix = "\tCaused by: "
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

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) <-chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine {
	kurtosisExecutionResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.KurtosisExecutionResponseLine)
	go func() {
		defer close(kurtosisExecutionResponseLineStream)
		environment := startosis_validator.NewValidatorEnvironment(validator.serviceNetwork.GetServiceIDs())
		for _, instruction := range instructions {
			err := instruction.ValidateAndUpdateEnvironment(environment)
			if err != nil {
				// this is intentionally not using stacktrace.Propagate, as we don't want to pollute the error with Go line, column numbers
				indentedError := fmt.Errorf("Error while validating instruction %v. The instruction can be found at %v\n%s%s", instruction.String(), instruction.GetPositionInOriginalScript().String(), causedByPrefix, err.Error())
				serializedError := binding_constructors.NewKurtosisValidationError(indentedError.Error())
				kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromValidationError(serializedError)
			}
		}
		errors := validator.dockerImagesValidator.Validate(ctx, environment)
		for _, err := range errors {
			// this is intentionally not using stacktrace.Propagate, as we don't want to pollute the error with Go line, column numbers
			indentedError := fmt.Sprintf("Error while validating final environment of script\n%s%s", causedByPrefix, err.Error())
			serializedError := binding_constructors.NewKurtosisValidationError(indentedError)
			kurtosisExecutionResponseLineStream <- binding_constructors.NewKurtosisExecutionResponseLineFromValidationError(serializedError)
		}
	}()
	return kurtosisExecutionResponseLineStream
}
