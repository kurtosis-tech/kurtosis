package object_attributes_provider

import (
	"strings"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	engineServerNamePrefix = "kurtosis-engine"
	logsAggregatorName     = "kurtosis-logs-aggregator"
	logsStorageVolumeName  = "kurtosis-logs-storage"
	reverseProxyNamePrefix = "kurtosis-reverse-proxy"
)

type DockerObjectAttributesProvider interface {
	ForEngineServer(
		guid engine.EngineGUID,
		grpcPortId string,
		grpcPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForEnclave(enclaveUuid enclave.EnclaveUUID) (DockerEnclaveObjectAttributesProvider, error)
	ForLogsAggregator() (DockerObjectAttributes, error)
	ForLogsStorageVolume() (DockerObjectAttributes, error)
	ForReverseProxy(engineGuid engine.EngineGUID) (DockerObjectAttributes, error)
}

func GetDockerObjectAttributesProvider() DockerObjectAttributesProvider {
	return newDockerObjectAttributesProviderImpl()
}

// Private so it can't be instantiated
type dockerObjectAttributesProviderImpl struct{}

func newDockerObjectAttributesProviderImpl() *dockerObjectAttributesProviderImpl {
	return &dockerObjectAttributesProviderImpl{}
}

func (provider *dockerObjectAttributesProviderImpl) ForEngineServer(
	guid engine.EngineGUID,
	grpcPortId string,
	grpcPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {

	nameStr := strings.Join(
		[]string{
			engineServerNamePrefix,
			string(guid),
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(guid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the engine GUID Docker label from string '%v'", guid)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		grpcPortId: grpcPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		docker_label_key.ContainerTypeDockerLabelKey: label_value_consts.EngineContainerTypeDockerLabelValue,
		docker_label_key.PortSpecsDockerLabelKey:     serializedPortsSpec,
		docker_label_key.IDDockerLabelKey:            idLabelValue,
		docker_label_key.GUIDDockerLabelKey:          guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForEnclave(enclaveUuid enclave.EnclaveUUID) (DockerEnclaveObjectAttributesProvider, error) {
	enclaveUuidStr := string(enclaveUuid)
	enclaveUuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(enclaveUuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value out of enclave ID string '%v'", enclaveUuidStr)
	}

	return newDockerEnclaveObjectAttributesProviderImpl(enclaveUuidLabelValue), nil
}

func (provider *dockerObjectAttributesProviderImpl) ForLogsAggregator() (DockerObjectAttributes, error) {
	name, err := docker_object_name.CreateNewDockerObjectName(logsAggregatorName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", logsAggregatorName)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		docker_label_key.ContainerTypeDockerLabelKey: label_value_consts.LogsAggregatorTypeDockerLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}
	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForLogsStorageVolume() (DockerObjectAttributes, error) {
	name, err := docker_object_name.CreateNewDockerObjectName(logsStorageVolumeName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", logsStorageVolumeName)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		docker_label_key.VolumeTypeDockerLabelKey: label_value_consts.LogsStorageVolumeTypeDockerLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}
	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForReverseProxy(engineGuid engine.EngineGUID) (DockerObjectAttributes, error) {

	nameStr := strings.Join(
		[]string{
			reverseProxyNamePrefix,
			string(engineGuid),
		},
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(engineGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the reverse proxy GUID Docker label from string '%v'", engineGuid)
	}
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(string(engineGuid))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the reverse proxy GUID Docker label from string '%v'", engineGuid)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		docker_label_key.ContainerTypeDockerLabelKey: label_value_consts.ReverseProxyTypeDockerLabelValue,
		docker_label_key.IDDockerLabelKey:            idLabelValue,
		docker_label_key.GUIDDockerLabelKey:          guidLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}
	return objectAttributes, nil
}
