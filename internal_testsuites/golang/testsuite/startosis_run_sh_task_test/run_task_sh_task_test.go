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
  result1 = plan.run_sh(run="echo kurtosis")
  result2 = plan.run_sh(run="mkdir -p {0} && cd {0} && echo $(pwd)".format(result1.output), workdir="src")
  plan.assert(result2.output, "==", "/src/kurtosis\n")
`
	runshStarlarkFileArtifact = `
def run(plan):
  result = plan.run_sh(run="mkdir -p src && echo kurtosis > /src/tech.txt", store=["/src/tech.txt", "/src"])
  file_artifacts = result.files_artifacts
  result2 = plan.run_sh(run="cat /src/temp/tech.txt", files={"temp": file_artifacts[0]}, workdir="/src")
  plan.assert(result2.output, "==", "kurtosis\n")
  result3 = plan.run_sh(run="cat ./src/tech.txt", files={"/task": file_artifacts[1]}, workdir="/task")
  plan.assert(result3.output, "==", "kurtosis\n")
`
)

func TestStarlark_RunshTaskSimple(t *testing.T) {
	ctx := context.Background()
	runResult, _ := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlarkSimple)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nCommand returned with exit code '0' and the following output:\n--------------------\n/src/kurtosis\n\n--------------------\nAssertion succeeded. Value is '\"/src/kurtosis\\n\"'.\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunshTaskFileArtifact(t *testing.T) {
	ctx := context.Background()
	runResult, _ := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runshTest, runshStarlarkFileArtifact)
	expectedOutput := "Command returned with exit code '0' with no output\nCommand returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nAssertion succeeded. Value is '\"kurtosis\\n\"'.\nCommand returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\nAssertion succeeded. Value is '\"kurtosis\\n\"'.\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}
