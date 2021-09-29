/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package object_labels_providers

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_store_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
)

const (
	labelNamespace = "com.kurtosistech."

	labelEnclaveIDKey     = labelNamespace + "enclave-id"
	labelContainerTypeKey = labelNamespace + "container-type"
	labelGUIDKey          = labelNamespace + "guid"

	containerTypeApiContainer               = "api-container"
	containerTypeTestsuiteContainer         = "testsuite"
	containerTypeUserServiceContainer       = "user-service"
	containerTypeNetworkingSidecarContainer = "networking-sidecar"
	containerTypeLambdaContainer            = "lambda"
)

type EnclaveObjectLabelsProvider struct {
	enclaveId string
}

func NewEnclaveObjectLabelsProvider(enclaveId string) *EnclaveObjectLabelsProvider {
	return &EnclaveObjectLabelsProvider{enclaveId: enclaveId}
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForApiContainer() map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[labelContainerTypeKey] = containerTypeApiContainer
	return labels
}

// TODO We don't want testsuites to be special - they should be Just Another Kurtosis Module - but we can't make them
//  unspecial (and thus delete this method) until the API container supports a container log-streaming endpoint
func (labelsProvider *EnclaveObjectLabelsProvider) ForTestRunningTestsuiteContainer() map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[labelContainerTypeKey] = containerTypeTestsuiteContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForUserServiceContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(serviceGUID))
	labels[labelContainerTypeKey] = containerTypeUserServiceContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForNetworkingSidecarContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(serviceGUID))
	labels[labelContainerTypeKey] = containerTypeNetworkingSidecarContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForLambdaContainer(lambdaGUID lambda_store_types.LambdaGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(lambdaGUID))
	labels[labelContainerTypeKey] = containerTypeLambdaContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsMapWithCommonsLabels() map[string]string {
	labels := map[string]string{}
	labels[labelEnclaveIDKey] = labelsProvider.enclaveId
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsMapWithCommonsLabelsAndGUIDLabel(guid string) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[labelGUIDKey] = guid
	return labels
}
