/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package container_name_provider

/*
const (
	userServiceNameLabel        = "user-service"
	networkingSidecarNameSuffix = "networking-sidecar"
	artifactExpanderNameLabel   = "files-artifact-expander"
	moduleNameLabel             = "module"
)

type ContainerNameElementsProvider struct {
	prefixElems []string
}

func NewContainerNameElementsProvider(prefixElems []string) *ContainerNameElementsProvider {
	return &ContainerNameElementsProvider{prefixElems: prefixElems}
}

func (provider ContainerNameElementsProvider) GetForUserService(serviceId service_network_types.ServiceID) []string {
	return provider.addPrefix([]string{
		userServiceNameLabel,
		string(serviceId),
	})
}

func (provider ContainerNameElementsProvider) GetForNetworkingSidecar(serviceIdSidecarAttachedTo service_network_types.ServiceID) []string {
	return provider.addPrefix([]string{
		string(serviceIdSidecarAttachedTo),
		networkingSidecarNameSuffix,
	})
}

func (provider ContainerNameElementsProvider) GetForFilesArtifactExpander(serviceId service_network_types.ServiceID, artifactId string) []string {
	return provider.addPrefix([]string{
		string(serviceId),
		artifactExpanderNameLabel,
		artifactId,
	})
}

func (provider ContainerNameElementsProvider) GetForModule(moduleId module_store_types.ModuleID) []string {
	return provider.addPrefix([]string{
		moduleNameLabel,
		string(moduleId),
	})
}

func (provider ContainerNameElementsProvider) addPrefix(toElems []string) []string {
	return append(provider.prefixElems, toElems...)
}


*/
