package docker

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
)
var AppIDLabelKey = MustCreateNewDockerLabelKey(appIdLabelKeyStr)
var IDLabelKey = MustCreateNewDockerLabelKey(idLabelKeyStr)
var GUIDLabelKey = MustCreateNewDockerLabelKey(guidLabelKeyStr)
var ContainerTypeLabelKey = MustCreateNewDockerLabelKey(containerTypeLabelKeyStr)
