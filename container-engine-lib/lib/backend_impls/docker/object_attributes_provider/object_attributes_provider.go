package object_attributes_provider

import (
	"fmt"
	"strconv"
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
	reverseProxyName       = "kurtosis-reverse-proxy"
	engineRESTAPIPortStr   = "engine-rest-api"
)

type DockerObjectAttributesProvider interface {
	ForEngineServer(
		guid engine.EngineGUID,
		grpcPortId string,
		grpcPortSpec *port_spec.PortSpec,
		restAPIPortId string,
		restAPIPortSpec *port_spec.PortSpec,
	) (DockerObjectAttributes, error)
	ForEnclave(enclaveUuid enclave.EnclaveUUID) (DockerEnclaveObjectAttributesProvider, error)
	ForLogsAggregator() (DockerObjectAttributes, error)
	ForLogsStorageVolume() (DockerObjectAttributes, error)
	ForReverseProxy() (DockerObjectAttributes, error)
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
	restAPIPortId string,
	restAPIPortSpec *port_spec.PortSpec,
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

	traefikLabels, err := provider.getTraefikLabelsForEngine(restAPIPortSpec)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting traefik labels for engine server")
	}
	for traefikLabelKey, traefikLabelValue := range traefikLabels {
		labels[traefikLabelKey] = traefikLabelValue
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

func (provider *dockerObjectAttributesProviderImpl) ForReverseProxy() (DockerObjectAttributes, error) {
	name, err := docker_object_name.CreateNewDockerObjectName(reverseProxyName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating a Docker object name object from string '%v'", reverseProxyName)
	}

	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{
		docker_label_key.ContainerTypeDockerLabelKey: label_value_consts.ReverseProxyTypeDockerLabelValue,
	}

	objectAttributes, err := newDockerObjectAttributesImpl(name, labels)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating the ObjectAttributesImpl with the name '%s' and labels '%+v'", name, labels)
	}
	return objectAttributes, nil
}

// Return Traefik labels
// Including the labels required to route traffic to the engine rest api port if the Host header is set to "engine".
//
//	"traefik.enable": "true",
//	"traefik.http.routers.engine-rest-api.rule": "Host(`engine`)",
//	"traefik.http.routers.engine-rest-api.service": "engine-rest-api",
//	"traefik.http.services.engine-rest-api.loadbalancer.server.port": "<engine rest api port number>"
func (provider *dockerObjectAttributesProviderImpl) getTraefikLabelsForEngine(restAPIPortSpec *port_spec.PortSpec) (map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue, error) {
	labels := map[*docker_label_key.DockerLabelKey]*docker_label_value.DockerLabelValue{}

	// Header Host rule
	ruleKeySuffix := fmt.Sprintf("http.routers.%s.rule", engineRESTAPIPortStr)
	ruleLabelKey, err := docker_label_key.CreateNewDockerTraefikLabelKey(ruleKeySuffix)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the traefik rule label key with suffix '%v'", ruleKeySuffix)
	}
	ruleValue := fmt.Sprintf("Host(`%s`)", engine.RESTAPIPortHostHeader)
	ruleLabelValue, err := docker_label_value.CreateNewDockerLabelValue(ruleValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the traefik rule label value with value '%v'", ruleValue)
	}
	labels[ruleLabelKey] = ruleLabelValue

	// Service name
	serviceKeySuffix := fmt.Sprintf("http.routers.%s.service", engineRESTAPIPortStr)
	serviceLabelKey, err := docker_label_key.CreateNewDockerTraefikLabelKey(serviceKeySuffix)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the traefik service label key with suffix '%v'", serviceKeySuffix)
	}
	serviceValue := engineRESTAPIPortStr
	serviceLabelValue, err := docker_label_value.CreateNewDockerLabelValue(serviceValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the traefik service label value with value '%v'", serviceValue)
	}
	labels[serviceLabelKey] = serviceLabelValue

	// Service port number
	portKeySuffix := fmt.Sprintf("http.services.%s.loadbalancer.server.port", engineRESTAPIPortStr)
	portLabelKey, err := docker_label_key.CreateNewDockerTraefikLabelKey(portKeySuffix)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the traefik port label key with suffix '%v'", portKeySuffix)
	}
	portValue := strconv.Itoa(int(restAPIPortSpec.GetNumber()))
	portLabelValue, err := docker_label_value.CreateNewDockerLabelValue(portValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the traefik port label value with value '%v'", portValue)
	}
	labels[portLabelKey] = portLabelValue

	// Enable Traefik
	traefikEnableLabelKey, err := docker_label_key.CreateNewDockerTraefikLabelKey("enable")
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the traefik enable label key")
	}
	traefikEnableValue := "true"
	traefikEnableLabelValue, err := docker_label_value.CreateNewDockerLabelValue(traefikEnableValue)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the traefik enable label value with value '%v'", traefikEnableValue)
	}
	labels[traefikEnableLabelKey] = traefikEnableLabelValue

	return labels, nil
}
