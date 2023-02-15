package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/engine"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	engineServerNamePrefix = "kurtosis-engine"
	logsDatabaseName       = "kurtosis-logs-db"

	//We always use the same name because we are going to have only one instance of this volume,
	//so when the engine is restarted it mounts the same volume with the previous logs
	logsDatabaseVolumeName = logsDatabaseName + "-vol"
)

type DockerObjectAttributesProvider interface {
	ForEngineServer(
		guid engine.EngineGUID,
		grpcPortId string,
		grpcPortSpec *port_spec.PortSpec,
		grpcProxyPortId string,
		grpcProxyPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForEnclave(enclaveUuid enclave.EnclaveUUID) (DockerEnclaveObjectAttributesProvider, error)
	ForLogsDatabase(
		httpApiPortId string,
		httpApiPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForLogsDatabaseVolume() (DockerObjectAttributes, error)
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
	grpcProxyPortId string,
	grpcProxyPortSpec *port_spec.PortSpec,
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
		grpcPortId:      grpcPortSpec,
		grpcProxyPortId: grpcProxyPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following engine server ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.ContainerTypeDockerLabelKey: label_value_consts.EngineContainerTypeDockerLabelValue,
		label_key_consts.PortSpecsDockerLabelKey:     serializedPortsSpec,
		label_key_consts.IDDockerLabelKey:            idLabelValue,
		label_key_consts.GUIDDockerLabelKey:          guidLabelValue,
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

func (provider *dockerObjectAttributesProviderImpl) ForLogsDatabase(
	httpApiPortId string,
	httpApiPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {

	name, err := docker_object_name.CreateNewDockerObjectName(logsDatabaseName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", logsDatabaseName)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		httpApiPortId: httpApiPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following logs-database-server-ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.ContainerTypeDockerLabelKey: label_value_consts.LogsDatabaseTypeDockerLabelValue,
		label_key_consts.PortSpecsDockerLabelKey:     serializedPortsSpec,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerObjectAttributesProviderImpl) ForLogsDatabaseVolume() (DockerObjectAttributes, error) {
	nameStr := logsDatabaseVolumeName
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", nameStr)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.VolumeTypeDockerLabelKey: label_value_consts.LogsDatabaseVolumeTypeDockerLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}
