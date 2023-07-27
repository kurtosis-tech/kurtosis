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

	runPythonWithGiganticOutputGreaterThan4Mb    = "run-large-python-output-test"
	runPythonWithLargeOutputGreaterThan4MbScript = `
def run(plan):
    plan.run_python(
        run = """
print("hello world"*2**20)
"""
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

func TestStarlark_RunPythonWithLargeOutputGreaterThan4Mb(t *testing.T) {
	ctx := context.Background()
	runResult, err := test_helpers.SetupSimpleEnclaveAndRunScript(t, ctx, runPythonWithGiganticOutputGreaterThan4Mb, runPythonWithLargeOutputGreaterThan4MbScript)
	require.Nil(t, err)
	require.Greater(t, len(runResult.RunOutput), 2^20)
}
