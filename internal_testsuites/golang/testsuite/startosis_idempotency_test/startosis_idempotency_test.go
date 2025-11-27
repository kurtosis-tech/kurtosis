package startosis_persistent_directory_test

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "persist-data-test"

	templateScript = `
IMAGE = "docker/getting-started"
SERVICE_NAME = "test-service"

def run(plan):
    # create a file artifact from a template
    template_data = {"World" : "World!"}
    plan.render_templates(
        name="render_template_file_artifact",
        config={
            "/output.txt": struct(template="%s {{.World}}", data={"World" : "World!"}),
        }
    )

    # create a file artifact from a service
    file_generator_service_name = "file-generator" 
    service = plan.add_service(
		name=file_generator_service_name, 
		config=ServiceConfig(
			image=IMAGE,
			cmd=[
				"/bin/sh",
				"-c",
				"echo '%s world !' >> /test.log && sleep 99999"
			]
		)
	)
    plan.store_service_files(service_name=file_generator_service_name, src="/test.log", name="service_file_artifact")

    service = plan.add_service(
        name=SERVICE_NAME,
		config=ServiceConfig(
            image=IMAGE,
            files={
                "/data/": "render_template_file_artifact", 
            }
        )
    )
`
)

func TestIdempotentRenderTemplateTest(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	firstScript := fmt.Sprintf(templateScript, "Hello", "Hello")
	firstRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, firstScript)
	logrus.Infof("Test Output: %v", firstRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, firstRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, firstRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, firstRunResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, `Templates artifact name 'render_template_file_artifact' rendered with artifact UUID '[a-z0-9]{32}'
Service 'file-generator' added with service UUID '[a-z0-9]{32}'
Files with artifact name 'service_file_artifact' uploaded with artifact UUID '[a-z0-9]{32}'
Service 'test-service' added with service UUID '[a-z0-9]{32}'`, firstRunResult.RunOutput)

	secondScript := fmt.Sprintf(templateScript, "Bonjour", "Hello")
	secondRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, secondScript)
	logrus.Infof("Test Output: %v", secondRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, secondRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, secondRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, secondRunResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, `Templates artifact name 'render_template_file_artifact' rendered with artifact UUID '[a-z0-9]{32}'
SKIPPED - This instruction has already been run in this enclave
SKIPPED - This instruction has already been run in this enclave
Service 'test-service' added with service UUID '[a-z0-9]{32}'`, secondRunResult.RunOutput)
}

func TestIdempotentStoreServiceFileTest(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	firstScript := fmt.Sprintf(templateScript, "Hello", "Hello")
	firstRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, firstScript)
	logrus.Infof("Test Output: %v", firstRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, firstRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, firstRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, firstRunResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, `Templates artifact name 'render_template_file_artifact' rendered with artifact UUID '[a-z0-9]{32}'
Service 'file-generator' added with service UUID '[a-z0-9]{32}'
Files with artifact name 'service_file_artifact' uploaded with artifact UUID '[a-z0-9]{32}'
Service 'test-service' added with service UUID '[a-z0-9]{32}'`, firstRunResult.RunOutput)

	secondScript := fmt.Sprintf(templateScript, "Hello", "Bonjour")
	secondRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, secondScript)
	logrus.Infof("Test Output: %v", secondRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, secondRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, secondRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, secondRunResult.ExecutionError, "Unexpected execution error")
	require.Regexp(t, `SKIPPED - This instruction has already been run in this enclave
Service 'file-generator' added with service UUID '[a-z0-9]{32}'
Files with artifact name 'service_file_artifact' uploaded with artifact UUID '[a-z0-9]{32}'
SKIPPED - This instruction has already been run in this enclave
`, secondRunResult.RunOutput)
}
