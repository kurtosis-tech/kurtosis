package object_attributes_provider

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_value"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_object_name"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_port_spec_serializer"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_key_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/label_value_consts"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/module"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/stacktrace"
	"net"
	"strings"
	"time"
)

const (
	artifactExpansionObjectTimestampFormat = "2006-01-02T15.04.05.000"

	apiContainerNameSuffix                 = "kurtosis-api"
	userServiceContainerNameFragment       = "user-service"
	networkingSidecarContainerNameFragment = "networking-sidecar"
	artifactExpansionNameFragment         = "files-artifact-expansion"
	moduleContainerNameFragment           = "module"
)

type DockerEnclaveObjectAttributesProvider interface {
	ForEnclaveNetwork(isPartitioningEnabled bool) (DockerObjectAttributes, error)
	ForEnclaveDataVolume() (DockerObjectAttributes, error)

	ForApiContainer(
		ipAddr net.IP,
		privateGrpcPortId string,
		privateGrpcPortSpec *port_spec.PortSpec,
		privateGrpcProxyPortId string,
		privateGrpcProxyPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForUserServiceContainer(
		serviceId service.ServiceID,
		serviceGuid service.ServiceGUID,
		privateIpAddr net.IP,
		privatePorts map[string]*port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForNetworkingSidecarContainer(
		serviceGUIDSidecarAttachedTo service.ServiceGUID,
	) (DockerObjectAttributes, error)
	/*
	ForFilesArtifactExpansionContainer(
		fileArtifactExpansionGUID files_artifact_expansion.FilesArtifactExpansionGUID,
		serviceGUID service.ServiceGUID,
	) (DockerObjectAttributes, error)
	 */
	ForFilesArtifactsExpansionVolume(
		serviceGUID service.ServiceGUID,
	) (DockerObjectAttributes, error)
	ForModuleContainer(
		privateIpAddr net.IP,
		moduleID module.ModuleID,
		moduleGUID module.ModuleGUID,
		privatePortId string,
		privatePortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
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

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForEnclaveNetwork(isPartitioningEnabled bool) (DockerObjectAttributes, error) {
	enclaveIdStr := provider.enclaveId.GetString()
	name, err := docker_object_name.CreateNewDockerObjectName(enclaveIdStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a name object from string '%v'", enclaveIdStr)
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
	serviceId service.ServiceID,
	serviceGuid service.ServiceGUID,
	privateIpAddr net.IP,
	privatePorts map[string]*port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject(
		[]string{
			userServiceContainerNameFragment,
			string(serviceGuid),
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

	serviceIdStr := string(serviceId)
	serviceGuidStr := string(serviceGuid)

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(serviceIdStr, serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave object with GUID '%v'", serviceGuid)
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

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForNetworkingSidecarContainer(serviceGUIDSidecarAttachedTo service.ServiceGUID) (DockerObjectAttributes, error) {
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

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForModuleContainer(
	privateIpAddr net.IP,
	moduleID module.ModuleID,
	moduleGUID module.ModuleGUID,
	privatePortId string,
	privatePortSpec *port_spec.PortSpec,
) (DockerObjectAttributes, error) {
	name, err := provider.getNameForEnclaveObject([]string{
		moduleContainerNameFragment,
		string(moduleGUID),
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the module container name object")
	}

	privateIpLabelValue, err := docker_label_value.CreateNewDockerLabelValue(privateIpAddr.String())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a Docker label value object from module container private IP address '%v'",
			privateIpAddr.String(),
		)
	}

	usedPorts := map[string]*port_spec.PortSpec{
		privatePortId: privatePortSpec,
	}
	serializedPortsSpec, err := docker_port_spec_serializer.SerializePortSpecs(usedPorts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred serializing the following module container ports object to a string for storing in the ports label: %+v", usedPorts)
	}

	moduleIDStr := string(moduleID)
	moduleGUIDStr := string(moduleGUID)

	labels, err := provider.getLabelsForEnclaveObjectWithIDAndGUID(moduleIDStr, moduleGUIDStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the module labels using ID '%v' and GUID '%v'", moduleID, moduleGUID)
	}
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.ModuleContainerTypeDockerLabelValue
	labels[label_key_consts.PortSpecsDockerLabelKey] = serializedPortsSpec
	labels[label_key_consts.PrivateIPDockerLabelKey] = privateIpLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

func (provider *dockerEnclaveObjectAttributesProviderImpl) ForFilesArtifactsExpansionVolume(
	serviceGUID service.ServiceGUID,
)(
	DockerObjectAttributes,
	error,
){
	serviceGuidStr := string(serviceGUID)
	name, err := provider.getNameForEnclaveObject([]string{
		artifactExpansionNameFragment,
		serviceGuidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifact expansion volume name object for service '%v'", serviceGuidStr)
	}

	labels := provider.getLabelsForEnclaveObject()

	serviceGuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from sevice GUID string '%v'", serviceGuidStr)
	}
	labels[label_key_consts.UserServiceGUIDDockerLabelKey] = serviceGuidLabelValue
	labels[label_key_consts.VolumeTypeDockerLabelKey] = label_value_consts.FilesArtifactExpansionsVolumeTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}

/*
func (provider *dockerEnclaveObjectAttributesProviderImpl) ForFilesArtifactsExpansionContainer(
	guid files_artifact_expansion.FilesArtifactExpansionGUID,
	serviceGUID service.ServiceGUID,
)(
	DockerObjectAttributes,
	error,
) {
	guidStr := string(guid)
	serviceGuidStr := string(serviceGUID)
	name, err := provider.getNameForEnclaveObject([]string{
		artifactExpansionNameFragment,
		guidStr,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the files artifact expander container name object")
	}

	labels, err := provider.getLabelsForEnclaveObjectWithGUID(guidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting labels for enclave object with GUID '%v'", guid)
	}

	serviceGuidLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceGuidStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker label value from GUID string '%v'", serviceGuidStr)
	}
	labels[label_key_consts.UserServiceGUIDDockerLabelKey] = serviceGuidLabelValue
	labels[label_key_consts.ContainerTypeDockerLabelKey] = label_value_consts.FilesArtifactExpanderContainerTypeDockerLabelValue

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}

	return objectAttributes, nil
}
 */

// ====================================================================================================
//                                      Private Helper Functions
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


func (provider *dockerEnclaveObjectAttributesProviderImpl) getLabelsForEnclaveObject() map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue {
	return map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		label_key_consts.EnclaveIDDockerLabelKey: provider.enclaveId,
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

// Gets the name for an artifact expansion object (either volume or container)
func (provider *dockerEnclaveObjectAttributesProviderImpl) getArtifactExpansionObjectName(
	objectLabel string,
	forServiceGUID string,
	artifactId string,
) (*docker_object_name.DockerObjectName, error) {
	name, err := provider.getNameForEnclaveObject([]string{
		objectLabel,
		"for",
		forServiceGUID,
		"using",
		artifactId,
		"at",
		time.Now().Format(artifactExpansionObjectTimestampFormat), // We add this timestamp so that if the same artifact for the same service GUID expanded twice, we won't get collisions
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the artifact expansion object name")
	}
	return name, nil
}
