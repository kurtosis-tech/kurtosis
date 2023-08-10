package builtins

import (
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	KurtosisModuleName = "kurtosis"
)

// TODO This module was created for storing Kurtosis constatns that then can be used for any Kurtosis module or package
// TODO we use to store the contants related with subnetworks (BLOCKED, ALLOWED) here but these were removed since we deprecate the network partitioning feature
// TODO it's planned to store some another constants like the port protols here, so we are leaving it here for this reason.
func KurtosisModule() (*starlarkstruct.Module, *startosis_errors.InterpretationError) {

	return &starlarkstruct.Module{
		Name:    KurtosisModuleName,
		Members: starlark.StringDict{
			//Add sub-modules here
		},
	}, nil
}
