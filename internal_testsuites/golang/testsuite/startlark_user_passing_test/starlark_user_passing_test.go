package startlark_user_passing_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "starlark-user-passing-test"

	starlarkScriptWithUserIdPassed = `
IMAGE = "hyperledger/besu:latest"
def run(plan, args):
	no_override = plan.add_service(
		name = "besu-no-override",
		config = ServiceConfig(
			image = IMAGE,
			cmd = ["tail", "-f", "/dev/null"],
		)
	)

	plan.exec(service_name = no_override.name, recipe = ExecRecipe(command = ["whoami"]))

	root_override = plan.add_service(
		name = "besu-root-override",
		config = ServiceConfig(
			image = IMAGE,
			cmd = ["tail", "-f", "/dev/null"],
			user = User(uid=0),
		)
	)

	plan.exec(service_name = root_override.name, recipe = ExecRecipe(command = ["whoami"]))
`
)

func TestUserIDOverridesWork(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() {
		err = destroyEnclaveFunc()
		require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	}()

	// ------------------------------------- TEST RUN ----------------------------------------------
	scriptRunResult, err := test_helpers.RunScriptWithDefaultConfig(ctx, enclaveCtx, starlarkScriptWithUserIdPassed)
	logrus.Infof("Test Output: %v", scriptRunResult)
	require.NoError(t, err, "Unexpected error executing starlark script")
	require.Nil(t, scriptRunResult.InterpretationError, "Unexpected interpretation error. This test requires you to be online for the read_file command to run")
	require.Empty(t, scriptRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, scriptRunResult.ExecutionError, "Unexpected execution error")
	expectedOutput := `Service 'besu-no-override' added with service UUID '[a-z0-9]{32}'
Command returned with exit code '0' and the following output:
--------------------
besu

--------------------
Service 'besu-root-override' added with service UUID '[a-z0-9]{32}'
Command returned with exit code '0' and the following output:
--------------------
root

--------------------
`

	require.Regexp(t, expectedOutput, scriptRunResult.RunOutput)

}
