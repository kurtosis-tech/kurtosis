package label_key_consts

import (
	"github.com/kurtosis-tech/kurtosis/libs/container-engine-lib/lib/backend_impls/docker/object_attributes_provider/docker_label_key"
)

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	// which will cause a resource leak on the user's system!
	//
	// If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	// These immutable values track resources between Kurtosis versions.
	labelNamespaceStr = "com.kurtosistech."
	appIdLabelKeyStr  = labelNamespaceStr + "app-id"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	containerTypeLabelKeyStr = labelNamespaceStr + "container-type"
	volumeTypeLabelKeyStr    = labelNamespaceStr + "volume-type"

	// A label to identify a Kurtosis resource (e.g. network, container, etc.) by its id
	idLabelKeyStr = labelNamespaceStr + "id"

	// Used for things like service GUID, etc.
	guidLabelKeyStr = labelNamespaceStr + "guid"

	userServiceGuidDockerLabelKeyStr = labelNamespaceStr + "user-service-guid"

	portSpecsLabelKeyStr = labelNamespaceStr + "ports"

	enclaveIdLabelKeyStr = labelNamespaceStr + "enclave-id"
	// TODO deprecate this in favor of storing in DB
	enclaveNameLabelKeyStr = labelNamespaceStr + "enclave-name"

	enclaveCreationTime = labelNamespaceStr + "enclave-creation-time"

	isNetworkPartitioningEnabledKeyStr = labelNamespaceStr + "is-network-partitioning-enabled"

	privateIpAddrLabelKeyStr = labelNamespaceStr + "private-ip"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//
//	which will cause a resource leak on the user's system!
//
//	 If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var AppIDDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(appIdLabelKeyStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var ContainerTypeDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(containerTypeLabelKeyStr)
var VolumeTypeDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(volumeTypeLabelKeyStr)
var IDDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(idLabelKeyStr)
var GUIDDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(guidLabelKeyStr)
var PortSpecsDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(portSpecsLabelKeyStr)
var EnclaveUUIDDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(enclaveIdLabelKeyStr)
var EnclaveNameDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(enclaveNameLabelKeyStr)
var EnclaveCreationTimeLabelKey = docker_label_key.MustCreateNewDockerLabelKey(enclaveCreationTime)
var IsNetworkPartitioningEnabledDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(isNetworkPartitioningEnabledKeyStr)
var PrivateIPDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(privateIpAddrLabelKeyStr)
var UserServiceGUIDDockerLabelKey = docker_label_key.MustCreateNewDockerLabelKey(userServiceGuidDockerLabelKeyStr)
