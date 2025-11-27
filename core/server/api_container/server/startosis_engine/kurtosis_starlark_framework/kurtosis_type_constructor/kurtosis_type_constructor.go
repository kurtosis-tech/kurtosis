package kurtosis_type_constructor

import (
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

type Instantiate func(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError)

// KurtosisTypeConstructor is the type for creating builtin used to instantiate a Kurtosis type
type KurtosisTypeConstructor struct {
	// The KurtosisBaseBuiltin listing the name of this builtin and all its argument types
	*kurtosis_starlark_framework.KurtosisBaseBuiltin

	// Instantiate is the function that converts the argument value set into a KurtosisValueType which itself
	// implements starlark.Value
	Instantiate
}

func (builtin *KurtosisTypeConstructor) GetName() string {
	return builtin.Name
}

func (builtin *KurtosisTypeConstructor) CreateBuiltin() func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		wrappedBuiltin, interpretationErr := kurtosis_starlark_framework.WrapKurtosisBaseBuiltin(builtin.KurtosisBaseBuiltin, thread, args, kwargs)
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		instructionWrapper := newKurtosisTypeConstructorInternal(wrappedBuiltin, builtin.Instantiate)
		result, interpretationErr := instructionWrapper.generateTypeInstance()
		if interpretationErr != nil {
			return nil, interpretationErr
		}
		return result, nil
	}
}
