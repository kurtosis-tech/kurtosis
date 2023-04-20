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
	defaultParallelism    = 4
	greetingsArg          = `{"greeting": "World!"}`

	serviceName                   = "example-datastore-server-1"
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
DATASTORE_SERVICE_NAME = "` + serviceName + `"
DATASTORE_PORT_ID = "` + portId + `"
DATASTORE_PORT_NUMBER = 1323
DATASTORE_PORT_PROTOCOL = "TCP"
FILE_TO_BE_CREATED = "` + fileToBeCreated + `"

SERVICE_DEPENDENT_ON_DATASTORE_SERVICE = "` + serviceIdForDependentService + `"
PATH_TO_MOUNT_ON_DEPENDENT_SERVICE =  "` + pathToCheckOnDependentService + `"

TEMPLATE_FILE_TO_RENDER="github.com/kurtosis-tech/eth2-package/static_files/prometheus-config/prometheus.yml.tmpl"
PATH_TO_MOUNT_RENDERED_CONFIG="` + renderedConfigMountPath + `"
RENDER_RELATIVE_PATH = "` + renderedConfigRelativePath + `"

def run(plan, args):
	plan.print("Hello " + args["greeting"]) 
	plan.print("Adding service " + DATASTORE_SERVICE_NAME + ".")
	
	config = ServiceConfig(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		}
	)
	
	result = plan.add_service(name = DATASTORE_SERVICE_NAME, config = config)
	plan.print("Service " + result.name + " deployed successfully.")
	plan.exec(
		recipe = ExecRecipe(
			command = ["touch", FILE_TO_BE_CREATED],
		),
		service_name = DATASTORE_SERVICE_NAME,
	)
	
	artifact_name = plan.store_service_files(name = "stored-file", service_name = DATASTORE_SERVICE_NAME, src = FILE_TO_BE_CREATED)
	plan.print("Stored file at " + artifact_name)
	
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
	
	rendered_artifact = plan.render_templates(template_data_by_path, "rendered-file")
	plan.print("Rendered file to " + rendered_artifact)
	
	dependent_config = ServiceConfig(
		image = DATASTORE_IMAGE,
		ports = {
			DATASTORE_PORT_ID: PortSpec(number = DATASTORE_PORT_NUMBER, transport_protocol = DATASTORE_PORT_PROTOCOL)
		},
		files = {
			PATH_TO_MOUNT_ON_DEPENDENT_SERVICE: artifact_name,
			PATH_TO_MOUNT_RENDERED_CONFIG: rendered_artifact
		}
	)
	deployed_service = plan.add_service(name = SERVICE_DEPENDENT_ON_DATASTORE_SERVICE, config = dependent_config)
	plan.print("Deployed " + SERVICE_DEPENDENT_ON_DATASTORE_SERVICE + " successfully")
	return {"ip-address": deployed_service.ip_address}
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

	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, startosisScript, greetingsArg, defaultDryRun, defaultParallelism)
	require.NoError(t, err, "Unexpected error executing startosis script")

	require.Nil(t, runResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, runResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, runResult.ExecutionError, "Unexpected execution error")

	expectedScriptOutput := `Hello World!
Adding service example-datastore-server-1.
Service 'example-datastore-server-1' added with service UUID '[a-z-0-9]+'
Service example-datastore-server-1 deployed successfully.
Command returned with exit code '0' with no output
Files with artifact name 'stored-file' uploaded with artifact UUID '[a-f0-9]{32}'
Stored file at stored-file
Templates artifact name 'rendered-file' rendered with artifact UUID '[a-f0-9]{32}'
Rendered file to rendered-file
Service 'example-datastore-server-2' added with service UUID '[a-z-0-9]+'
Deployed example-datastore-server-2 successfully
{
	"ip-address": "[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}"
}
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
	require.NoError(
		t,
		test_helpers.ValidateDatastoreServiceHealthy(context.Background(), enclaveCtx, serviceIdForDependentService, portId),
		"Error validating datastore server '%s' is healthy",
		serviceIdForDependentService,
	)
	logrus.Infof("All services added via the module work as expected")

	// Check that the file got created on the first service
	logrus.Infof("Checking that the file got created on " + serviceName)
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceName)
	require.Nil(t, err, "Unexpected Error Creating Service Context")
	exitCode, _, err := serviceCtx.ExecCommand([]string{"ls", fileToBeCreated})
	require.Nil(t, err, "Unexpected err running verification on created file on "+serviceName)
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
