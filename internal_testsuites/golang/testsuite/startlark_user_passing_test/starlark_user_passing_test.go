package startlark_user_passing_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName                = "starlark-user-passing-test"
	noOverrideServiceName   = "no-override"
	userOverrideServiceName = "user-override"

	// We deliberately use a small image that (a) defaults to a non-root user so the
	// "no override" case can be distinguished from the root override, and (b) is already
	// pulled by the rest of the suite so it adds no node disk pressure. besu:24.3 was a
	// huge image whose only role here was "has a non-root default user", and pulling it
	// on a loaded kind node risked DiskPressure evictions; combined with
	// RestartPolicy=Never that killed the container and made exec fail with a confusing
	// "container not found". mendhak/http-https-echo defaults to the non-root 'node' user.
	starlarkScriptWithUserIdPassed = `
IMAGE = "mendhak/http-https-echo:26"
def run(plan, args):
	no_override = plan.add_service(
		name = "` + noOverrideServiceName + `",
		config = ServiceConfig(
			image = IMAGE,
			cmd = ["tail", "-f", "/dev/null"],
		)
	)

	plan.exec(service_name = no_override.name, recipe = ExecRecipe(command = ["whoami"]))

	root_override = plan.add_service(
		name = "` + userOverrideServiceName + `",
		config = ServiceConfig(
			image = IMAGE,
			cmd = ["tail", "-f", "/dev/null"],
			user = User(uid=0, gid=0),
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
	require.Nil(t, scriptRunResult.InterpretationError, "Unexpected interpretation error")
	require.Empty(t, scriptRunResult.ValidationErrors, "Unexpected validation error")
	require.Nil(t, scriptRunResult.ExecutionError, "Unexpected execution error")
	expectedOutput := `Service '` + noOverrideServiceName + `' added with service UUID '[a-z0-9]{32}'
Command returned with exit code '0' and the following output:
--------------------
node

--------------------
Service '` + userOverrideServiceName + `' added with service UUID '[a-z0-9]{32}'
Command returned with exit code '0' and the following output:
--------------------
root

--------------------
`

	require.Regexp(t, expectedOutput, scriptRunResult.RunOutput)

}
