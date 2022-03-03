package label_key_consts

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	//  which will cause a resource leak on the user's system!
	//
	//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	//
	labelNamespaceStr = "com.kurtosistech."
	appIdLabelKeyStr = labelNamespaceStr + "app-id"
	containerTypeLabelKeyStr = labelNamespaceStr + "container-type"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	// A label to identify a Kurtosis resource (e.g. network, container, etc.) by its id
	idLabelKeyStr = labelNamespaceStr + "id"

	// Used for things like service GUID, module GUID, etc.
	guidLabelKeyStr = labelNamespaceStr + "guid"

	portSpecsLabelKeyStr = labelNamespaceStr + "ports"
)
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//  which will cause a resource leak on the user's system!
//
//   If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
//
var AppIDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(appIdLabelKeyStr)
var ContainerTypeLabelKey = docker_label_key.MustCreateNewDockerLabelKey(containerTypeLabelKeyStr)
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var IDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(idLabelKeyStr)
var GUIDLabelKey = docker_label_key.MustCreateNewDockerLabelKey(guidLabelKeyStr)
var PortSpecsLabelKey = docker_label_key.MustCreateNewDockerLabelKey(portSpecsLabelKeyStr)
