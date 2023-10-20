package kubernetes_label_key

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	// which will cause a resource leak on the user's system!
	//
	// If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	// These immutable values track resources between Kurtosis versions.
	kurtosisDomain               = "kurtosistech.com"
	labelKeyPrefixStr            = kurtosisDomain + "/"
	appIdLabelKeyStr             = labelKeyPrefixStr + "app-id"
	resourceTypeLabelKeyStr      = labelKeyPrefixStr + "resource-type"
	customUserLabelsKeyPrefixStr = kurtosisDomain + ".custom/"
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	volumeTypeLabelKeyStr = labelKeyPrefixStr + "volume-type"

	// A label to identify a Kurtosis resource (e.g. network, container, etc.) by its id
	idLabelKeyStr = labelKeyPrefixStr + "id"

	// Used for things like service GUID, etc.
	guidLabelKeyStr = labelKeyPrefixStr + "guid"

	enclaveIdLabelKeyStr = labelKeyPrefixStr + "enclave-id"
	// TODO deprecate this in favor of storing in DB
	enclaveNameLabelKeyStr = labelKeyPrefixStr + "enclave-name"

	// As of 2022-05-17, these get attached to files artifact expansion volumes
	userServiceGuidKeyStr = labelKeyPrefixStr + "user-service-guid"
)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// If these value change, it will lead to the Kurtosis engine losing track of old resources
//
//	which will cause a resource leak on the user's cluster!
//
//	 If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
var AppIDKubernetesLabelKey = MustCreateNewKubernetesLabelKey(appIdLabelKeyStr)
var KurtosisResourceTypeKubernetesLabelKey = MustCreateNewKubernetesLabelKey(resourceTypeLabelKeyStr)

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

var KurtosisVolumeTypeKubernetesLabelKey = MustCreateNewKubernetesLabelKey(volumeTypeLabelKeyStr)
var IDKubernetesLabelKey = MustCreateNewKubernetesLabelKey(idLabelKeyStr)
var GUIDKubernetesLabelKey = MustCreateNewKubernetesLabelKey(guidLabelKeyStr)
var EnclaveUUIDKubernetesLabelKey = MustCreateNewKubernetesLabelKey(enclaveIdLabelKeyStr)
var EnclaveNameKubernetesLabelKey = MustCreateNewKubernetesLabelKey(enclaveNameLabelKeyStr)
var UserServiceGUIDKubernetesLabelKey = MustCreateNewKubernetesLabelKey(userServiceGuidKeyStr)
