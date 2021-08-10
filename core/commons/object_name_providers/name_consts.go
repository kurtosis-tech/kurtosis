/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package object_name_providers

const (
	objectNameElementSeparator = "__"

	// The name that a testsuite container will receive
	testsuiteContainerNameSuffix = "testsuite"

	apiContainerNameSuffix       = "kurtosis-api"
	userServiceContainerNameLabel        = "user-service"
	networkingSidecarContainerNameSuffix = "networking-sidecar"
	artifactExpanderContainerNameLabel   = "files-artifact-expander"
	artifactExpansionVolumeNameLabel     = "files-artifact-expansion"
	lambdaContainerNameLabel             = "lambda"
)
