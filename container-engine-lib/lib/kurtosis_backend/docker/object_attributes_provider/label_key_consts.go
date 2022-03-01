package object_attributes_provider

import "github.com/kurtosis-tech/container-engine-lib/lib/kurtosis_backend/docker/object_attributes_provider/docker_label_key"

const (
	// TODO MOVE THIS TO FOREVER CONSTANTS!!!
	labelNamespaceStr = "com.kurtosistech."

	// TODO MOVE TO FOREVER CONSTANTS!!!
	appIdLabelKeyStr = labelNamespaceStr + "app-id"

	// A label to identify a Kurtosis resource (e.g. network, container, etc.) by its id
	idLabelKeyStr = labelNamespaceStr + "id"

	// Used for things like service GUID, module GUID, etc.
	guidLabelKeyStr = labelNamespaceStr + "guid"

	// TODO MOVE THIS TO FOREVER CONSTANTS!!!!
	containerTypeLabelKeyStr = labelNamespaceStr + "container-type"

	portSpecsLabelKeyStr = labelNamespaceStr + "ports"
)
var AppIDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(appIdLabelKeyStr)
var IDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(idLabelKeyStr)
var GUIDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(guidLabelKeyStr)
var ContainerTypeLabelKey = docker_label_key.MustCreateNewDockerLabelKey(containerTypeLabelKeyStr)
var PortSpecsLabelKey = docker_label_key.MustCreateNewDockerLabelKey(portSpecsLabelKeyStr)
