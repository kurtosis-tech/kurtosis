package startosis_upload_files_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "module"
	isPartitioningEnabled = false

	serviceId = "example-datastore-server-1"
	portId    = "grpc"

	pathToMountUploadedDir     = "/uploads"
	pathToCheckForUploadedFile = "/uploads/lib.star"

	startosisScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

DIR_TO_UPLOAD = "github.com/kurtosis-tech/datastore-army-module-demo/lib"
PATH_TO_MOUNT_UPLOADED_DIR = "` + pathToMountUploadedDir + `"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

uploaded_artifact_uuid = upload_files(DIR_TO_UPLOAD)
print("Uploaded " + uploaded_artifact_uuid)


service_config = struct(
    container_image_name = DATASTORE_IMAGE,
    used_ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    },
	files_artifact_mount_dirpaths = {
		uploaded_artifact_uuid: PATH_TO_MOUNT_UPLOADED_DIR
	}
)

add_service(service_id = DATASTORE_SERVICE_ID, service_config = service_config)`
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", startosisScript)

	executionResult, err := enclaveCtx.ExecuteStartosisScript(startosisScript)
	require.NoError(t, err, "Unexpected error executing startosis script")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Uploaded {{kurtosis:FILENAME_NOT_USED-13:38.artifact_uuid}}
`

	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the upload_file command to run")
	require.Lenf(t, executionResult.ValidationErrors, 0, "Unexpected validation error")
	require.Empty(t, executionResult.ExecutionError, "Unexpected execution error")
	require.Equal(t, expectedScriptOutput, executionResult.SerializedScriptOutput)
	logrus.Infof("Successfully ran Startosis script")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceId, portId),
		"Error validating datastore server '%s' is healthy",
		serviceId,
	)
	// Check that the file got uploaded on the service
	logrus.Infof("Checking that the file got uploaded on " + serviceId)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, _, err := serviceCtx.ExecCommand([]string{"ls", pathToCheckForUploadedFile})
	require.Nil(t, err, "Unexpected err running verification on upload file on "+serviceId)
	require.Equal(t, int32(0), exitCode)
}
