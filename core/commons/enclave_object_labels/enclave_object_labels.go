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

	// TODO RENAME THIS!!! It's being used for more than just containers (also networks & volumes)
	//  What we should really do though is have a separate type per object type (e.g. ContainerLabels, VolumeLabels, etc.)
	//  and then just duplicate things like "enclave-id" that are duplicated, because even though they have the same value
	//  they're actually entirely distinct, different labels
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
	ContainerTypeFilesArtifactExpander      = "files-artifact-expander"

	// Testsuite type label + values
	TestsuiteTypeLabelKey = labelNamespace + "testsuite-type"
	TestsuiteTypeLabelValue_MetadataAcquisition = "metadata-acquisition"
	TestsuiteTypeLabelValue_TestRunning = "test-running"
)