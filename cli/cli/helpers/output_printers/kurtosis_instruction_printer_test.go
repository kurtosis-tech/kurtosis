package output_printers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/stretchr/testify/require"
	"testing"
)

func testInstruction() *kurtosis_core_rpc_api_bindings.KurtosisInstruction {
	position := binding_constructors.NewKurtosisInstructionPosition("dummyFile", 12, 4)
	return binding_constructors.NewKurtosisInstruction(
		position,
		"my_instruction",
		`my_instruction("foo", ["bar", "doo"], kwarg1="serviceA", kwarg2=struct(bonjour=42, hello="world"))`,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
			binding_constructors.NewKurtosisInstructionArg("foo", true),
			binding_constructors.NewKurtosisInstructionArg(`["bar", "doo"]`, false),
			binding_constructors.NewKurtosisInstructionKwarg("serviceA", "kwarg1", true),
			binding_constructors.NewKurtosisInstructionKwarg(`struct(bonjour=42, hello="world")`, "kwarg2", false),
		})
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
	instruction := binding_constructors.NewKurtosisInstruction(
		binding_constructors.NewKurtosisInstructionPosition("dummyFile", 12, 4),
		"print",
		// This has issues with the quotes not being escaped
		`print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{})
	formattedInstruction := formatInstruction(instruction, run.Executable)
	// failure to format -> the instruction is returned with no formatting applied
	expectedResult := `# from dummyFile[12:4]
print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`
	require.Equal(t, expectedResult, formattedInstruction)
}
