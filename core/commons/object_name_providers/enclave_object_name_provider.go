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
	uniqueTimestampFormat = "2006-01-02T15.04.05.000"
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

// TODO We don't want testsuites to be special - they should be Just Another Kurtosis Module - but we can't make them
//  unspecial (and thus delete this method) until the API container supports a container log-streaming endpoint
func (nameProvider *EnclaveObjectNameProvider) ForTestRunningTestsuiteContainer() string {
	return strings.Join(
		[]string{nameProvider.enclaveId, testsuiteContainerNameSuffix},
		objectNameElementSeparator,
	)
}

func (nameProvider *EnclaveObjectNameProvider) ForUserServiceContainer(serviceGUID service_network_types.ServiceGUID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		userServiceContainerNameLabel,
		string(serviceGUID),
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForNetworkingSidecarContainer(serviceGUIDSidecarAttachedTo service_network_types.ServiceGUID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{// TODO Switch order of these
		string(serviceGUIDSidecarAttachedTo),
		networkingSidecarContainerNameSuffix,
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForFilesArtifactExpanderContainer(serviceGUID service_network_types.ServiceGUID, artifactId string) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		string(serviceGUID),
		artifactExpanderContainerNameLabel,
		artifactId,
		time.Now().Format(uniqueTimestampFormat), // We add this timestamp so that if the same artifact for the same service GUID expanded twice, we won't get collisions
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForFilesArtifactExpansionVolume(serviceGUID service_network_types.ServiceGUID, artifactId string) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		artifactExpansionVolumeNameLabel,
		string(serviceGUID),
		artifactId,
		time.Now().Format(uniqueTimestampFormat), // We add this timestamp so that if the same artifact for the same service GUID expanded twice, we won't get collisions
	})
}

func (nameProvider *EnclaveObjectNameProvider) ForLambdaContainer(lambdaGUID lambda_store_types.LambdaGUID) string {
	return nameProvider.combineElementsWithEnclaveId([]string{
		lambdaContainerNameLabel,
		string(lambdaGUID),
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
