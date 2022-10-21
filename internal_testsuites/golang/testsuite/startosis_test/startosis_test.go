package startosis_test

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

	serviceId                     = "example-datastore-server-1"
	serviceIdForDependentService  = "example-datastore-server-2"
	portId                        = "grpc"
	fileToBeCreated               = "/tmp/foo"
	mountPathOnDependentService   = "/tmp/doo"
	pathToCheckOnDependentService = mountPathOnDependentService + "/foo"
	fileToRead                    = "github.com/kurtosis-tech/sample-startosis-load/sample.star"

	startosisScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"
FILE_TO_BE_CREATED = "` + fileToBeCreated + `"
FILE_TO_READ = "` + fileToRead + `"

SERVICE_DEPENDENT_ON_DATASTORE_SERVICE = "` + serviceIdForDependentService + `"
PATH_TO_MOUNT_ON_DEPENDENT_SERVICE =  "` + pathToCheckOnDependentService + `"

print("Adding service " + DATASTORE_SERVICE_ID + ".")

service_config = struct(
    container_image_name = DATASTORE_IMAGE,
    used_ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    }
)

add_service(service_id = DATASTORE_SERVICE_ID, service_config = service_config)
print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")
exec(service_id = DATASTORE_SERVICE_ID, command = ["touch", FILE_TO_BE_CREATED])

artifact_uuid = store_file_from_service(service_id = DATASTORE_SERVICE_ID, src_path = FILE_TO_BE_CREATED)

dependent_service_config = struct(
    container_image_name = DATASTORE_IMAGE,
    used_ports = {
        DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
    },
	files_artifact_mount_dirpaths = {
		artifact_uuid : PATH_TO_MOUNT_ON_DEPENDENT_SERVICE
	}
)
add_service(service_id = SERVICE_DEPENDENT_ON_DATASTORE_SERVICE, service_config = dependent_service_config)

file_contents = read_file(FILE_TO_READ)
print("file_contents = " + file_contents)
`
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
Service example-datastore-server-1 deployed successfully.
file_contents = a = "World!"

`
	require.Empty(t, executionResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
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
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceIdForDependentService, portId),
		"Error validating datastore server '%s' is healthy",
		serviceIdForDependentService,
	)
	logrus.Infof("All services added via the module work as expected")

	// Check that the file got created on the first service
	logrus.Infof("Checking that the file got created on " + serviceId)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, _, err := serviceCtx.ExecCommand([]string{"ls", fileToBeCreated})
	require.Nil(t, err, "Unexpected err running verification on created file on "+serviceId)
	require.Equal(t, int32(0), exitCode)

	// Check that the file got mounted on the second service
	logrus.Infof("Checking that the file got mounted on " + serviceIdForDependentService)
	serviceCtx, err = enclaveCtx.GetServiceContext(serviceIdForDependentService)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, _, err = serviceCtx.ExecCommand([]string{"ls", pathToCheckOnDependentService})
	require.Nil(t, err, "Unexpected err running verification on mounted file on "+serviceIdForDependentService)
	require.Equal(t, int32(0), exitCode)
}
