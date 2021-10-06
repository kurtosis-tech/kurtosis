/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package object_labels_providers

import (
	"github.com/kurtosis-tech/kurtosis/api_container/server/lambda_store/lambda_store_types"
	"github.com/kurtosis-tech/kurtosis/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_object_labels"
	"net"
	"strconv"
)

type EnclaveObjectLabelsProvider struct {
	enclaveId string
}

func NewEnclaveObjectLabelsProvider(enclaveId string) *EnclaveObjectLabelsProvider {
	return &EnclaveObjectLabelsProvider{enclaveId: enclaveId}
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForApiContainer(apiContainerIPAddress net.IP, apiContainerListenPort uint) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeApiContainer
	labels[enclave_object_labels.ApiContainerURLLabel] = apiContainerIPAddress.String() + strconv.FormatUint(uint64(apiContainerListenPort), 10)
	return labels
}

// TODO We don't want testsuites to be special - they should be Just Another Kurtosis Module - but we can't make them
//  unspecial (and thus delete this method) until the API container supports a container log-streaming endpoint
func (labelsProvider *EnclaveObjectLabelsProvider) ForTestRunningTestsuiteContainer() map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeTestsuiteContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForUserServiceContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(serviceGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForNetworkingSidecarContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(serviceGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeNetworkingSidecarContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForLambdaContainer(lambdaGUID lambda_store_types.LambdaGUID) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabelsAndGUIDLabel(string(lambdaGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeLambdaContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsMapWithCommonsLabels() map[string]string {
	labels := map[string]string{}
	labels[enclave_object_labels.EnclaveIDContainerLabel] = labelsProvider.enclaveId
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsMapWithCommonsLabelsAndGUIDLabel(guid string) map[string]string {
	labels := labelsProvider.getLabelsMapWithCommonsLabels()
	labels[enclave_object_labels.GUIDLabel] = guid
	return labels
}
