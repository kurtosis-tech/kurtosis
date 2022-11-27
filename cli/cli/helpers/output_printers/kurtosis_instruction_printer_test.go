package output_printers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormatInstruction(t *testing.T) {
	instruction := binding_constructors.NewKurtosisInstruction(
		binding_constructors.NewKurtosisInstructionPosition("dummyFile", 12, 4),
		"my_instruction",
		`my_instruction("foo", ["bar", "doo"], kwarg1="serviceA", kwarg2=struct(bonjour=42, hello="world"))`,
		// TODO(gb): for now result is appended manually in the exec command code. This is change once we start doing streaming where the result is displayed right after the instruction code
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{
			binding_constructors.NewKurtosisInstructionArg("foo", true),
			binding_constructors.NewKurtosisInstructionArg(`["bar", "doo"]`, false),
			binding_constructors.NewKurtosisInstructionKwarg("serviceA", "kwarg1", true),
			binding_constructors.NewKurtosisInstructionKwarg(`struct(bonjour=42, hello="world")`, "kwarg2", false),
		})
	formattedInstruction := formatInstruction(instruction)
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

func TestFormatInstruction_FormattingFail(t *testing.T) {
	instruction := binding_constructors.NewKurtosisInstruction(
		binding_constructors.NewKurtosisInstructionPosition("dummyFile", 12, 4),
		"print",
		// This has issues with the quotes not being escaped
		`print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`,
		[]*kurtosis_core_rpc_api_bindings.KurtosisInstructionArg{})
	formattedInstruction := formatInstruction(instruction)
	// failure to format -> the instruction is returned with no formatting applied
	expectedResult := `# from dummyFile[12:4]
print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`
	require.Equal(t, expectedResult, formattedInstruction)
}
