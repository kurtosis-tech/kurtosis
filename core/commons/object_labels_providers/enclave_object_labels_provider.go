/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package object_labels_providers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/module_store/module_store_types"
	"github.com/kurtosis-tech/kurtosis-core/api_container/server/service_network/service_network_types"
	"github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"
	"net"
)

type EnclaveObjectLabelsProvider struct {
	enclaveId string
}

func NewEnclaveObjectLabelsProvider(enclaveId string) *EnclaveObjectLabelsProvider {
	return &EnclaveObjectLabelsProvider{enclaveId: enclaveId}
}

// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// Be VERY careful modifying these! If you add a new label here, it's possible to leak Kurtosis resources:
//  1) the user creates an enclave using the old engine, and the network & volume get the old labels
//  2) the user upgrades their CLI, and restarts with the new engine
//  3) the new engine searches for enclaves/volumes using the new labels, and doesn't find the old network/volume
// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (labelsProvider *EnclaveObjectLabelsProvider) ForEnclaveNetwork() map[string]string {
	return labelsProvider.getLabelsForEnclaveObject()
}

// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// Be VERY careful modifying these! If you add a new label here, it's possible to leak Kurtosis resources:
//  1) the user creates an enclave using the old engine, and the network & volume get the old labels
//  2) the user upgrades their CLI, and restarts with the new engine
//  3) the new engine searches for enclaves/volumes using the new labels, and doesn't find the old network/volume
// !!!!!!!!!!!!!!!!!!! WARNING WARNING WARNING WARNING WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func (labelsProvider *EnclaveObjectLabelsProvider) ForEnclaveDataVolume() map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForApiContainer(
	ipAddr net.IP,
) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeAPIContainer
	labels[enclave_object_labels.APIContainerIPLabel] = ipAddr.String()
	labels[enclave_object_labels.APIContainerPortNumLabel] = fmt.Sprintf("%v", kurtosis_core_rpc_api_consts.ListenPort)
	labels[enclave_object_labels.APIContainerPortProtocolLabel] = kurtosis_core_rpc_api_consts.ListenProtocol
	return labels
}

// TODO We don't want testsuites to be special - they should be Just Another Kurtosis Module - but we can't make them
//  unspecial (and thus delete this method) until the API container supports a container log-streaming endpoint
func (labelsProvider *EnclaveObjectLabelsProvider) ForTestRunningTestsuiteContainer() map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeTestsuiteContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForUserServiceContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObjectWithGUID(string(serviceGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeUserServiceContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForNetworkingSidecarContainer(serviceGUID service_network_types.ServiceGUID) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObjectWithGUID(string(serviceGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeNetworkingSidecarContainer
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForModuleContainer(moduleGUID module_store_types.ModuleGUID) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObjectWithGUID(string(moduleGUID))
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeModuleContainer
	return labels
}

// This is a liiiiittle strange to have here, since this is only used by the CLI (not anything inside this repo)
func (labelsProvider *EnclaveObjectLabelsProvider) ForInteractiveREPLContainer(interactiveReplGuid string) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObjectWithGUID(interactiveReplGuid)
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeInteractiveREPL
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForFilesArtifactExpanderContainer() map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeFilesArtifactExpander
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) ForFilesArtifactExpansionVolume() map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsForEnclaveObject() map[string]string {
	labels := getLabelsForKurtosisObject()
	labels[enclave_object_labels.EnclaveIDContainerLabel] = labelsProvider.enclaveId
	return labels
}

func (labelsProvider *EnclaveObjectLabelsProvider) getLabelsForEnclaveObjectWithGUID(guid string) map[string]string {
	labels := labelsProvider.getLabelsForEnclaveObject()
	labels[enclave_object_labels.GUIDLabel] = guid
	return labels
}
