package time_module_with_now_disabled

import (
	"go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func GetTimeModuleWithNowDisabled() *starlarkstruct.Module {
	time.Module.Members[TimeNowBuiltinName] = starlark.NewBuiltin(TimeNowBuiltinName, GenerateTimeNowBuiltin())
	return time.Module
}
