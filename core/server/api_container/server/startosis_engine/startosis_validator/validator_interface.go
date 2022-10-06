package startosis_validator

import "context"

type ValidatorInterface interface {
	// ValidateDynamicEnvironment validates if an environment is valid after the execution of a set of instructions.
	ValidateDynamicEnvironment(ctx context.Context, environment *ValidatorEnvironment) error
	// ValidateStaticEnvironment validates if an environment is valid after all instructions have been executed
	// and no further mutations will be performed to environment.
	ValidateStaticEnvironment(ctx context.Context, environment *ValidatorEnvironment) error
}
