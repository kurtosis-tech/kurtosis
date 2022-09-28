package kurtosis_instruction

import (
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func ParseServiceId(serviceIdRaw starlark.String) (service.ServiceID, error) {
	// TODO(gb): maybe prohibit certain characters for service ids
	serviceId, err := safeCastToString(serviceIdRaw, "service_id")
	if err != nil {
		return "", err
	}
	if len(serviceId) == 0 {
		return "", errors.New("serviceId cannoe be empty")
	}
	return service.ServiceID(serviceId), nil
}

func ParseServiceConfigArg(serviceConfig *starlarkstruct.Struct) (*kurtosis_core_rpc_api_bindings.ServiceConfig, error) {
	containerImageName, err := parseServiceConfigContainerImageName(serviceConfig)
	if err != nil {
		return nil, err
	}

	privatePorts, err := parseServiceConfigPrivatePorts(serviceConfig)
	if err != nil {
		return nil, err
	}

	return &kurtosis_core_rpc_api_bindings.ServiceConfig{
		ContainerImageName: containerImageName,
		PrivatePorts:       privatePorts,
	}, nil
}

func parseServiceConfigContainerImageName(serviceConfig *starlarkstruct.Struct) (string, error) {
	// containerImageName should be a simple string
	containerImageName, err := extractStringValue(serviceConfig, "container_image_name")
	if err != nil {
		return "", err
	}
	return containerImageName, nil
}

func parseServiceConfigPrivatePorts(serviceConfig *starlarkstruct.Struct) (map[string]*kurtosis_core_rpc_api_bindings.Port, error) {
	privatePortsArg, err := serviceConfig.Attr("used_ports")
	if err != nil {
		return nil, errors.New("Missing `used_ports` as part of the service config")
	}
	switch privatePortsArg.(type) {
	default:
		return nil, errors.New("`used_ports` is expected to be a dictionary")
	case *starlark.Dict:
		break
	}

	var privatePorts = make(map[string]*kurtosis_core_rpc_api_bindings.Port)
	for _, portNameRaw := range privatePortsArg.(*starlark.Dict).Keys() {
		portDefinition, found, err := privatePortsArg.(*starlark.Dict).Get(portNameRaw)
		if !found || err != nil {
			return nil, errors.New("unable to find a value in a dict associated with a key that exists - this is a product bug")
		}

		portName, err := safeCastToString(portNameRaw, "port_used key")
		if err != nil {
			return nil, err
		}

		switch portDefinition.(type) {
		default:
			return nil, errors.New("`used_ports` is expected to be a dictionary")
		case *starlarkstruct.Struct:
			break
		}

		port, err := parsePort(portDefinition.(*starlarkstruct.Struct))
		if err != nil {
			return nil, err
		}
		privatePorts[portName] = port
	}
	return privatePorts, nil
}

func parsePort(portArg *starlarkstruct.Struct) (*kurtosis_core_rpc_api_bindings.Port, error) {
	portNumber, err := extractUint32Value(portArg, "number")
	if err != nil {
		return nil, err
	}
	if portNumber > 65535 {
		return nil, errors.New("port number should be strictly lower than 65536")
	}

	protocolRaw, err := extractStringValue(portArg, "protocol")
	if err != nil {
		return nil, err
	}
	protocol, err := parsePortProtocol(protocolRaw)
	if err != nil {
		return nil, err
	}

	return &kurtosis_core_rpc_api_bindings.Port{
		Number:   portNumber,
		Protocol: protocol,
	}, nil
}

func parsePortProtocol(portProtocol string) (kurtosis_core_rpc_api_bindings.Port_Protocol, error) {
	switch portProtocol {
	case "TCP":
		return kurtosis_core_rpc_api_bindings.Port_TCP, nil
	case "SCTP":
		return kurtosis_core_rpc_api_bindings.Port_SCTP, nil
	case "UDP":
		return kurtosis_core_rpc_api_bindings.Port_UDP, nil
	default:
		return -1, errors.New("port protocol should be either `TCP`, `UDP` or `SCTP`")
	}
}

func extractStringValue(structField *starlarkstruct.Struct, key string) (string, error) {
	value, err := structField.Attr(key)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Missing `%s` as part of the struct object", key))
	}
	stringValue, err := safeCastToString(value, key)
	if err != nil {
		return "", err
	}
	return stringValue, nil
}

func extractUint32Value(structField *starlarkstruct.Struct, key string) (uint32, error) {
	value, err := structField.Attr(key)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Missing `%s` as part of the struct object", key))
	}
	uint32Value, err := safeCastToUint32(value, key)
	if err != nil {
		return 0, err
	}
	return uint32Value, nil
}

func safeCastToString(expectedValueString starlark.Value, argNameForLogging string) (string, error) {
	switch expectedValueString.(type) {
	default:
		return "", errors.New(fmt.Sprintf("`%s` arg is expected to be a string", argNameForLogging))
	case starlark.String:
		return expectedValueString.(starlark.String).GoString(), nil
	}
}

func safeCastToUint32(expectedValueString starlark.Value, argNameForLogging string) (uint32, error) {
	switch expectedValueString.(type) {
	default:
		return 0, errors.New(fmt.Sprintf("`%s` arg is expected to be a uint32", argNameForLogging))
	case starlark.Int:
		uint64Value, ok := expectedValueString.(starlark.Int).Uint64()
		if !ok {
			return 0, errors.New(fmt.Sprintf("`%s` arg is expected to be a uint32", argNameForLogging))
		}
		if uint64Value != uint64(uint32(uint64Value)) {
			// safeguard against "overflow"
			return 0, errors.New(fmt.Sprintf("`%s` arg is expected to be a uint32", argNameForLogging))
		}
		return uint32(uint64Value), nil
	}
}
