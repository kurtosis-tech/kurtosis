package docker_label_key

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	// which will cause a resource leak on the user's system!
	//
	// If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	// These immutable values track resources between Kurtosis versions.
	labelNamespaceStr            = "com.kurtosistech."
	appIdLabelKeyStr             = labelNamespaceStr + "app-id"
	customUserLabelsKeyPrefixStr = labelNamespaceStr + "custom."
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

	privateIpAddrLabelKeyStr = labelNamespaceStr + "private-ip"

	// We create a duplicate of the enclave uuid and service uuid label key because:
	// the logs aggregator (vector) needs the enclave uuid and service uuid label keys to create the filepath where logs are stored in persistent volume
	// but vectors template syntax can't interpret the "com.kurtosistech." prefix, so we can't use the existing label keys
	logsLabelPrefixStr                = "kurtosis_"
	logsOnlyEnclaveUuidLabelKeyStr    = logsLabelPrefixStr + "enclave_uuid"
	logsOnlyServiceUuidDockerLabelKey = logsLabelPrefixStr + "service_uuid"
	logsOnlyServiceNameDockerLabelKey = logsLabelPrefixStr + "service_name"

	// Traefik label keys
	traefikLabelKeyPrefixStr = "traefik."
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old containers
//
//	which will cause a resource leak on the user's system!
//
//	 If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var AppIDDockerLabelKey = MustCreateNewDockerLabelKey(appIdLabelKeyStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var ContainerTypeDockerLabelKey = MustCreateNewDockerLabelKey(containerTypeLabelKeyStr)
var VolumeTypeDockerLabelKey = MustCreateNewDockerLabelKey(volumeTypeLabelKeyStr)
var IDDockerLabelKey = MustCreateNewDockerLabelKey(idLabelKeyStr)
var GUIDDockerLabelKey = MustCreateNewDockerLabelKey(guidLabelKeyStr)
var PortSpecsDockerLabelKey = MustCreateNewDockerLabelKey(portSpecsLabelKeyStr)
var EnclaveUUIDDockerLabelKey = MustCreateNewDockerLabelKey(enclaveIdLabelKeyStr)
var EnclaveNameDockerLabelKey = MustCreateNewDockerLabelKey(enclaveNameLabelKeyStr)
var EnclaveCreationTimeLabelKey = MustCreateNewDockerLabelKey(enclaveCreationTime)
var PrivateIPDockerLabelKey = MustCreateNewDockerLabelKey(privateIpAddrLabelKeyStr)
var UserServiceGUIDDockerLabelKey = MustCreateNewDockerLabelKey(userServiceGuidDockerLabelKeyStr)
var LogsEnclaveUUIDDockerLabelKey = MustCreateNewDockerLabelKey(logsOnlyEnclaveUuidLabelKeyStr)
var LogsServiceUUIDDockerLabelKey = MustCreateNewDockerLabelKey(logsOnlyServiceUuidDockerLabelKey)
var LogsServiceNameDockerLabelKey = MustCreateNewDockerLabelKey(logsOnlyServiceNameDockerLabelKey)
