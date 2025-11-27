package time_now_builtin

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	TimeNowBuiltinName = "now"
	// UseRunPythonInsteadOfTimeNowError We want users to use plan.run_python while we come up with a deterministic alternative
	UseRunPythonInsteadOfTimeNowError = "The default `time.now()` function isn't supported please use `plan.run_python` while we work on supporting an alternative"
)

// GenerateTimeNowBuiltin This only exists to throw a nice interpretation error when print without plan is used
func GenerateTimeNowBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return nil, startosis_errors.NewInterpretationError(UseRunPythonInsteadOfTimeNowError)
	}
}
