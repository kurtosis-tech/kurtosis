package port_spec

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
	"reflect"
	"strings"
)

const (
	PortSpecTypeName = "PortSpec"

	PortNumberAttr              = "number"
	TransportProtocolAttr       = "transport_protocol"
	PortApplicationProtocolAttr = "application_protocol"
	WaitAttr                    = "wait"

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
						return builtin_argument.Uint64InRange(value, PortNumberAttr, minPortNumber, maxPortNumber)
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
				{
					Name:              WaitAttr,
					IsOptional:        true,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.Value],
					// the value can be a string duration, or it can be a Starlark none value (because we are preparing
					// the signature to receive a custom type in the future) when users want to disable it
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.DurationOrNone(value, WaitAttr)
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
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

func CreatePortSpecUsingGoValues(
	portNumber uint16,
	transportProtocol port_spec.TransportProtocol,
	maybeApplicationProtocol *string,
	maybeWaitTimeout string,
) (*PortSpec, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		starlark.MakeInt(int(portNumber)),
		starlark.String(transportProtocol.String()),
	}
	if maybeApplicationProtocol == nil || *maybeApplicationProtocol == "" {
		args = append(args, nil)
	} else {
		args = append(args, starlark.String(*maybeApplicationProtocol))
	}

	if maybeWaitTimeout != "" {
		args = append(args, starlark.String(maybeWaitTimeout))
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

func (portSpec *PortSpec) ToKurtosisType() (*port_spec.PortSpec, *startosis_errors.InterpretationError) {
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

	waitValue, found, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.Value](
		portSpec.KurtosisValueTypeDefault, WaitAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	if !found || waitValue == nil {
		waitValue = starlark.String(port_spec.DefaultWaitTimeoutDurationStr)
	}
	parsedWait, interpretationErr := parsePortWait(waitValue)
	if interpretationErr != nil {
		return nil, interpretationErr
	}

	parsedPortSpec, err := port_spec.NewPortSpec(
		parsedPortNumber,
		parsedTransportProtocol,
		parsedPortApplicationProtocol,
		parsedWait,
	)
	if err != nil {
		// this should never happen since we're checking every attribute defensively here
		return nil, startosis_errors.NewInterpretationError("Unexpected error occurred parsing the following port spec: '%s'", portSpec.String())
	}
	return parsedPortSpec, nil
}

func parsePortNumber(isSet bool, portNumberStarlark starlark.Int) (uint16, *startosis_errors.InterpretationError) {
	if !isSet {
		return 0, startosis_errors.NewInterpretationError("'%s' argument on '%s' is mandatory but was unset. This is a Kurtosis internal bug", PortNumberAttr, PortSpecTypeName)
	}

	portNumber, ok := portNumberStarlark.Uint64()
	if !ok || portNumber > maxPortNumber || portNumber < minPortNumber {
		return 0, startosis_errors.NewInterpretationError("Port number should be in range [%d - %d]", minPortNumber, maxPortNumber)
	}
	return uint16(portNumber), nil
}

func parseTransportProtocol(isSet bool, portProtocol starlark.String) (port_spec.TransportProtocol, *startosis_errors.InterpretationError) {
	if !isSet || portProtocol.GoString() == "" {
		// TODO: to not break backcompat, we allow empty string here and convert it to the default value.
		return port_spec.TransportProtocol_TCP, nil
	}

	transportProtocol, err := port_spec.TransportProtocolString(portProtocol.GoString())
	if err != nil {
		return -1, startosis_errors.NewInterpretationError("Port protocol should be one of %s", strings.Join(port_spec.TransportProtocolStrings(), ", "))
	}
	return transportProtocol, nil
}

func parsePortApplicationProtocol(isSet bool, applicationProtocol starlark.String) (string, *startosis_errors.InterpretationError) {
	if !isSet || applicationProtocol.GoString() == "" {
		// TODO: to not break backcompat, we allow empty string here
		return "", nil
	}
	// validation against the regexp has been run already
	return applicationProtocol.GoString(), nil
}

func parsePortWait(portWaitStarlark starlark.Value) (*port_spec.Wait, *startosis_errors.InterpretationError) {
	if _, ok := portWaitStarlark.(starlark.NoneType); ok {
		return nil, nil
	}
	waitValueStr, ok := portWaitStarlark.(starlark.String)
	if !ok {
		return nil, startosis_errors.NewInterpretationError(
			"Port wait attribute should be string representing a duration. Got '%s'",
			reflect.TypeOf(portWaitStarlark))
	}
	portWait, err := port_spec.CreateWait(waitValueStr.GoString())
	if err != nil {
		return nil, startosis_errors.NewInterpretationError(
			"Port wait attribute should be a duration, i.e. a decimal number followed by a time unit, such as 'ns', 'us' (or 'µs'), 'ms', 's', 'm', 'h'. '%s' is invalid",
			portWait)
	}
	return portWait, nil
}
