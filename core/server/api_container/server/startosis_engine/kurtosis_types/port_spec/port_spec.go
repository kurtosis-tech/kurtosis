package port_spec

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_warning"
	"go.starlark.net/starlark"
	"strings"
)

const (
	PortSpecTypeName = "PortSpec"

	PortNumberAttr              = "number"
	TransportProtocolAttr       = "transport_protocol"
	PortApplicationProtocolAttr = "application_protocol"

	maxPortNumber                 = 65535
	minPortNumber                 = 1
	validApplicationProtocolRegex = "^[a-zA-Z0-9+.-]*$"
)

func NewPortSpecType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: PortSpecTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              PortNumberAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Int],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.Uint64InRange(value, TransportProtocolAttr, minPortNumber, maxPortNumber)
					},
				},
				{
					Name:              TransportProtocolAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringValues(value, TransportProtocolAttr, port_spec.TransportProtocolStrings())
					},
				},
				{
					Name:              PortApplicationProtocolAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.StringRegexp(value, PortApplicationProtocolAttr, validApplicationProtocolRegex)
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	startosis_warning.Printf("This is Warning From Stream; lets' go !!!")

	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(PortSpecTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &PortSpec{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type PortSpec struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func CreatePortSpec(portNumber uint32, transportProtocol kurtosis_core_rpc_api_bindings.Port_TransportProtocol, maybeApplicationProtocol string) (*PortSpec, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		starlark.MakeInt(int(portNumber)),
		starlark.String(transportProtocol.String()),
	}
	if maybeApplicationProtocol != "" {
		args = append(args, starlark.String(maybeApplicationProtocol))
	} else {
		args = append(args, nil)
	}

	argumentDefinitions := NewPortSpecType().KurtosisBaseBuiltin.Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(PortSpecTypeName, argumentValuesSet)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &PortSpec{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func (portSpec *PortSpec) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := portSpec.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &PortSpec{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (portSpec *PortSpec) ToKurtosisType() (*kurtosis_core_rpc_api_bindings.Port, *startosis_errors.InterpretationError) {
	portNumber, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Int](
		portSpec.KurtosisValueTypeDefault, PortNumberAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	parsedPortNumber, interpretationErr := parsePortNumber(found, portNumber)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	transportProtocol, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		portSpec.KurtosisValueTypeDefault, TransportProtocolAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	parsedTransportProtocol, interpretationErr := parseTransportProtocol(found, transportProtocol)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	portApplicationProtocol, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		portSpec.KurtosisValueTypeDefault, PortApplicationProtocolAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	parsedPortApplicationProtocol, interpretationErr := parsePortApplicationProtocol(found, portApplicationProtocol)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return binding_constructors.NewPort(parsedPortNumber, parsedTransportProtocol, parsedPortApplicationProtocol), nil
}

func parsePortNumber(isSet bool, portNumberStarlark starlark.Int) (uint32, *startosis_errors.InterpretationError) {
	if !isSet {
		return 0, startosis_errors.NewInterpretationError("'%s' argument on '%s' is mandatory but was unset. This is a Kurtosis internal bug", PortNumberAttr, PortSpecTypeName)
	}

	portNumber, ok := portNumberStarlark.Uint64()
	if !ok || portNumber > maxPortNumber || portNumber < minPortNumber {
		return 0, startosis_errors.NewInterpretationError("Port number should be in range [%d - %d]", minPortNumber, maxPortNumber)
	}
	return uint32(portNumber), nil
}

func parseTransportProtocol(isSet bool, portProtocol starlark.String) (kurtosis_core_rpc_api_bindings.Port_TransportProtocol, *startosis_errors.InterpretationError) {
	if !isSet || portProtocol == "" {
		// TODO: to not break backcompat, we allow empty string here and convert it to the default value.
		return kurtosis_core_rpc_api_bindings.Port_TCP, nil
	}

	parsedTransportProtocol, found := kurtosis_core_rpc_api_bindings.Port_TransportProtocol_value[portProtocol.GoString()]
	if !found {
		return -1, startosis_errors.NewInterpretationError("Port protocol should be one of %s", strings.Join(port_spec.TransportProtocolStrings(), ", "))
	}
	return kurtosis_core_rpc_api_bindings.Port_TransportProtocol(parsedTransportProtocol), nil
}

func parsePortApplicationProtocol(isSet bool, applicationProtocol starlark.String) (string, *startosis_errors.InterpretationError) {
	if !isSet || applicationProtocol == "" {
		// TODO: to not break backcompat, we allow empty string here
		return "", nil
	}
	// validation against the regexp has been run already
	return applicationProtocol.GoString(), nil
}
