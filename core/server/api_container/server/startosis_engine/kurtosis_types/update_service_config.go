package kurtosis_types

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"strings"
)

const (
	UpdateServiceConfigTypeName = "UpdateServiceConfig"
	subnetworkAttr              = "subnetwork"
)

type UpdateServiceConfig struct {
	subnetwork starlark.String
}

func NewUpdateServiceConfig(subnetwork starlark.String) *UpdateServiceConfig {
	return &UpdateServiceConfig{
		subnetwork: subnetwork,
	}
}

func MakeUpdateServiceConfig(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var subnetwork starlark.String
	if err := starlark.UnpackArgs(b.Name(), args, kwargs, subnetworkAttr, &subnetwork); err != nil {
		return nil, startosis_errors.WrapWithValidationError(err, "Cannot construct '%s' from the provided arguments. Expecting a single argument '%s'", UpdateServiceConfigTypeName, subnetworkAttr)
	}
	return NewUpdateServiceConfig(subnetwork), nil
}

// String the starlark.Value interface
func (config *UpdateServiceConfig) String() string {
	buffer := new(strings.Builder)
	buffer.WriteString(UpdateServiceConfigTypeName + "(")
	buffer.WriteString(subnetworkAttr + "=")
	buffer.WriteString(fmt.Sprintf("%v)", config.subnetwork))
	return buffer.String()
}

// Type implements the starlark.Value interface
func (config *UpdateServiceConfig) Type() string {
	return UpdateServiceConfigTypeName
}

// Freeze implements the starlark.Value interface
func (config *UpdateServiceConfig) Freeze() {
	// this is a no-op its already immutable
}

// Truth implements the starlark.Value interface
func (config *UpdateServiceConfig) Truth() starlark.Bool {
	// As long as it is instantiated, UpdateServiceConfig is necessarily true
	// empty string is also a valid value since it points to DefaultPartitionID
	return true
}

// Hash implements the starlark.Value interface
// This shouldn't be hashed
func (config *UpdateServiceConfig) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: '%s'", ConnectionConfigTypeName)
}

// Attr implements the starlark.HasAttrs interface.
func (config *UpdateServiceConfig) Attr(name string) (starlark.Value, error) {
	switch name {
	case subnetworkAttr:
		return config.subnetwork, nil
	default:
		return nil, fmt.Errorf("'%s' has no attribute '%s'", UpdateServiceConfigTypeName, name)
	}
}

// AttrNames implements the starlark.HasAttrs interface.
func (config *UpdateServiceConfig) AttrNames() []string {
	return []string{subnetworkAttr}
}

func (config *UpdateServiceConfig) ToKurtosisType() *kurtosis_core_rpc_api_bindings.UpdateServiceConfig {
	return binding_constructors.NewUpdateServiceConfig(config.subnetwork.GoString())
}
