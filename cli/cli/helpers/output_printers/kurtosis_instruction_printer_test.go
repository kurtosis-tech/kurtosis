package output_printers

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	dryRun      = true
	executedRun = false
	isSkipped   = true
)

func testInstruction() *kurtosis_core_rpc_api_bindings.StarlarkInstruction {
	position := binding_constructors.NewStarlarkInstructionPosition("dummyFile", 12, 4)
	return binding_constructors.NewStarlarkInstruction(
		position,
		"my_instruction",
		`my_instruction("foo", ["bar", "doo"], kwarg1="serviceA", kwarg2=struct(bonjour=42, hello="world"))`,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{
			binding_constructors.NewStarlarkInstructionArg("foo", true),
			binding_constructors.NewStarlarkInstructionArg(`["bar", "doo"]`, false),
			binding_constructors.NewStarlarkInstructionKwarg("serviceA", "kwarg1", true),
			binding_constructors.NewStarlarkInstructionKwarg(`struct(bonjour=42, hello="world")`, "kwarg2", false),
		},
		isSkipped,
		"description")
}

func TestFormatInstruction_Executable(t *testing.T) {
	instruction := testInstruction()
	formattedInstruction := formatInstruction(instruction, run.Executable)
	expectedResult := `# from dummyFile[12:4]
my_instruction(
    "foo",
    [
        "bar",
        "doo",
    ],
    kwarg1 = "serviceA",
    kwarg2 = struct(
        bonjour = 42,
        hello = "world",
    ),
)`
	require.Equal(t, expectedResult, formattedInstruction)
}

func TestFormatInstruction_Detailed(t *testing.T) {
	instruction := testInstruction()
	formattedInstruction := formatInstruction(instruction, run.Detailed)
	expectedResult := `> my_instruction
> 	foo
> 	["bar", "doo"]
> 	kwarg1=serviceA
> 	kwarg2=struct(bonjour=42, hello="world")`
	require.Equal(t, expectedResult, formattedInstruction)
}

func TestFormatInstruction_Brief(t *testing.T) {
	instruction := testInstruction()
	formattedInstruction := formatInstruction(instruction, run.Brief)
	expectedResult := `> my_instruction foo kwarg1=serviceA`
	require.Equal(t, expectedResult, formattedInstruction)
}

func TestFormatInstruction_FormattingFail(t *testing.T) {
	instruction := binding_constructors.NewStarlarkInstruction(
		binding_constructors.NewStarlarkInstructionPosition("dummyFile", 12, 4),
		"print",
		// This has issues with the quotes not being escaped
		`print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`,
		[]*kurtosis_core_rpc_api_bindings.StarlarkInstructionArg{},
		isSkipped,
		"description")
	formattedInstruction := formatInstruction(instruction, run.Executable)
	// failure to format -> the instruction is returned with no formatting applied
	expectedResult := `# from dummyFile[12:4]
print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`
	require.Equal(t, expectedResult, formattedInstruction)
}

func TestFormatError(t *testing.T) {
	errorMessage := "There was an error"
	formattedErrorMessage := FormatError(errorMessage)
	require.Equal(t, errorMessage, formattedErrorMessage)
}

func TestFormatResult(t *testing.T) {
	resultMsg := "Hello world"
	instructionResult := binding_constructors.NewStarlarkRunResponseLineFromInstructionResult(resultMsg).GetInstructionResult()
	formattedResultMessage := formatInstructionResult(instructionResult)
	require.Equal(t, resultMsg, formattedResultMessage)
}

func TestFormatProgressBar(t *testing.T) {
	progressBar := formatProgressBar(4, 10, "=")
	expectedResult := fmt.Sprintf("%s%s", colorizeProgressBarIsDone("========"), colorizeProgressBarRemaining("============"))
	require.Equal(t, expectedResult, progressBar)
}

func TestFormatProgressBar_FloatingPointDivision(t *testing.T) {
	progressBar := formatProgressBar(10, 17, "=")
	expectedResult := fmt.Sprintf("%s%s", colorizeProgressBarIsDone("==========="), colorizeProgressBarRemaining("========="))
	require.Equal(t, expectedResult, progressBar)
}

func TestFormatProgressBar_NoTotalSteps(t *testing.T) {
	progressBar := formatProgressBar(0, 0, "=")
	expectedResult := colorizeProgressBarRemaining("====================")
	require.Equal(t, expectedResult, progressBar)
}

func TestFormatRunOutput_Successful_NoOutput_DryRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(``).GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, dryRun)
	expectedMessage := `Starlark code successfully run in dry-run mode. No output was returned.`
	require.Equal(t, expectedMessage, message)
}

func TestFormatRunOutput_Successful_WithOutput_DryRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(`{"hello": "world"}`).GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, dryRun)
	expectedMessage := `Starlark code successfully run in dry-run mode. Output was:
{"hello": "world"}`
	require.Equal(t, expectedMessage, message)
}

func TestFormatRunOutput_Successful_NoOutput_ExecutedRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(``).GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, executedRun)
	expectedMessage := `Starlark code successfully run. No output was returned.`
	require.Equal(t, expectedMessage, message)
}

func TestFormatRunOutput_Successful_WithOutput_ExecutedRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunSuccessEvent(`{"hello": "world"}`).GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, executedRun)
	expectedMessage := `Starlark code successfully run. Output was:
{"hello": "world"}`
	require.Equal(t, expectedMessage, message)
}

func TestFormatRunOutput_Failure_DryRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent().GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, dryRun)
	expectedMessage := `Error encountered running Starlark code in dry-run mode.`
	require.Equal(t, expectedMessage, message)
}

func TestFormatRunOutput_Failure_ExecutedRun(t *testing.T) {
	runFinishedEvent := binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent().GetRunFinishedEvent()
	message := formatRunOutput(runFinishedEvent, executedRun)
	expectedMessage := `Error encountered running Starlark code.`
	require.Equal(t, expectedMessage, message)
}
