package kurtosis_type_constructor

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

type KurtosisTypeConstructorInternal struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltinInternal

	Instantiate
}

func newKurtosisTypeConstructorInternal(internalBuiltin *kurtosis_starlark_framework.KurtosisBaseBuiltinInternal, instantiate Instantiate) *KurtosisTypeConstructorInternal {
	return &KurtosisTypeConstructorInternal{
		KurtosisBaseBuiltinInternal: internalBuiltin,
		Instantiate:                 instantiate,
	}
}

func (builtin *KurtosisTypeConstructorInternal) generateTypeInstance() (starlark.Value, *startosis_errors.InterpretationError) {
	kurtosisType, interpretationErr := builtin.Instantiate(builtin.GetArguments())
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return kurtosisType, nil
}
