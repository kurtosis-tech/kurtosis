package api_container

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/container_status"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"net"
)

// Represents point-in-time information about an API container
// WARNING: Store this at your own risk!
type APIContainer struct {
	// The ID of the enclave the API container manages
	enclaveId enclave.EnclaveID

	status container_status.ContainerStatus

	// Private (i.e. internal to enclave) information about the API container
	privateIpAddr        net.IP
	privateGrpcPort      *port_spec.PortSpec
	privateGrpcProxyPort *port_spec.PortSpec

	// Public (i.e. external to Kurtosis) information about the API container
	// This information will be nil if the API container isn't running
	publicIpAddr        net.IP
	publicGrpcPort      *port_spec.PortSpec
	publicGrpcProxyPort *port_spec.PortSpec
}

func NewAPIContainer(
	enclaveId enclave.EnclaveID,
	status container_status.ContainerStatus,
	privateIpAddr net.IP,
	privateGrpcPort *port_spec.PortSpec,
	privateGrpcProxyPort *port_spec.PortSpec,
	publicIpAddr net.IP,
	publicGrpcPort *port_spec.PortSpec,
	publicGrpcProxyPort *port_spec.PortSpec,
) *APIContainer {
	return &APIContainer{
		enclaveId:            enclaveId,
		status:               status,
		privateIpAddr:        privateIpAddr,
		privateGrpcPort:      privateGrpcPort,
		privateGrpcProxyPort: privateGrpcProxyPort,
		publicIpAddr:         publicIpAddr,
		publicGrpcPort:       publicGrpcPort,
		publicGrpcProxyPort:  publicGrpcProxyPort}
}

func (apiContainer *APIContainer) GetEnclaveID() enclave.EnclaveID {
	return apiContainer.enclaveId
}

func (apiContainer *APIContainer) GetStatus() container_status.ContainerStatus {
	return apiContainer.status
}

func (apiContainer *APIContainer) GetPrivateIPAddress() net.IP {
	return apiContainer.privateIpAddr
}

func (apiContainer *APIContainer) GetPrivateGRPCPort() *port_spec.PortSpec {
	return apiContainer.privateGrpcPort
}

func (apiContainer *APIContainer) GetPrivateGRPCProxyPort() *port_spec.PortSpec {
	return apiContainer.privateGrpcPort
}

func (apiContainer *APIContainer) GetPublicIPAddress() net.IP {
	return apiContainer.publicIpAddr
}

func (apiContainer *APIContainer) GetPublicGRPCPort() *port_spec.PortSpec {
	return apiContainer.publicGrpcPort
}

func (apiContainer *APIContainer) GetPublicGRPCProxyPort() *port_spec.PortSpec {
	return apiContainer.publicGrpcProxyPort
}
