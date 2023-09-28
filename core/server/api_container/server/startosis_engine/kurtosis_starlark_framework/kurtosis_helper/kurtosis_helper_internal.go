package kurtosis_helper

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"go.starlark.net/starlark"
)

type kurtosisHelperInternal struct {
	*kurtosis_starlark_framework.KurtosisBaseBuiltinInternal

	capabilities KurtosisHelperCapabilities
}

func newKurtosisHelperInternal(wrappedBuiltin *kurtosis_starlark_framework.KurtosisBaseBuiltinInternal, capabilities KurtosisHelperCapabilities) *kurtosisHelperInternal {
	return &kurtosisHelperInternal{
		KurtosisBaseBuiltinInternal: wrappedBuiltin,

		capabilities: capabilities,
	}
}

func (builtin *kurtosisHelperInternal) interpret(_ string) (starlark.Value, *startosis_errors.InterpretationError) {

	logrus.Infof("[LEO-DEBUG-8] builtin filename: %s", builtin.GetPosition().GetFilename())
	return builtin.capabilities.Interpret(builtin.GetPosition().GetFilename(), builtin.GetArguments())
}
