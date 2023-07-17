package run_python_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	runPythonTest         = "run-python-test"
	runPythonSimpleScript = `
def run(plan):
	python_script = """
print("kurtosis")
	"""

	plan.run_python(
		run = python_script,
	)
`
)

func TestStarlark_RunPython(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonTest, runPythonSimpleScript)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}
