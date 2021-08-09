/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package object_name_providers

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_store_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"strings"
	"time"
)

const (

	// These should represent the same format of YYYY-MM-DDTHH.mm.ss.SSS
	goTimestampFormat = "2006-01-02T15.04.05.000"
)

type EnclaveObjectNameProvider struct {
	enclaveId string
}

func NewEnclaveObjectNameProvider(enclaveId string) *EnclaveObjectNameProvider {
	return &EnclaveObjectNameProvider{enclaveId: enclaveId}
}

func (nameProvider *EnclaveObjectNameProvider) ForApiContainer() string {
	return strings.Join(
		[]string{nameProvider.enclaveId, apiContainerNameSuffix},
		objectNameElementSeparator,
	)
}

func (nameProvider *EnclaveObjectNameProvider) ForTestRunningTestsuiteContainer() string {
	return strings.Join(
		[]string{nameProvider.enclaveId, testsuiteContainerNameSuffix},
		objectNameElementSeparator,
	)
}

func (nameProvider *EnclaveObjectNameProvider) ForUserServiceContainer(serviceId service_network_types.ServiceID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		userServiceContainerNameLabel,
		string(serviceId),
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForNetworkingSidecarContainer(serviceIdSidecarAttachedTo service_network_types.ServiceID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		string(serviceIdSidecarAttachedTo),
		networkingSidecarContainerNameSuffix,
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForFilesArtifactExpanderContainer(serviceId service_network_types.ServiceID, artifactId string) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		string(serviceId),
		artifactExpanderContainerNameLabel,
		artifactId,
		time.Now().Format(goTimestampFormat), // We add this timestamp so that if the same artifact for the same service ID expanded twice, we won't get collisions
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForFilesArtifactExpansionVolume(serviceId string, artifactId string) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		artifactExpansionVolumeNameLabel,
		serviceId,
		artifactId,
		time.Now().Format(goTimestampFormat), // We add this timestamp so that if the same artifact for the same service ID expanded twice, we won't get collisions
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForLambdaContainer(lambdaId lambda_store_types.LambdaID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		lambdaContainerNameLabel,
		string(lambdaId),
	})
}


func (nameProvider *EnclaveObjectNameProvider) combineElementsWithEnclaveId(elems []string) string {
	toJoin := []string{nameProvider.enclaveId}
	toJoin = append(toJoin, elems...)
	return strings.Join(
		toJoin,
		objectNameElementSeparator,
	)
}