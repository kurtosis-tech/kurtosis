package kurtosis_helper

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"go.starlark.net/starlark"
)

type KurtosisHelper struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltin

	Capabilities KurtosisHelperCapabilities
}

func (builtin *KurtosisHelper) GetName() string {
	return builtin.Name
}

func (builtin *KurtosisHelper) CreateBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		wrappedBuiltin, interpretationErr := kurtosis_starlark_framework.WrapKurtosisBaseBuiltin(builtin.KurtosisBaseBuiltin, thread, args, kwargs)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		helperInternal := newKurtosisHelperInternal(wrappedBuiltin, builtin.Capabilities)
		result, interpretationErr := helperInternal.interpret()
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		return result, nil
	}
}
