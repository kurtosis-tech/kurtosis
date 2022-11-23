package output_printers

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFormatInstruction(t *testing.T) {
	instruction := binding_constructors.NewKurtosisInstruction(
		binding_constructors.NewKurtosisInstructionPosition("dummyFile", 12, 4),
		`my_instruction("foo", ["bar", "doo"], kwarg1="serviceA", kwarg2=struct(bonjour=42, hello="world"))`,
		// TODO(gb): for now result is appended manually in the exec command code. This is change once we start doing streaming where the result is displayed right after the instruction code
		nil)
	formattedInstruction := FormatInstruction(instruction)
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
		// This has issues with the quotes not being escaped
		`print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`,
		nil)
	formattedInstruction := FormatInstruction(instruction)
	// failure to format -> the instruction is returned with no formatting applied
	expectedResult := `# from dummyFile[12:4]
print("UNSUPPORTED_TYPE['ModuleOutput(grafana_info=GrafanaInfo(dashboard_path="/d/QdTOwy-nz/eth2-merge-kurtosis-module-dashboard?orgId=1", user="admin", password="admin"))']")`
	require.Equal(t, expectedResult, formattedInstruction)
}
