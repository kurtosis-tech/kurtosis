/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_object_labels

const (
	labelNamespace = "com.kurtosistech."

	EnclaveIDContainerLabel = labelNamespace + "enclave-id"
	ContainerTypeLabel      = labelNamespace + "container-type"

	// Used for things like service GUID, lambda GUID, etc.
	GUIDLabel               = labelNamespace + "guid"

	// A label for the API container IP address that it's running on
	APIContainerIPLabel    = labelNamespace + "api-container-ip"
	// A label for the API container port that it's running on
	APIContainerPortLabel    = labelNamespace + "api-container-port"

	// Values for ContainerTypeLabel
	ContainerTypeAPIContainer               = "api-container"
	ContainerTypeTestsuiteContainer         = "testsuite"
	ContainerTypeUserServiceContainer       = "user-service"
	ContainerTypeNetworkingSidecarContainer = "networking-sidecar"
	ContainerTypeLambdaContainer            = "lambda"
	// This is a little weird to have here because  this is only used by the CLI (which depends on this repo)
	ContainerTypeInteractiveREPL            = "interactive-repl"
)