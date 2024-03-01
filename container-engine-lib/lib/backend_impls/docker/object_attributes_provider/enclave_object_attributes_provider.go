package object_attributes_provider

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service_directory"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	networkPrefix          = "kt-"
	apiContainerNamePrefix = "kurtosis-api"

	artifactExpansionVolumeNameFragment = "files-artifact-expansion"

	artifactsExpanderContainerNameFragment = "files-artifacts-expander"
	logsCollectorFragment                  = "kurtosis-logs-collector"
	// The collector is per enclave so this is a suffix
	logsCollectorVolumeFragment = logsCollectorFragment + "-vol"
)

type DockerEnclaveObjectAttributesProvider interface {
	ForEnclaveNetwork(enclaveName string, creationTime time.Time) (DockerObjectAttributes, error)
	ForEnclaveDataVolume() (DockerObjectAttributes, error)
	ForApiContainer(
		ipAddr net.IP,
		privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForUserServiceContainer(
		serviceName service.ServiceName,
		serviceUuid service.ServiceUUID,
		privateIpAddr net.IP,
		privatePorts map[string]*port_spec.PortSpec,
		userLabels map[string]string,
	) (DockerObjectAttributes, error)
	ForFilesArtifactsExpanderContainer(
		serviceUUID service.ServiceUUID,
	) (DockerObjectAttributes, error)
	ForSingleFilesArtifactExpansionVolume(
		serviceUUID service.ServiceUUID,
	) (DockerObjectAttributes, error)
	ForSinglePersistentDirectoryVolume(
		persistentKey service_directory.DirectoryPersistentKey,
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

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForEnclaveNetwork(enclaveName string, creationTime time.Time) (DockerObjectAttributes, error) {
	// TODO: might need to revert this if we have multiple users on the same cluster (what if two people create enclaves with name test?)
	enclaveNetworkNameStr := networkPrefix + enclaveName
	name, err := docker_object_name.CreateNewDockerObjectName(enclaveNetworkNameStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", enclaveNetworkNameStr)
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

	labels[docker_label_key.EnclaveCreationTimeLabelKey] = creationTimeLabelValue
	labels[docker_label_key.EnclaveNameDockerLabelKey] = enclaveNameLabelValue

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
	labels[docker_label_key.VolumeTypeDockerLabelKey] = label_value_consts.EnclaveDataVolumeTypeDockerLabelValue

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
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject(
		[]string{
			apiContainerNamePrefix,
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
	labels[docker_label_key.ContainerTypeDockerLabelKey] = label_value_consts.APIContainerContainerTypeDockerLabelValue
	labels[docker_label_key.PrivateIPDockerLabelKey] = privateIpLabelValue

	usedPorts := map[string]*port_spec.PortSpec{
		privateGrpcPortId: privateGrpcPortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following API container ports object to a string for storing in the ports label: %+v", usedPorts)
	}
	labels[docker_label_key.PortSpecsDockerLabelKey] = serializedPortsSpec

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
	userLabels map[string]string,
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForUserServiceContainer(
		serviceName,
		serviceUuid,
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

	serviceNameStr := string(serviceName)
	serviceUuidStr := string(serviceUuid)
	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(serviceNameStr, serviceUuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave object with UUID '%v'", serviceUuid)
	}
	labels[docker_label_key.ContainerTypeDockerLabelKey] = label_value_consts.UserServiceContainerTypeDockerLabelValue
	labels[docker_label_key.PortSpecsDockerLabelKey] = serializedPortsSpec
	labels[docker_label_key.PrivateIPDockerLabelKey] = privateIpLabelValue

	// add user custom label
	for userLabelKey, userLabelValue := range userLabels {
		dockerLabelKey, err := docker_label_key.CreateNewDockerUserCustomLabelKey(userLabelKey)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new user custom Docker label key '%s'", userLabelKey)
		}
		dockerLabelValue, err := docker_label_value.CreateNewDockerLabelValue(userLabelValue)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating a new user custom Docker label value '%s'", userLabelValue)
		}
		labels[dockerLabelKey] = dockerLabelValue
	}

	traefikLabels, err := provider.getTraefikLabelsForEnclaveObject(serviceUuidStr, privatePorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting traefik labels for enclave object with UUID '%v'", serviceUuid)
	}
	for traefikLabelKey, traefikLabelValue := range traefikLabels {
		labels[traefikLabelKey] = traefikLabelValue
	}

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
	serviceUUID service.ServiceUUID,
) (
	DockerObjectAttributes,
	error,
) {
	serviceUuidStr := string(serviceUUID)

	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the files artifact expnasion volume for service '%v'", serviceUuidStr)
	}

	name, err := provider.getNameForEnclaveObject([]string{
		artifactExpansionVolumeNameFragment,
		guidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifact expansion volume name object using GUID '%v' and service GUID '%v'", guidStr, serviceUuidStr)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for files artifact expansion volume with UUID '%v'", guidStr)
	}

	serviceUuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceUuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from service GUID string '%v'", serviceUuidStr)
	}
	labels[docker_label_key.UserServiceGUIDDockerLabelKey] = serviceUuidLabelValue
	labels[docker_label_key.VolumeTypeDockerLabelKey] = label_value_consts.FilesArtifactExpansionVolumeTypeDockerLabelValue
	// TODO Create a KurtosisResourceDockerLabelKey object, like Kubernetes, and apply the "user-service" label here?

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

// In Docker we get one volume per persistent directory
func (provider *dockerEnclaveObjectAttributesProviderImpl) ForSinglePersistentDirectoryVolume(
	persistentKey service_directory.DirectoryPersistentKey,
) (
	DockerObjectAttributes,
	error,
) {
	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the persistent directory volume for persistentKey '%v'", persistentKey)
	}

	name, err := provider.getNameForEnclaveObject([]string{
		string(persistentKey),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the persistent volume name object using GUID '%v' and persistent key '%v'", guidStr, persistentKey)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for files artifact expansion volume with UUID '%v'", guidStr)
	}

	labels[docker_label_key.VolumeTypeDockerLabelKey] = label_value_consts.PersistentDirectoryVolumeTypeDockerLabelValue
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
	serviceUUID service.ServiceUUID,
) (
	DockerObjectAttributes,
	error,
) {
	serviceUuidStr := string(serviceUUID)

	guidStr, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating a UUID for the files artifacts expander container for service '%v'", serviceUuidStr)
	}

	name, err := provider.getNameForEnclaveObject([]string{
		artifactsExpanderContainerNameFragment,
		guidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifacts expander container name with UUID '%v'", guidStr)
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for files artifacts expander container with UUID '%v'", guidStr)
	}

	serviceUuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceUuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from service GUID string '%v'", serviceUuidStr)
	}
	labels[docker_label_key.UserServiceGUIDDockerLabelKey] = serviceUuidLabelValue
	labels[docker_label_key.ContainerTypeDockerLabelKey] = label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue

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

	labels[docker_label_key.ContainerTypeDockerLabelKey] = label_value_consts.LogsCollectorTypeDockerLabelValue
	labels[docker_label_key.PortSpecsDockerLabelKey] = serializedPortsSpec

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

	labels[docker_label_key.VolumeTypeDockerLabelKey] = label_value_consts.LogsCollectorVolumeTypeDockerLabelValue

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
	elems = append(elems, provider.enclaveId.GetString())
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

// Gets the name of the service container (service_name--service_uuid)
func (provider *dockerEnclaveObjectAttributesProviderImpl) getNameForUserServiceContainer(serviceName service.ServiceName, serviceUuid service.ServiceUUID) (*docker_object_name.DockerObjectName, error) {
	nameStr := strings.Join(
		[]string{
			string(serviceName), string(serviceUuid),
		},
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
		docker_label_key.EnclaveUUIDDockerLabelKey:     provider.enclaveId,
		docker_label_key.LogsEnclaveUUIDDockerLabelKey: provider.enclaveId,
	}
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithGUID(guid string) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels := provider.getLabelsForEnclaveObject()
	guidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from GUID string '%v'", guid)
	}
	shortGuidStr := uuid_generator.ShortenedUUIDString(guid)
	shortGuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(shortGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a short GUID Docker label value from GUID string '%v'", guid)
	}
	labels[docker_label_key.GUIDDockerLabelKey] = guidLabelValue
	labels[docker_label_key.LogsServiceUUIDDockerLabelKey] = guidLabelValue
	labels[docker_label_key.LogsServiceShortUUIDDockerLabelKey] = shortGuidLabelValue
	return labels, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObjectWithIDAndGUID(id, guid string) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guid)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the enclave object labels with UUID '%v'", guid)
	}
	idLabelValue, err := docker_label_value.CreateNewDockerLabelValue(id)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from ID string '%v'", id)
	}
	labels[docker_label_key.IDDockerLabelKey] = idLabelValue
	labels[docker_label_key.LogsServiceNameDockerLabelKey] = idLabelValue
	return labels, nil
}

// Return Traefik labels
// Including the labels required to route traffic to the user service ports based on the Host header:
// <port number>-<service short uuid>-<enclave short uuid>
// The Traefik service name format is: <enclave short uuid>-<service short uuid>-<port number>
// With the following input:
//
//	Enclave short UUID: 65d2fb6d6732
//	Service short UUID: 3771c85af16a
//	HTTP Port 1 number: 80
//	HTTP Port 2 number: 81
//
// the following labels are returned:
//
//	"traefik.enable": "true",
//	"traefik.http.routers.65d2fb6d6732-3771c85af16a-80.rule": "Host(`80-3771c85af16a-65d2fb6d6732`)",
//	"traefik.http.routers.65d2fb6d6732-3771c85af16a-80.service": "65d2fb6d6732-3771c85af16a-80",
//	"traefik.http.services.65d2fb6d6732-3771c85af16a-80.loadbalancer.server.port": "80"
//	"traefik.http.routers.65d2fb6d6732-3771c85af16a-80.rule": "Host(`81-3771c85af16a-65d2fb6d6732`)",
//	"traefik.http.routers.65d2fb6d6732-3771c85af16a-81.service": "65d2fb6d6732-3771c85af16a-81",
//	"traefik.http.services.65d2fb6d6732-3771c85af16a-81.loadbalancer.server.port": "81"
func (provider *dockerEnclaveObjectAttributesProviderImpl) getTraefikLabelsForEnclaveObject(serviceUuid string, ports map[string]*port_spec.PortSpec) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{}
	labelKeyValuePairs := map[string]string{}

	for _, portSpec := range ports {
		maybeApplicationProtocol := ""
		if portSpec.GetMaybeApplicationProtocol() != nil {
			maybeApplicationProtocol = *portSpec.GetMaybeApplicationProtocol()
		}
		if maybeApplicationProtocol == consts.HttpApplicationProtocol {
			shortEnclaveUuid := uuid_generator.ShortenedUUIDString(provider.enclaveId.GetString())
			shortServiceUuid := uuid_generator.ShortenedUUIDString(serviceUuid)
			servicePortStr := fmt.Sprintf("%s-%s-%d", shortEnclaveUuid, shortServiceUuid, portSpec.GetNumber())

			labelKeyValuePairs[fmt.Sprintf("http.routers.%s.rule", servicePortStr)] = fmt.Sprintf("Host(`%d-%s-%s`)", portSpec.GetNumber(), shortServiceUuid, shortEnclaveUuid)
			labelKeyValuePairs[fmt.Sprintf("http.routers.%s.service", servicePortStr)] = servicePortStr
			labelKeyValuePairs[fmt.Sprintf("http.services.%s.loadbalancer.server.port", servicePortStr)] = strconv.Itoa(int(portSpec.GetNumber()))
		}
	}

	if len(labelKeyValuePairs) > 0 {
		labelKeyValuePairs["enable"] = "true"
	}

	for key, value := range labelKeyValuePairs {
		labelKey, err := docker_label_key.CreateNewDockerTraefikLabelKey(key)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred getting the traefik  label key with suffix '%v'", key)
		}
		labelValue, err := docker_label_value.CreateNewDockerLabelValue(value)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the traefik  label value with value '%v'", value)
		}
		labels[labelKey] = labelValue
	}

	return labels, nil
}

func getLabelKeyValuesAsStrings(labels map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue) map[string]string {
	result := map[string]string{}
	for key, value := range labels {
		result[key.GetString()] = value.GetString()
	}
	return result
}
