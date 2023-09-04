package time_module_with_now_disabled

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	TimeNowBuiltinName = "now"
	// UseRunPythonInsteadOfTimeNowError We want users to use plan.run_python while we come up with a deterministic alternative
	UseRunPythonInsteadOfTimeNowError = "The default `time.now()` function isn't supported please use `plan.run_python` while we work on supporting an alternative"
)

// GenerateTimeNowBuiltin This only exists to throw a nice interpretation error when print without plan is used
// This lives here instead of buiitins as we don't add it to kurtosis_module.go
func GenerateTimeNowBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return nil, startosis_errors.NewInterpretationError(UseRunPythonInsteadOfTimeNowError)
	}
}
