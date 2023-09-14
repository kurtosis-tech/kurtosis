package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
)

type KurtosisTypeConstructorBaseTest interface {
	// GetStarlarkCode should return the Starlark code corresponding to the type constructor being tested.
	GetStarlarkCode() string

	// Assert is called after the Starlark code returned by GetStarlarkCode has been sent to the interpreter
	// The typeValue argument corresponds to the value that was instantiated based on the starlark code provided
	Assert(typeValue builtin_argument.KurtosisValueType)
}
