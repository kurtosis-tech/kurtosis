package to_http

import (
	"fmt"

	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/types"
	"github.com/dzobbe/PoTE-kurtosis/engine/server/engine/utils"
	"github.com/sirupsen/logrus"

	rpc_api "github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	api_type "github.com/dzobbe/PoTE-kurtosis/api/golang/http_rest/api_types"
)

func warnUnmatchedValue[T any](value T) {
	logrus.Warnf("Unmatched gRPC %T to Http mapping, returning empty value", value)
}

func ToHttpEnclaveStatus(status types.EnclaveStatus) api_type.EnclaveStatus {
	switch status {
	case types.EnclaveStatus_EMPTY:
		return api_type.EnclaveStatusEMPTY
	case types.EnclaveStatus_STOPPED:
		return api_type.EnclaveStatusSTOPPED
	case types.EnclaveStatus_RUNNING:
		return api_type.EnclaveStatusRUNNING
	default:
		warnUnmatchedValue(status)
		panic(fmt.Sprintf("Undefined mapping of value: %s", status))
	}
}

func ToHttpApiContainerStatus(status types.ContainerStatus) api_type.ApiContainerStatus {
	switch status {
	case types.ContainerStatus_NONEXISTENT:
		return api_type.ApiContainerStatusNONEXISTENT
	case types.ContainerStatus_STOPPED:
		return api_type.ApiContainerStatusSTOPPED
	case types.ContainerStatus_RUNNING:
		return api_type.ApiContainerStatusRUNNING
	default:
		warnUnmatchedValue(status)
		panic(fmt.Sprintf("Undefined mapping of value: %s", status))
	}
}

func ToHttpEnclaveAPIContainerInfo(info types.EnclaveAPIContainerInfo) api_type.EnclaveAPIContainerInfo {
	port := int(info.GrpcPortInsideEnclave)
	return api_type.EnclaveAPIContainerInfo{
		ContainerId:           info.ContainerId,
		IpInsideEnclave:       info.IpInsideEnclave,
		GrpcPortInsideEnclave: port,
		BridgeIpAddress:       info.BridgeIpAddress,
	}
}

func ToHttpApiContainerHostMachineInfo(info types.EnclaveAPIContainerHostMachineInfo) api_type.EnclaveAPIContainerHostMachineInfo {
	port := int(info.GrpcPortOnHostMachine)
	return api_type.EnclaveAPIContainerHostMachineInfo{
		IpOnHostMachine:       info.IpOnHostMachine,
		GrpcPortOnHostMachine: port,
	}
}

func ToHttpEnclaveMode(mode types.EnclaveMode) api_type.EnclaveMode {
	switch mode {
	case types.EnclaveMode_PRODUCTION:
		return api_type.PRODUCTION
	case types.EnclaveMode_TEST:
		return api_type.TEST
	default:
		warnUnmatchedValue(mode)
		panic(fmt.Sprintf("Undefined mapping of value: %s", mode))
	}
}

func ToHttpEnclaveInfo(info types.EnclaveInfo) api_type.EnclaveInfo {
	return api_type.EnclaveInfo{
		EnclaveUuid:                 info.EnclaveUuid,
		ShortenedUuid:               info.ShortenedUuid,
		Name:                        info.Name,
		ContainersStatus:            ToHttpEnclaveStatus(info.EnclaveStatus),
		ApiContainerStatus:          ToHttpApiContainerStatus(info.ApiContainerStatus),
		ApiContainerInfo:            utils.MapPointer(info.ApiContainerInfo, ToHttpEnclaveAPIContainerInfo),
		ApiContainerHostMachineInfo: utils.MapPointer(info.ApiContainerHostMachineInfo, ToHttpApiContainerHostMachineInfo),
		CreationTime:                info.CreationTime,
		Mode:                        ToHttpEnclaveMode(info.Mode),
	}
}

func ToHttpEnclaveIdentifiers(identifier *types.EnclaveIdentifiers) api_type.EnclaveIdentifiers {
	return api_type.EnclaveIdentifiers{
		EnclaveUuid:   identifier.EnclaveUuid,
		Name:          identifier.Name,
		ShortenedUuid: identifier.ShortenedUuid,
	}
}

func ToHttpEnclaveNameAndUuid(identifier *types.EnclaveNameAndUuid) api_type.EnclaveNameAndUuid {
	return api_type.EnclaveNameAndUuid{
		Uuid: identifier.Uuid,
		Name: identifier.Name,
	}
}

func ToHttpContainerStatus(status rpc_api.Container_Status) api_type.ContainerStatus {
	switch status {
	case rpc_api.Container_RUNNING:
		return api_type.ContainerStatusRUNNING
	case rpc_api.Container_STOPPED:
		return api_type.ContainerStatusSTOPPED
	case rpc_api.Container_UNKNOWN:
		return api_type.ContainerStatusUNKNOWN
	default:
		warnUnmatchedValue(status)
		panic(fmt.Sprintf("Missing conversion of Container Status Enum value: %s", status))
	}
}

func ToHttpTransportProtocol(protocol rpc_api.Port_TransportProtocol) api_type.TransportProtocol {
	switch protocol {
	case rpc_api.Port_TCP:
		return api_type.TCP
	case rpc_api.Port_UDP:
		return api_type.UDP
	case rpc_api.Port_SCTP:
		return api_type.SCTP
	default:
		warnUnmatchedValue(protocol)
		panic(fmt.Sprintf("Missing conversion of Transport Protocol Enum value: %s", protocol))
	}
}

func ToHttpServiceStatus(status rpc_api.ServiceStatus) api_type.ServiceStatus {
	switch status {
	case rpc_api.ServiceStatus_RUNNING:
		return api_type.ServiceStatusRUNNING
	case rpc_api.ServiceStatus_STOPPED:
		return api_type.ServiceStatusSTOPPED
	case rpc_api.ServiceStatus_UNKNOWN:
		return api_type.ServiceStatusUNKNOWN
	default:
		warnUnmatchedValue(status)
		panic(fmt.Sprintf("Missing conversion of Service Status Enum value: %s", status))
	}
}

func ToHttpContainer(container *rpc_api.Container) api_type.Container {
	status := ToHttpContainerStatus(container.Status)
	return api_type.Container{
		CmdArgs:        container.CmdArgs,
		EntrypointArgs: container.EntrypointArgs,
		EnvVars:        container.EnvVars,
		ImageName:      container.ImageName,
		Status:         status,
	}
}

func ToHttpPorts(port *rpc_api.Port) api_type.Port {
	protocol := ToHttpTransportProtocol(port.TransportProtocol)
	return api_type.Port{
		ApplicationProtocol: &port.MaybeApplicationProtocol,
		WaitTimeout:         &port.MaybeWaitTimeout,
		Number:              int32(port.Number),
		TransportProtocol:   protocol,
	}
}

func ToHttpServiceInfo(service *rpc_api.ServiceInfo) api_type.ServiceInfo {
	container := ToHttpContainer(service.Container)
	serviceStatus := ToHttpServiceStatus(service.ServiceStatus)
	publicPorts := utils.MapMapValues(service.MaybePublicPorts, ToHttpPorts)
	privatePorts := utils.MapMapValues(service.PrivatePorts, ToHttpPorts)
	return api_type.ServiceInfo{
		Container:     container,
		PublicIpAddr:  &service.MaybePublicIpAddr,
		PublicPorts:   &publicPorts,
		Name:          service.Name,
		PrivateIpAddr: service.PrivateIpAddr,
		PrivatePorts:  privatePorts,
		ServiceStatus: serviceStatus,
		ServiceUuid:   service.ServiceUuid,
		ShortenedUuid: service.ShortenedUuid,
	}
}

func ToHttpFeatureFlag(flag rpc_api.KurtosisFeatureFlag) api_type.KurtosisFeatureFlag {
	switch flag {
	case rpc_api.KurtosisFeatureFlag_NO_INSTRUCTIONS_CACHING:
		return api_type.NOINSTRUCTIONSCACHING
	default:
		warnUnmatchedValue(flag)
		panic(fmt.Sprintf("Missing conversion of Feature Flag Enum value: %s", flag))
	}
}

func ToHttpRestartPolicy(policy rpc_api.RestartPolicy) api_type.RestartPolicy {
	switch policy {
	case rpc_api.RestartPolicy_ALWAYS:
		return api_type.RestartPolicyALWAYS
	case rpc_api.RestartPolicy_NEVER:
		return api_type.RestartPolicyNEVER
	default:
		warnUnmatchedValue(policy)
		panic(fmt.Sprintf("Missing conversion of Restart Policy Enum value: %s", policy))
	}
}
