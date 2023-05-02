package kurtosis_starlark_framework

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/starlark_warning"
	"go.starlark.net/starlark"
)

// KurtosisBaseBuiltin is the object all Kurtosis builtin should be composed of.
//
// It includes the mandatory metadata for each builtin, in particular its name and the structure of its arguments
type KurtosisBaseBuiltin struct {
	Name string

	Arguments []*builtin_argument.BuiltinArgument

	Deprecation *starlark_warning.DeprecationNotice `exhaustruct:"optional"`
}

type KurtosisConstructibleBuiltin interface {
	GetName() string

	CreateBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)
}

func (baseBuiltin *KurtosisBaseBuiltin) GetName() string {
	return baseBuiltin.Name
}
