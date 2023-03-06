package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_types/connection_config"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	KurtosisModuleName = "kurtosis"

	connectionSubmoduleName            = "connection"
	connectionSubmoduleBlockedAttrName = "BLOCKED"
	connectionSubmoduleAllowedAttrName = "ALLOWED"
)

func KurtosisModule() (*starlarkstruct.Module, *startosis_errors.InterpretationError) {
	connectionModule, interpretationErr := newConnectionModule()
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &starlarkstruct.Module{
		Name: KurtosisModuleName,
		Members: starlark.StringDict{
			connectionSubmoduleName: connectionModule,
		},
	}, nil
}

func newConnectionModule() (*starlarkstruct.Module, *startosis_errors.InterpretationError) {
	blockedConnectionConfig, interpretationErr := connection_config.CreateConnectionConfig(starlark.Float(100))
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Unable to initiate the connection module. This is a Kurtosis internal bug")
	}
	allowedConnectionConfig, interpretationErr := connection_config.CreateConnectionConfig(starlark.Float(0))
	if interpretationErr != nil {
		return nil, startosis_errors.WrapWithInterpretationError(interpretationErr, "Unable to initiate the connection module. This is a Kurtosis internal bug")
	}
	return &starlarkstruct.Module{
		Name: connectionSubmoduleName,
		Members: starlark.StringDict{
			connectionSubmoduleBlockedAttrName: blockedConnectionConfig,
			connectionSubmoduleAllowedAttrName: allowedConnectionConfig,
		},
	}, nil
}
