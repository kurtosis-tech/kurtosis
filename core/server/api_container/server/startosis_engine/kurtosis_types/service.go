package kurtosis_types

import (
	"github.com/dzobbe/PoTE-kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/builtin_argument"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/kurtosis_starlark_framework/kurtosis_type_constructor"
	"github.com/dzobbe/PoTE-kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"go.starlark.net/starlark"
)

const (
	ServiceTypeName = "Service"

	HostnameAttr    = "hostname"
	IpAddressAttr   = "ip_address"
	PortsAttr       = "ports"
	ServiceNameAttr = "name"
)

func NewServiceType() *kurtosis_type_constructor.KurtosisTypeConstructor {
	return &kurtosis_type_constructor.KurtosisTypeConstructor{
		KurtosisBaseBuiltin: &kurtosis_starlark_framework.KurtosisBaseBuiltin{
			Name: ServiceTypeName,

			Arguments: []*builtin_argument.BuiltinArgument{
				{
					Name:              ServiceNameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, ServiceNameAttr)
					},
				},
				{
					Name:              HostnameAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, HostnameAttr)
					},
				},
				{
					Name:              IpAddressAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[starlark.String],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return builtin_argument.NonEmptyString(value, IpAddressAttr)
					},
				},
				{
					Name:              PortsAttr,
					IsOptional:        false,
					ZeroValueProvider: builtin_argument.ZeroValueProvider[*starlark.Dict],
					Validator: func(value starlark.Value) *startosis_errors.InterpretationError {
						return nil
					},
				},
			},
		},

		Instantiate: instantiate,
	}
}

func instantiate(arguments *builtin_argument.ArgumentValuesSet) (builtin_argument.KurtosisValueType, *startosis_errors.InterpretationError) {
	kurtosisValueType, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ServiceTypeName, arguments)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Service{
		KurtosisValueTypeDefault: kurtosisValueType,
	}, nil
}

type Service struct {
	*kurtosis_type_constructor.KurtosisValueTypeDefault
}

func CreateService(serviceName starlark.String, hostname starlark.String, ipAddress starlark.String, ports *starlark.Dict) (*Service, *startosis_errors.InterpretationError) {
	args := []starlark.Value{
		serviceName,
		hostname,
		ipAddress,
		ports,
	}

	argumentDefinitions := NewServiceType().Arguments
	argumentValuesSet := builtin_argument.NewArgumentValuesSet(argumentDefinitions, args)
	kurtosisDefaultValue, interpretationErr := kurtosis_type_constructor.CreateKurtosisStarlarkTypeDefault(ServiceTypeName, argumentValuesSet)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return &Service{
		KurtosisValueTypeDefault: kurtosisDefaultValue,
	}, nil
}

func (serviceObj *Service) Copy() (builtin_argument.KurtosisValueType, error) {
	copiedValueType, err := serviceObj.KurtosisValueTypeDefault.Copy()
	if err != nil {
		return nil, err
	}
	return &Service{
		KurtosisValueTypeDefault: copiedValueType,
	}, nil
}

func (serviceObj *Service) GetName() (service.ServiceName, *startosis_errors.InterpretationError) {
	serviceName, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		serviceObj.KurtosisValueTypeDefault, ServiceNameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return service.ServiceName(serviceName.GoString()), nil
}

func (serviceObj *Service) GetHostname() (string, *startosis_errors.InterpretationError) {
	hostname, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		serviceObj.KurtosisValueTypeDefault, HostnameAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return hostname.GoString(), nil
}

func (serviceObj *Service) GetIpAddress() (string, *startosis_errors.InterpretationError) {
	ipAddress, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[starlark.String](
		serviceObj.KurtosisValueTypeDefault, IpAddressAttr)
	if interpretationErr != nil {
		return "", interpretationErr
	}
	return ipAddress.GoString(), nil
}

func (serviceObj *Service) GetPorts() (*starlark.Dict, *startosis_errors.InterpretationError) {
	ports, _, interpretationErr := kurtosis_type_constructor.ExtractAttrValue[*starlark.Dict](
		serviceObj.KurtosisValueTypeDefault, PortsAttr)
	if interpretationErr != nil {
		return nil, interpretationErr
	}
	return ports, nil
}
