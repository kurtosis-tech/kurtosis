package startosis_upload_files_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "upload-files-test"
	isPartitioningEnabled = false
	defaultDryRun         = false

	serviceName = "example-datastore-server-1"
	portId      = "grpc"

	pathToMountUploadedDir     = "/uploads"
	pathToCheckForUploadedFile = "/uploads/helpers.star"

	emptyParams     = "{}"
	startosisScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_NAME = "` + serviceName + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"

DIR_TO_UPLOAD = "github.com/kurtosis-tech/datastore-army-package/src"
PATH_TO_MOUNT_UPLOADED_DIR = "` + pathToMountUploadedDir + `"

def run(plan):
	plan.print("Adding service " + DATASTORE_SERVICE_NAME + ".")
	
	artifact_name = plan.upload_files(DIR_TO_UPLOAD, name = "test-artifact")
	plan.print("Uploaded " + artifact_name)
	
	
	config = ServiceConfig(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		},
		files = {
			PATH_TO_MOUNT_UPLOADED_DIR: artifact_name
		}
	)
	
	plan.add_service(service_name = DATASTORE_SERVICE_NAME, config = config)`
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

	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, startosisScript)
	require.NoError(t, err, "Unexpected error executing startosis script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the upload_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `Adding service example-datastore-server-1.
Files with artifact name 'test-artifact' uploaded with artifact UUID '[a-f0-9]{32}'
Uploaded test-artifact
Service 'example-datastore-server-1' added with service UUID '[a-z-0-9]+'
`
	require.Regexp(t, expectedScriptOutput, string(runResult.RunOutput))
	logrus.Infof("Successfully ran Startosis script")

	// Check that the service added by the script is functional
	logrus.Infof("Checking that services are all healthy")
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceName, portId),
		"Error validating datastore server '%s' is healthy",
		serviceName,
	)
	// Check that the file got uploaded on the service
	logrus.Infof("Checking that the file got uploaded on " + serviceName)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceName)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, _, err := serviceCtx.ExecCommand([]string{"ls", pathToCheckForUploadedFile})
	require.Nil(t, err, "Unexpected err running verification on upload file on "+serviceName)
	require.Equal(t, int32(0), exitCode)
}
