package startosis_run_sh_task_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	runshTest           = "run-sh-test"
	runshStarlarkSimple = `
def run(plan):
  result1 = plan.run_sh(run="echo $EXAMPLE | tr -d '\n'", env_vars={"EXAMPLE": "kurtosis"})
  result2 = plan.run_sh(run="mkdir -p /src/{0} && cd /src/{0} && echo $(pwd)".format(result1.output))
  plan.verify(result2.output, "==", "/src/kurtosis\n")
`
	runshStarlarkFileArtifact = `
def run(plan):
  result = plan.run_sh(run="mkdir -p /src && echo kurtosis > /src/tech.txt && echo example > /src/example.txt", store=["/src/tech.txt", StoreSpec(src="/src", name="src"), StoreSpec(src="/src/*")], image="ethpandaops/ethereum-genesis-generator:1.0.14")
  file_artifacts = result.files_artifacts
  result2 = plan.run_sh(run="cat /temp/tech.txt", files={"/temp": file_artifacts[0]})
  plan.verify(result2.output, "==", "kurtosis\n")
  result3 = plan.run_sh(run="cat /task/tech.txt", files={"/task": file_artifacts[1]})
  plan.verify(result3.output, "==", "kurtosis\n")
  result4 = plan.run_sh(run = "cat /task/example.txt", files={"/task": file_artifacts[2]})
  plan.verify(result4.output, "==", "example\n")
`

	runshStarlarkWithTimeout = `
def run(plan):
  result = plan.run_sh(run="sleep 10s", wait="5s")
  plan.verify(value=result.code, assertion="==", target_value="0")
`
)

func TestStarlark_RunshTaskSimple(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlarkSimple)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output: kurtosis\nCommand returned with exit code '0' and the following output:\n--------------------\n/src/kurtosis\n\n--------------------\nVerification succeeded. Value is '\"/src/kurtosis\\n\"'.\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunshTaskFileArtifact(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlarkFileArtifact)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' with no output\nCommand returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nVerification succeeded. Value is '\"kurtosis\\n\"'.\nCommand returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nVerification succeeded. Value is '\"kurtosis\\n\"'.\nCommand returned with exit code '0' and the following output:\n--------------------\nexample\n\n--------------------\nVerification succeeded. Value is '\"example\\n\"'.\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunshTimesoutSuccess(t *testing.T) {
	ctx := context.Background()
	runResult, _ := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlarkWithTimeout)
	expectedErrorMessage := "The exec request timed out after 5 seconds"
	require.NotNil(t, runResult.ExecutionError)
	require.Contains(t, runResult.ExecutionError.GetErrorMessage(), expectedErrorMessage)
}
