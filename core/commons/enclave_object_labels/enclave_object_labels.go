/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package enclave_object_labels

const (
	labelNamespace = "com.kurtosistech."

	// This label will get created on every Kurtosis object, and has the same value every time
	// This allows us to easily filter down to only Kurtosis objects
	AppIDLabel = labelNamespace + "app-id"

	// This is the static value that every single Kurtosis object will receive for the app ID label
	AppIDValue = "kurtosis"

	EnclaveIDContainerLabel = labelNamespace + "enclave-id"
	ContainerTypeLabel      = labelNamespace + "container-type"

	// Used for things like service GUID, module GUID, etc.
	GUIDLabel               = labelNamespace + "guid"

	// A label for the API container IP address that it's running on
	APIContainerIPLabel    = labelNamespace + "api-container-ip"
	// Port number that the API container is listening on, INSIDE the network
	APIContainerPortNumLabel = labelNamespace + "api-container-port-number"
	// Protocol of the port that the API container is listening on
	APIContainerPortProtocolLabel = labelNamespace + "api-container-port-protocol"

	// Values for ContainerTypeLabel
	ContainerTypeAPIContainer               = "api-container"
	ContainerTypeTestsuiteContainer         = "testsuite"
	ContainerTypeUserServiceContainer       = "user-service"
	ContainerTypeNetworkingSidecarContainer = "networking-sidecar"
	ContainerTypeModuleContainer            = "module"
	// This is a little weird to have here because  this is only used by the CLI (which depends on this repo)
	ContainerTypeInteractiveREPL            = "interactive-repl"
)