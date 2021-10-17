/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/advanced_network_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/basic_datastore_and_api_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/basic_datastore_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/bulk_command_execution_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/exec_command_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/files_artifact_mounting_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/local_static_file_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/module_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/network_partition_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/test_internal_state_persistence_test"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/testsuite_impl/wait_for_endpoint_availability_test"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
)


type InternalTestsuite struct {
	apiServiceImage string
	datastoreServiceImage string
}

func NewInternalTestsuite(apiServiceImage string, datastoreServiceImage string) *InternalTestsuite {
	return &InternalTestsuite{apiServiceImage: apiServiceImage, datastoreServiceImage: datastoreServiceImage}
}

func (suite InternalTestsuite) GetTests() map[string]testsuite.Test {
	tests := map[string]testsuite.Test{
		"basicDatastoreTest": basic_datastore_test.NewBasicDatastoreTest(suite.datastoreServiceImage),
		"basicDatastoreAndApiTest": basic_datastore_and_api_test.NewBasicDatastoreAndApiTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		),
		"advancedNetworkTest": advanced_network_test.NewAdvancedNetworkTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		),
		"networkPartitionTest": network_partition_test.NewNetworkPartitionTest(
			suite.datastoreServiceImage,
			suite.apiServiceImage,
		),
		"filesArtifactMountingTest": files_artifact_mounting_test.FilesArtifactMountingTest{},
		"execCommandTest": exec_command_test.ExecCommandTest{},
		"waitForEndpointAvailabilityTest": wait_for_endpoint_availability_test.NewWaitForEndpointAvailabilityTest(
			suite.datastoreServiceImage,
		),
		"localStaticFileTest":              local_static_file_test.LocalStaticFileTest{},
		"bulkCommandExecutionTest":         bulk_command_execution_test.NewBulkCommandExecutionTest(suite.datastoreServiceImage),
		"moduleTest":                       module_test.ModuleTest{},
		"testInternalStatePersistenceTest": test_internal_state_persistence_test.NewTestInternalStatePersistenceTest(),
	}
	return tests
}
