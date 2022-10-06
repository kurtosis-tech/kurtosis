package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisValidator struct {
	validators []startosis_validator.ValidatorInterface
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		validators: []startosis_validator.ValidatorInterface{
			dockerImagesValidator,
		},
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) error {
	environment := startosis_validator.NewValidatorEnvironment()
	for _, instruction := range instructions {
		instruction.UpdateValidationEnvironment(environment)
		err := validator.validateIntermediateEnvironment(ctx, environment)
		if err != nil {
			return stacktrace.Propagate(err, "Error while validating instruction %v", instruction.String())
		}
	}
	err := validator.validateFinalEnvironment(ctx, environment)
	if err != nil {
		return stacktrace.Propagate(err, "Error while validating script")
	}
	return nil
}

func (validator *StartosisValidator) validateIntermediateEnvironment(ctx context.Context, environment *startosis_validator.ValidatorEnvironment) error {
	for _, validator := range validator.validators {
		err := validator.ValidateDynamicEnvironment(ctx, environment)
		if err != nil {
			return stacktrace.Propagate(err, "Error while validating intermediate state of script")
		}
	}
	return nil
}

func (validator *StartosisValidator) validateFinalEnvironment(ctx context.Context, environment *startosis_validator.ValidatorEnvironment) error {
	for _, validator := range validator.validators {
		err := validator.ValidateStaticEnvironment(ctx, environment)
		if err != nil {
			return stacktrace.Propagate(err, "Error while validating final environment of script")
		}
	}
	return nil
}
