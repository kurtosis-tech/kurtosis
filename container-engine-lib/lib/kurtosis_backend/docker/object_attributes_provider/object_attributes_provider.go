package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	engineServerNamePrefix                   = "kurtosis-engine"
	engineServerPortProtocol                 = port_spec.PortProtocol_TCP
)

type DockerObjectAttributesProvider interface {
	ForEngineServer(id string, grpcListenPortNum uint16, grpcProxyListenPortNum uint16) (DockerObjectAttributes, error)
	// ForEnclave(enclaveId string) EnclaveObjectAttributesProvider
}

func GetDockerObjectAttributesProvider() DockerObjectAttributesProvider {
	return newDockerObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type dockerObjectAttributesProviderImpl struct{}
func newDockerObjectAttributesProviderImpl() *dockerObjectAttributesProviderImpl {
	return &dockerObjectAttributesProviderImpl{}
}

func (provider *dockerObjectAttributesProviderImpl) ForEngineServer(id string, grpcListenPortNum uint16, grpcProxyListenPortNum uint16) (DockerObjectAttributes, error) {

	nameStr := strings.Join(
		[]string{
			engineServerNamePrefix,
			id,
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine ID Docker label from string '%v'", id)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", id)
	}

	grpcPortSpec, err := port_spec.NewPortSpec(grpcListenPortNum, engineServerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating grpc port spec object from num '%v' and protocol '%v'",
			grpcListenPortNum,
			engineServerPortProtocol,
		)
	}

	grpcProxyPortSpec, err := port_spec.NewPortSpec(grpcProxyListenPortNum, engineServerPortProtocol)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating grpc-proxy port spec object from num '%v' and protocol '%v'",
			grpcProxyListenPortNum,
			engineServerPortProtocol,
		)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		kurtosisInternalContainerGrpcPortId: grpcPortSpec,
		kurtosisInternalContainerGrpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		ContainerTypeLabelKey: EngineContainerTypeLabelValue,
		PortSpecsLabelKey:     serializedPortsSpec,
		IDLabelKey: idLabelValue,
		GUIDLabelKey: guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

// TODO Fix this!
/*
func (provider *dockerObjectAttributesProviderImpl) ForEnclave(enclaveId string) EnclaveObjectAttributesProvider {
	return newEnclaveObjectAttributesProviderImpl(enclaveId)
}

 */
