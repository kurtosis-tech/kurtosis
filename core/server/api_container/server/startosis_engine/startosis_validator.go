package startosis_engine

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/validator_state"
	"github.com/kurtosis-tech/stacktrace"
)

type StartosisValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidator {
	return &StartosisValidator{
		kurtosisBackend: kurtosisBackend,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) error {
	validatorState := validator_state.NewStartosisValidatorState(validator.kurtosisBackend)
	for _, instruction := range instructions {
		err := instruction.Validate(validatorState)
		if err != nil {
			return stacktrace.Propagate(err, "Error while validating instruction %v", instruction.String())
		}
	}
	err := validatorState.Validate(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Error while validating end state of script")
	}
	return nil
}
