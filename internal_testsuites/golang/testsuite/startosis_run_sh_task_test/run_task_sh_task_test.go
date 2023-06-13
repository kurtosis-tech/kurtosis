package startosis_run_sh_task_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	runshTest     = "run-sh-test"
	runshStarlark = `
def run(plan):
  result1 = plan.run_sh(run="echo kurtosis", workdir="src")
  result2 = plan.run_sh(run="mkdir -p {0} && cd {0} && echo $(pwd)".format(result1["output"]))
  plan.assert(result2["output"], "==", "/src/kurtosis\n")
`
)

func TestStarlark_RunshTask(t *testing.T) {
	ctx := context.Background()
	runResult, _ := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlark)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nCommand returned with exit code '0' and the following output:\n--------------------\n/src/kurtosis\n\n--------------------\nAssertion succeeded. Value is '\"/src/kurtosis\\n\"'.\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}
