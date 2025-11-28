package test_engine

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_helper"
	"go.starlark.net/starlark"
)

type KurtosisHelperBaseTest interface {
	// GetHelper should return the helper this test is testing
	GetHelper() *kurtosis_helper.KurtosisHelper

	// GetStarlarkCode should return the Starlark code corresponding to the instruction being tested.
	GetStarlarkCode() string

	// Assert is called after the Starlark code returned by GetStarlarkCode has been sent to the interpreter
	// The result argument corresponds to the result of the builtin, if any
	Assert(result starlark.Value)
}
