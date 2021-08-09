/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package object_name_providers

import "strings"

const (
	metadataAcquisitionContainerNameLabel = "metadata-acquisition"
)


// This struct is responsible for providing names to the objects used in the testing framework
type TestsuiteExecutionObjectNameProvider struct {
	executionId string // Execution ID identifying a run of a testsuite
}

func NewTestsuiteExecutionObjectNameProvider(executionId string) *TestsuiteExecutionObjectNameProvider {
	return &TestsuiteExecutionObjectNameProvider{executionId: executionId}
}

func (namer *TestsuiteExecutionObjectNameProvider) ForMetadataAcquiringTestsuiteContainer() string {
	return strings.Join(
		[]string{namer.executionId, metadataAcquisitionContainerNameLabel, testsuiteContainerNameSuffix},
		objectNameElementSeparator,
	)
}

func (namer *TestsuiteExecutionObjectNameProvider) ForTestEnclave(testName string) (string, *EnclaveObjectNameProvider) {
	enclaveId := strings.Join(
		[]string{namer.executionId, testName},
		objectNameElementSeparator,
	)
	enclaveObjectNameProvider := NewEnclaveObjectNameProvider(enclaveId)
	return enclaveId, enclaveObjectNameProvider
}
