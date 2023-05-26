package render_templates_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	enclaveTestName       = "render-templates-test"
	isPartitioningEnabled = false

	starlarkScript = `
def run(plan):
	template_data = {
		"Name" : "Stranger",
		"Answer": 6,
		"Numbers": [1, 2, 3],
		"UnixTimeStamp": 1257894000,
		"LargeFloat": 1231231243.43,
		"Alive": True
	}
	template = "Hello {{.Name}}. The sum of {{.Numbers}} is {{.Answer}}. My favorite moment in history {{.UnixTimeStamp}}. My favorite number {{.LargeFloat}}. Am I Alive? {{.Alive}}"
	expected_contents  = "Hello Stranger. The sum of [1 2 3] is 6. My favorite moment in history 1257894000. My favorite number 1231231243.43. Am I Alive? true"
	template_dict = {
		"grafana/config.yml": struct(
			template=template,
			data=template_data,
		),
		"config.yml": struct(
			template=template,
			data=template_data,
		)
	}
	
	artifact_name = plan.render_templates(config = template_dict, name="rendered-artifact")
	
	service = plan.add_service(
		name = "file-server",
		config = ServiceConfig(
			image = "flashspys/nginx-static",
			ports = {
				"http": PortSpec(
					number = 80,
					transport_protocol = "TCP",
				)
			},
			files = {
				"/static": artifact_name,
			},
		)
	)
	for filePath in template_dict:
		get_recipe = GetHttpRequestRecipe(
			port_id = "http",
			endpoint = "/" + filePath,
		)
		response = plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, service_name="file-server")
		plan.assert(response["body"], "==", expected_contents)
`
	noStarlarkParams = "{}"
	doNotDryRun      = false
)

func TestRenderTemplates(t *testing.T) {
	ctx := context.Background()
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, enclaveTestName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// -------------------------------------- SCRIPT RUN -----------------------------------------------
	runResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, starlarkScript)
	require.NoError(t, err, "An unexpected error occurred while running Starlark script")
	require.Empty(t, runResult.InterpretationError, "An unexpected error occurred while interpreting Starlark script")
	require.Empty(t, runResult.ValidationErrors, "An unexpected error occurred while validating Starlark script")
	require.Empty(t, runResult.ExecutionError, "An unexpected error occurred while executing Starlark script")
}
