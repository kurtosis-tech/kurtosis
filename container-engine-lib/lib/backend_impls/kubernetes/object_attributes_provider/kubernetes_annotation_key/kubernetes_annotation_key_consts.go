package kubernetes_annotation_key

const (
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// If these value change, it will lead to the Kurtosis engine losing track of old containers
	// which will cause a resource leak on the user's system!
	//
	// If you add new immutable values to this section, MAKE SURE TO UPDATE THE UNIT TEST!
	// These immutable values track resources between Kurtosis versions.
	keyNamespaceStr = "com.kurtosistech."
	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! DO NOT CHANGE THESE VALUES !!!!!!!!!!!!!!!!!!!!!!!!!!!!!

	portSpecsKubernetesAnnotationKeyStr = keyNamespaceStr + "ports"
)

var PortSpecsKubernetesAnnotationKey = MustCreateNewKubernetesAnnotationKey(portSpecsKubernetesAnnotationKeyStr)
