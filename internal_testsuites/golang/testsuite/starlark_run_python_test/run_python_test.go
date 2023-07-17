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

	runPythonWithPackagesTest   = "run-python-package-test"
	runPythonWithPackagesScript = `
def run(plan):
	python_script = """
import requests
response = requests.get("https://docs.kurtosis.com")
print(response.status_code)
	"""

	plan.run_python(
		run = python_script,
		packages = ["requests"]
	)
`
)

func TestStarlark_RunPython(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonTest, runPythonSimpleScript)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunPythonWithExternalPacakges(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonWithPackagesTest, runPythonWithPackagesScript)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\n200\n\n--------------------\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}
