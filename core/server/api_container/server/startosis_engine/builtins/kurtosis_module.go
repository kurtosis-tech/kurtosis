package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	KurtosisModuleName = "kurtosis"

	connectionSubmoduleName = "connection"
)

func KurtosisModule() *starlarkstruct.Module {
	return &starlarkstruct.Module{
		Name: KurtosisModuleName,
		Members: starlark.StringDict{
			connectionSubmoduleName: kurtosis_types.PreBuiltConnectionConfigs,
		},
	}
}
