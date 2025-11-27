package run_python_test

import (
	"context"
	"github.com/dzobbe/PoTE-kurtosis-cli/golang_internal_testsuite/test_helpers"
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

	runPythonWithPackagesArgsTest      = "run-python-package-args-test"
	runPythonWithPackagesScriptAndArgs = `
def run(plan):
	python_script = """
import requests
import sys
response = requests.get("https://docs.kurtosis.com")
print(response.status_code)
print(sys.argv[1])
	"""

	plan.run_python(
		run = python_script,
		packages = ["requests"],
		args = ["Kurtosis"]
	)
`
	runPythonWithSingleQuotesTest   = "run-python-single-quote"
	runPythonWithSingleQuotesScript = `
def run(plan):
	python_script = """
print('kurtosis')
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
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunPythonWithExternalPacakges(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonWithPackagesArgsTest, runPythonWithPackagesScriptAndArgs)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\n200\nKurtosis\n\n--------------------\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}

func TestStarlark_RunPythonWithSingleQuotes(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonWithSingleQuotesTest, runPythonWithSingleQuotesScript)
	require.Nil(t, err)
	expectedOutput := "Command returned with exit code '0' and the following output:\n--------------------\nkurtosis\n\n--------------------\n"
	require.Equal(t, expectedOutput, string(runResult.RunOutput))
}
