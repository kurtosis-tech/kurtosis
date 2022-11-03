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
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		dockerImagesValidator,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, serviceNetwork service_network.ServiceNetwork, instructions []kurtosis_instruction.KurtosisInstruction) []*kurtosis_core_rpc_api_bindings.StartosisValidationError {
	environment := startosis_validator.NewValidatorEnvironment(map[string]bool{}, serviceNetwork.GetServiceIDs())
	for _, instruction := range instructions {
		err := instruction.ValidateAndUpdateEnvironment(environment)
		if err != nil {
			return []*kurtosis_core_rpc_api_bindings.StartosisValidationError{
				binding_constructors.NewStartosisValidationError(stacktrace.Propagate(err, "Error while validating instruction %v. The instruction can be found at %v", instruction.String(), instruction.GetPositionInOriginalScript().String()).Error()),
			}
		}
	}
	errors := validator.dockerImagesValidator.Validate(ctx, environment)
	if errors != nil {
		propagatedErrors := []*kurtosis_core_rpc_api_bindings.StartosisValidationError{}
		for _, err := range errors {
			propagatedErrors = append(propagatedErrors, binding_constructors.NewStartosisValidationError(stacktrace.Propagate(err, "Error while validating final environment of script").Error()))
		}
		return propagatedErrors
	}
	return nil
}
