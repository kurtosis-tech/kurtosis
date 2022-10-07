package startosis_validator

import "context"

type Validator interface {
	// Validate validates if an environment is valid after all instructions have been executed
	// and no further mutations will be performed to environment.
	Validate(ctx context.Context, environment *ValidatorEnvironment) error
}
