/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package object_labels_providers

import "github.com/kurtosis-tech/kurtosis-core/commons/enclave_object_labels"

// TODO Move this to the CLI, which contains the testing machinery
// This struct is responsible for providing labels to the objects used in the testing framework
type TestsuiteExecutionObjectLabelsProvider struct {
	executionId string // Execution ID identifying a run of a testsuite
}

func NewTestsuiteExecutionObjectLabelsProvider(executionId string) *TestsuiteExecutionObjectLabelsProvider {
	return &TestsuiteExecutionObjectLabelsProvider{executionId: executionId}
}


func (provider *TestsuiteExecutionObjectLabelsProvider) ForMetadataAcquiringTestsuiteContainer() map[string]string {
	labels := getLabelsForKurtosisObject()
	labels[enclave_object_labels.ContainerTypeLabel] = enclave_object_labels.ContainerTypeTestsuiteContainer
	labels[enclave_object_labels.TestsuiteTypeLabelKey] = enclave_object_labels.TestsuiteTypeLabelValue_MetadataAcquisition
	return labels
}
