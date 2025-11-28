package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/builtins/time_now_builtin"
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func TimeModuleWithNowDisabled() *starlarkstruct.Module {
	time.Module.Members[time_now_builtin.TimeNowBuiltinName] = starlark.NewBuiltin(time_now_builtin.TimeNowBuiltinName, time_now_builtin.GenerateTimeNowBuiltin())
	return time.Module
}
