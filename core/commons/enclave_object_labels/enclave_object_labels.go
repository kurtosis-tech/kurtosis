/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_object_labels

const (
	labelNamespace = "com.kurtosistech."

	EnclaveIDContainerLabel = labelNamespace + "enclave-id"
	ContainerTypeLabel      = labelNamespace + "container-type"
	GUIDLabel               = labelNamespace + "guid"
	ApiContainerURLLabel    = labelNamespace + "api-container-url"

	ContainerTypeApiContainer               = "api-container"
	ContainerTypeTestsuiteContainer         = "testsuite"
	ContainerTypeUserServiceContainer       = "user-service"
	ContainerTypeNetworkingSidecarContainer = "networking-sidecar"
	ContainerTypeLambdaContainer            = "lambda"
)
