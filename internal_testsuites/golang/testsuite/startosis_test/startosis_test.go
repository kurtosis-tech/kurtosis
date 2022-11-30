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
	defaultDryRun         = false
	emptyInputArgs        = "{}"

	serviceId                     = "example-datastore-server-1"
	serviceIdForDependentService  = "example-datastore-server-2"
	portId                        = "grpc"
	fileToBeCreated               = "/tmp/foo"
	mountPathOnDependentService   = "/tmp/doo"
	pathToCheckOnDependentService = mountPathOnDependentService + "/foo"
	renderedConfigMountPath       = "/config"
	renderedConfigRelativePath    = "foo/bar.yml"
	renderedConfigFile            = renderedConfigMountPath + "/" + renderedConfigRelativePath

	startosisScript = `
DATASTORE_IMAGE = "kurtosistech/example-datastore-server"
DATASTORE_SERVICE_ID = "` + serviceId + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"
FILE_TO_BE_CREATED = "` + fileToBeCreated + `"

SERVICE_DEPENDENT_ON_DATASTORE_SERVICE = "` + serviceIdForDependentService + `"
PATH_TO_MOUNT_ON_DEPENDENT_SERVICE =  "` + pathToCheckOnDependentService + `"

TEMPLATE_FILE_TO_RENDER="github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/static_files/prometheus-config/prometheus.yml.tmpl"
PATH_TO_MOUNT_RENDERED_CONFIG="` + renderedConfigMountPath + `"
RENDER_RELATIVE_PATH = "` + renderedConfigRelativePath + `"

def run(args):
	print("Adding service " + DATASTORE_SERVICE_ID + ".")
	
	config = struct(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	add_service(service_id = DATASTORE_SERVICE_ID, config = config)
	print("Service " + DATASTORE_SERVICE_ID + " deployed successfully.")
	exec(service_id = DATASTORE_SERVICE_ID, command = ["touch", FILE_TO_BE_CREATED])
	
	artifact_id = store_service_files(service_id = DATASTORE_SERVICE_ID, src = FILE_TO_BE_CREATED)
	print("Stored file at " + artifact_id)
	
	template_str = read_file(TEMPLATE_FILE_TO_RENDER)
	
	template_data = {
		"CLNodesMetricsInfo" : [{"name" : "foo", "path": "/foo/path", "url": "foobar.com"}]
	}
	
	template_data_by_path = {
		RENDER_RELATIVE_PATH : struct(
			template= template_str,
			data= template_data
		)
	}
	
	rendered_artifact = render_templates(template_data_by_path)
	print("Rendered file to " + rendered_artifact)
	
	dependent_config = struct(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: struct(number = DATASTORE_PORT_NUMBER, protocol = DATASTORE_PORT_PROTOCOL)
		},
		files = {
			artifact_id : PATH_TO_MOUNT_ON_DEPENDENT_SERVICE,
			rendered_artifact : PATH_TO_MOUNT_RENDERED_CONFIG
		}
	)
	add_service(service_id = SERVICE_DEPENDENT_ON_DATASTORE_SERVICE, config = dependent_config)
	print("Deployed " + SERVICE_DEPENDENT_ON_DATASTORE_SERVICE + " successfully")
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

	outputStream, _, err := enclaveCtx.RunStarlarkScript(ctx, startosisScript, emptyInputArgs, defaultDryRun)
	require.NoError(t, err, "Unexpected error executing startosis script")
	scriptOutput, _, interpretationError, validationErrors, executionError := test_helpers.ReadStreamContentUntilClosed(outputStream)

	expectedScriptOutput := `Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with internal ID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
Command returned with exit code '0' with no output
Files stored with artifact ID '[a-f0-9-]{36}'
Stored file at [a-f0-9-]{36}
Templates rendered and stored with artifact ID '[a-f0-9-]{36}'
Rendered file to [a-f0-9-]{36}
Service 'example-datastore-server-2' added with internal ID '[a-z-0-9]+'
Deployed example-datastore-server-2 successfully
`
	require.Nil(t, interpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, validationErrors, "Unexpected validation error")
	require.Nil(t, executionError, "Unexpected execution error")
	require.Regexp(t, expectedScriptOutput, scriptOutput)
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

	// Check that the file got rendered on the second service
	expectedConfigFile := `global:
  scrape_interval:     15s # By default, scrape targets every 15 seconds.

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
   
   - job_name: 'foo'
     metrics_path: /foo/path
     static_configs:
       - targets: ['foobar.com']
   `
	logrus.Infof("Checking that the file got mounted on " + serviceIdForDependentService)
	serviceCtx, err = enclaveCtx.GetServiceContext(serviceIdForDependentService)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, configFileContent, err := serviceCtx.ExecCommand([]string{"cat", renderedConfigFile})
	require.Nil(t, err, "Unexpected err running verification on rendered file on "+serviceIdForDependentService)
	require.Equal(t, int32(0), exitCode)
	require.Equal(t, expectedConfigFile, configFileContent, "Rendered file contents don't match expected contents")
}
