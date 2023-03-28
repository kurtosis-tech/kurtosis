package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
)

type KurtosisTypeConstructorBaseTest interface {
	// GetId is a unique identifier for this test that will be used in errors when a test fails.
	GetId() string

	// GetHelper should return the helper this test is testing
	GetTypeConstructor() *kurtosis_type_constructor.KurtosisTypeConstructor

	// GetStarlarkCode should return the Starlark code corresponding to the type constructor being tested.
	GetStarlarkCode() string

	// Assert is called after the Starlark code returned by GetStarlarkCode has been sent to the interpreter
	// The typeValue argument corresponds to the value that was instantiated based on the starlark code provided
	Assert(typeValue builtin_argument.KurtosisValueType)
}
