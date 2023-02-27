package object_attributes_provider

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strings"
	"time"
)

const (
	apiContainerNameSuffix                 = "kurtosis-api"
	networkingSidecarContainerNameFragment = "networking-sidecar"
	artifactExpansionVolumeNameFragment    = "files-artifact-expansion"
	artifactsExpanderContainerNameFragment = "files-artifacts-expander"
	logsCollectorFragment                  = "kurtosis-logs-collector"
	// The collector is per enclave so this is a suffix
	logsCollectorVolumeFragment = logsCollectorFragment + "-vol"
)

type DockerEnclaveObjectAttributesProvider interface {
	ForEnclaveNetwork(enclaveName string, creationTime time.Time, isPartitioningEnabled bool) (DockerObjectAttributes, error)
	ForEnclaveDataVolume() (DockerObjectAttributes, error)
	ForApiContainer(
		ipAddr net.IP,
		privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
		privateGrpcProxyPortId string,
		privateGrpcProxyPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForUserServiceContainer(
		serviceId service.ServiceName,
		serviceUuid service.ServiceUUID,
		privateIpAddr net.IP,
		privatePorts map[string]*port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForNetworkingSidecarContainer(
		serviceGUIDSidecarAttachedTo service.ServiceUUID,
	) (DockerObjectAttributes, error)
	ForFilesArtifactsExpanderContainer(
		serviceGUID service.ServiceUUID,
	) (DockerObjectAttributes, error)
	ForSingleFilesArtifactExpansionVolume(
		serviceGUID service.ServiceUUID,
	) (DockerObjectAttributes, error)
	ForLogsCollector(tcpPortId string, tcpPortSpec *port_spec.PortSpec, httpPortId string, httpPortSpec *port_spec.PortSpec) (DockerObjectAttributes, error)
	ForLogsCollectorVolume() (DockerObjectAttributes, error)
}

// Private so it can't be instantiated
type dockerEnclaveObjectAttributesProviderImpl struct {
	enclaveId *docker_label_value.DockerLabelValue
}

func newDockerEnclaveObjectAttributesProviderImpl(
	enclaveId *docker_label_value.DockerLabelValue,
) *dockerEnclaveObjectAttributesProviderImpl {
	return &dockerEnclaveObjectAttributesProviderImpl{
		enclaveId: enclaveId,
	}
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForEnclaveNetwork(enclaveName string, creationTime time.Time, isPartitioningEnabled bool) (DockerObjectAttributes, error) {
	enclaveIdStr := provider.enclaveId.GetString()
	name, err := docker_object_name.CreateNewDockerObjectName(enclaveIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", enclaveIdStr)
	}

	creationTimeStr := creationTime.Format(time.RFC3339)

	creationTimeLabelValue, err := docker_label_value.CreateNewDockerLabelValue(creationTimeStr)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a Docker label value object from enclave creation time string '%v'",
			creationTimeStr,
		)
	}

	enclaveNameLabelValue, err := docker_label_value.CreateNewDockerLabelValue(enclaveName)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a Docker label value object from enclave name string '%v'",
			enclaveName,
		)
	}

	// Enclave ID and GUID are the same for an enclave network
	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(
		provider.enclaveId.GetString(),
		provider.enclaveId.GetString(),
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave network using ID '%v'", provider.enclaveId)
	}

	isPartitioningEnabledLabelValue := label_value_consts.NetworkPartitioningDisabledDockerLabelValue
	if isPartitioningEnabled {
		isPartitioningEnabledLabelValue = label_value_consts.NetworkPartitioningEnabledDockerLabelValue
	}
	labels[label_key_consts.IsNetworkPartitioningEnabledDockerLabelKey] = isPartitioningEnabledLabelValue

	labels[label_key_consts.EnclaveCreationTimeLabelKey] = creationTimeLabelValue
	labels[label_key_consts.EnclaveNameDockerLabelKey] = enclaveNameLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the Docker object attributes impl with the name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForEnclaveDataVolume() (DockerObjectAttributes, error) {
	enclaveIdStr := provider.enclaveId.GetString()
	name, err := docker_object_name.CreateNewDockerObjectName(enclaveIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", enclaveIdStr)
	}

	labels := provider.getLabelsForEnclaveObject()
	labels[label_key_consts.VolumeTypeDockerLabelKey] = label_value_consts.EnclaveDataVolumeTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the ObjectAttributesImpl with name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForApiContainer(
	ipAddr net.IP,
	privateGrpcPortId string,
	privateGrpcPortSpec *port_spec.PortSpec,
	privateGrpcProxyPortId string,
	privateGrpcProxyPortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject(
		[]string{
			apiContainerNameSuffix,
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the API container Docker name object")
	}

	privateIpLabelValue, err := docker_label_value.CreateNewDockerLabelValue(ipAddr.String())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a Docker label value object from API container private IP address '%v'",
			ipAddr.String(),
		)
	}

	labels := provider.getLabelsForEnclaveObject()
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.APIContainerContainerTypeDockerLabelValue
	labels[label_key_consts.PrivateIPDockerLabelKey] = privateIpLabelValue

	usedPorts := map[string]*port_spec.PortSpec{
		privateGrpcPortId:      privateGrpcPortSpec,
		privateGrpcProxyPortId: privateGrpcProxyPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following API container ports object to a string for storing in the ports label: %+v", usedPorts)
	}
	labels[label_key_consts.PortSpecsDockerLabelKey] = serializedPortsSpec

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForUserServiceContainer(
	serviceName service.ServiceName,
	serviceUuid service.ServiceUUID,
	privateIpAddr net.IP,
	privatePorts map[string]*port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForUserServiceContainer(
		[]string{
			string(serviceName),
			string(serviceUuid),
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the user service Docker container name object")
	}

	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following user service ports object to a string for storing in the ports label: %+v", privatePorts)
	}

	privateIpLabelValue, err := docker_label_value.CreateNewDockerLabelValue(privateIpAddr.String())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a Docker label value object from user service container private IP address '%v'",
			privateIpAddr.String(),
		)
	}

	serviceIdStr := string(serviceName)
	serviceGuidStr := string(serviceUuid)

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(serviceIdStr, serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave object with GUID '%v'", serviceUuid)
	}
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.UserServiceContainerTypeDockerLabelValue
	labels[label_key_consts.PortSpecsDockerLabelKey] = serializedPortsSpec
	labels[label_key_consts.PrivateIPDockerLabelKey] = privateIpLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForNetworkingSidecarContainer(serviceGUIDSidecarAttachedTo service.ServiceUUID) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject(
		[]string{
			networkingSidecarContainerNameFragment,
			string(serviceGUIDSidecarAttachedTo),
		},
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the networking sidecar Docker container name object")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(string(serviceGUIDSidecarAttachedTo))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave object with GUID '%v'", serviceGUIDSidecarAttachedTo)
	}
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.NetworkingSidecarContainerTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'",
			name.GetString(),
			getLabelKeyValuesAsStrings(labels),
		)
	}

	return objectAttributes, nil
}

// In Docker we get one volume per artifact being expanded
func (provider *dockerEnclaveObjectAttributesProviderImpl) ForSingleFilesArtifactExpansionVolume(
	serviceGUID service.ServiceUUID,
) (
	DockerObjectAttributes,
	error,
) {
	serviceGuidStr := string(serviceGUID)

	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the files artifact expnasion volume for service '%v'", serviceGuidStr)
	}

	name, err := provider.getNameForEnclaveObject([]string{
		artifactExpansionVolumeNameFragment,
		guidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifact expansion volume name object using GUID '%v' and service GUID '%v'", guidStr, serviceGuidStr)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for files artifact expansion volume with GUID '%v'", guidStr)
	}

	serviceGuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from service GUID string '%v'", serviceGuidStr)
	}
	labels[label_key_consts.UserServiceGUIDDockerLabelKey] = serviceGuidLabelValue
	labels[label_key_consts.VolumeTypeDockerLabelKey] = label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue
	// TODO Create a KurtosisResourceDockerLabelKey object, like Kubernetes, and apply the "user-service" label here?

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

// We'll have at most one files artifact expansion container per service, because the single container will handle
// all expansion
func (provider *dockerEnclaveObjectAttributesProviderImpl) ForFilesArtifactsExpanderContainer(
	serviceGUID service.ServiceUUID,
) (
	DockerObjectAttributes,
	error,
) {
	serviceGuidStr := string(serviceGUID)

	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the files artifacts expander container for service '%v'", serviceGuidStr)
	}

	name, err := provider.getNameForEnclaveObject([]string{
		artifactsExpanderContainerNameFragment,
		guidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifacts expander container name with GUID '%v'", guidStr)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for files artifacts expander container with GUID '%v'", guidStr)
	}

	serviceGuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from service GUID string '%v'", serviceGuidStr)
	}
	labels[label_key_consts.UserServiceGUIDDockerLabelKey] = serviceGuidLabelValue
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForLogsCollector(tcpPortId string, tcpPortSpec *port_spec.PortSpec, httpPortId string, httpPortSpec *port_spec.PortSpec) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject([]string{logsCollectorFragment})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the name for the logs collector object")
	}

	labels := provider.getLabelsForEnclaveObject()

	usedPorts := map[string]*port_spec.PortSpec{
		tcpPortId:  tcpPortSpec,
		httpPortId: httpPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following logs-collector-server-ports to a string for storing in the ports label: %+v", usedPorts)
	}

	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.LogsCollectorTypeDockerLabelValue
	labels[label_key_consts.PortSpecsDockerLabelKey] = serializedPortsSpec

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForLogsCollectorVolume() (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject([]string{logsCollectorVolumeFragment})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the name for the logs collector volume object")
	}

	labels := provider.getLabelsForEnclaveObject()

	labels[label_key_consts.VolumeTypeDockerLabelKey] = label_value_consts.LogsCollectorVolumeTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
// Gets the name for an enclave object, making sure to put the enclave ID first and join using the standardized separator
func (provider *dockerEnclaveObjectAttributesProviderImpl) getNameForEnclaveObject(elems []string) (*docker_object_name.DockerObjectName, error) {
	toJoin := []string{
		provider.enclaveId.GetString(),
	}
	toJoin = append(toJoin, elems...)
	nameStr := strings.Join(
		toJoin,
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Docker object name from string '%v'", nameStr)
	}
	return name, nil
}

// Gets the name of the service container
// Gets the name for an enclave object, making sure to put the enclave ID first and join using the standardized separator
func (provider *dockerEnclaveObjectAttributesProviderImpl) getNameForUserServiceContainer(elems []string) (*docker_object_name.DockerObjectName, error) {
	nameStr := strings.Join(
		elems,
		objectNameElementSeparator,
	)
	name, err := docker_object_name.CreateNewDockerObjectName(nameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating Docker object name from string '%v'", nameStr)
	}
	return name, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue {
	return map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.EnclaveUUIDDockerLabelKey: provider.enclaveId,
	}
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithGUID(guid string) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels := provider.getLabelsForEnclaveObject()
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from GUID string '%v'", guid)
	}
	labels[label_key_consts.GUIDDockerLabelKey] = guidLabelValue
	return labels, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithIDAndGUID(id, guid string) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave object labels with GUID '%v'", guid)
	}
	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from ID string '%v'", id)
	}
	labels[label_key_consts.IDDockerLabelKey] = idLabelValue
	return labels, nil
}

func getLabelKeyValuesAsStrings(labels map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range labels {
		result[key.GetString()] = value.GetString()
	}
	return result
}
