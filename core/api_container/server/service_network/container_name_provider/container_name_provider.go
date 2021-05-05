/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package container_name_provider

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
)

const (
	networkingSidecarNameSuffix = "networking-sidecar"
	artifactExpanderNameLabel   = "files-artifact-expander"
)

type ContainerNameElementsProvider struct {
	prefixElems []string
}

func NewContainerNameElementsProvider(prefixElems []string) *ContainerNameElementsProvider {
	return &ContainerNameElementsProvider{prefixElems: prefixElems}
}

func (provider ContainerNameElementsProvider) GetForUserService(serviceId service_network_types.ServiceID) []string {
	return provider.addPrefix([]string{
		string(serviceId),
	})
}

func (provider ContainerNameElementsProvider) GetForNetworkingSidecar(serviceIdSidecarAttachedTo service_network_types.ServiceID) []string {
	return provider.addPrefix([]string{
		string(serviceIdSidecarAttachedTo),
		networkingSidecarNameSuffix,
	})
}

func (provider ContainerNameElementsProvider) GetForFilesArtifactExpander(serviceId service_network_types.ServiceID, artifactUrlHash string) []string {
	return provider.addPrefix([]string{
		string(serviceId),
		artifactExpanderNameLabel,
		artifactUrlHash,
	})
}

func (provider ContainerNameElementsProvider) addPrefix(toElems []string) []string {
	return append(provider.prefixElems, toElems...)
}