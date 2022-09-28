package kurtosis_instruction

import "context"

type KurtosisInstruction interface {
	GetCanonicalInstruction() string

	Execute(ctx context.Context) error

	// String is only for easy printing in logs and error messages.
	// Most of the time it will just call GetCanonicalInstruction()
	String() string
}
