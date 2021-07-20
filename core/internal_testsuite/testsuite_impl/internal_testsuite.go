/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package testsuite_impl

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/advanced_network_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/basic_datastore_and_api_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/basic_datastore_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/bulk_command_execution_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/exec_command_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/files_artifact_mounting_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/lambda_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/local_static_file_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/network_partition_test"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/static_file_consts"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/wait_for_endpoint_availability_test"
	"path"
)

const (

	// Directory where static files live inside the testsuite container
	staticFilesDirpath = "/static-files"
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
		"localStaticFileTest": local_static_file_test.LocalStaticFileTest{},
		"bulkCommandExecutionTest": bulk_command_execution_test.NewBulkCommandExecutionTest(suite.datastoreServiceImage),
		"lambdaTest": lambda_test.LambdaTest{},
	}
	return tests
}

func (suite InternalTestsuite) GetNetworkWidthBits() uint32 {
	return 8
}

func (suite InternalTestsuite) GetStaticFiles() map[services.StaticFileID]string {
	return map[services.StaticFileID]string{
		static_file_consts.TestStaticFile1ID: path.Join(staticFilesDirpath, static_file_consts.TestStaticFile1Filename),
		static_file_consts.TestStaticFile2ID: path.Join(staticFilesDirpath, static_file_consts.TestStaticFile2Filename),
	}
}



