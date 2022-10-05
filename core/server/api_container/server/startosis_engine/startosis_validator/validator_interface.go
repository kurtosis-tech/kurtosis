package startosis_validator

import "context"

type ValidatorInterface interface {
	ValidateIntermediateEnvironment(ctx context.Context, environment *ValidatorEnvironment) error
	ValidateFinalEnvironment(ctx context.Context, environment *ValidatorEnvironment) error
}
