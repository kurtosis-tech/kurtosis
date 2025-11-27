package startosis_persistent_directory_test

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/enclaves"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	enclaveNameTemplate = "idempotent-run-test-%d"

	skippedInstructionMessage = "SKIPPED - This instruction has already been run in this enclave"
)

func TestStartosisIdempotentRun_RunSameScriptTwice(t *testing.T) {
	ctx := context.Background()
	enclaveName := fmt.Sprintf(enclaveNameTemplate, 1)
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveName)
	require.NoError(t, err)

	// Run 1
	scriptToRun := `def run(plan):
	plan.print("Script run twice")
`
	result := mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun1Output := `Script run twice
`
	require.Equal(t, expectedRun1Output, result)

	// Run 2
	result = mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun2Output := fmt.Sprintf(`%s
`, skippedInstructionMessage)
	require.Equal(t, expectedRun2Output, result)

	require.NoError(t, destroyEnclaveFunc())
}

func TestStartosisIdempotentRun_IterativelyBuildEnclave(t *testing.T) {
	ctx := context.Background()
	enclaveName := fmt.Sprintf(enclaveNameTemplate, 2)
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveName)
	require.NoError(t, err)

	// iteration 1
	scriptToRun := `def run(plan):
	plan.print("First instruction")
`
	result := mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun1Output := `First instruction
`
	require.Equal(t, expectedRun1Output, result)

	// iteration 2
	scriptToRun = `def run(plan):
	plan.print("First instruction")
	plan.print("Second instruction")
`
	result = mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun2Output := fmt.Sprintf(`%s
Second instruction
`, skippedInstructionMessage)
	require.Equal(t, expectedRun2Output, result)

	// iteration 3
	scriptToRun = `def run(plan):
	plan.print("First instruction")
	plan.print("Second instruction")
	plan.print("Third instruction")
`
	result = mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun3Output := fmt.Sprintf(`%s
%s
Third instruction
`, skippedInstructionMessage, skippedInstructionMessage)
	require.Equal(t, expectedRun3Output, result)

	require.NoError(t, destroyEnclaveFunc())
}

func TestStartosisIdempotentRun_RunTwoDifferentScripts(t *testing.T) {
	ctx := context.Background()
	enclaveName := fmt.Sprintf(enclaveNameTemplate, 3)
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveName)
	require.NoError(t, err)

	// Run 1
	scriptToRun := `def run(plan):
	plan.print("Running script 1")
`
	result := mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun1Output := `Running script 1
`
	require.Equal(t, expectedRun1Output, result)

	// Run 2
	scriptToRun = `def run(plan):
	plan.print("Running script 2")
`
	result = mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun2Output := `Running script 2
`
	require.Equal(t, expectedRun2Output, result)

	require.NoError(t, destroyEnclaveFunc())
}

func TestStartosisIdempotentRun_RunTwoOverlappingScripts(t *testing.T) {
	ctx := context.Background()
	enclaveName := fmt.Sprintf(enclaveNameTemplate, 4)
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveName)
	require.NoError(t, err)

	// Run 1
	scriptToRun := `def run(plan):
	plan.print("Script 1 specific instruction")
	plan.print("Overlapping instruction")
`
	result := mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun1Output := `Script 1 specific instruction
Overlapping instruction
`
	require.Equal(t, expectedRun1Output, result)

	// Run 2
	scriptToRun = `def run(plan):
	plan.print("Overlapping instruction")
	plan.print("Script 2 specific instruction")
`
	result = mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun2Output := fmt.Sprintf(`%s
Script 2 specific instruction
`, skippedInstructionMessage)
	require.Equal(t, expectedRun2Output, result)

	require.NoError(t, destroyEnclaveFunc())
}

func TestStartosisIdempotentRun_DeactivateInstructionsCaching(t *testing.T) {
	ctx := context.Background()
	enclaveName := fmt.Sprintf(enclaveNameTemplate, 5)
	enclaveCtx, _, destroyEnclaveFunc, err := test_helpers.CreateEnclave(t, ctx, enclaveName)
	require.NoError(t, err)

	// Run 1
	scriptToRun := `def run(plan):
	plan.print("Script run twice")
`
	result := mustRunStarlarkScript(t, enclaveCtx, ctx, scriptToRun)
	expectedRun1Output := `Script run twice
`
	require.Equal(t, expectedRun1Output, result)

	// Run 2
	deactivateInstructionsCaching := []kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag{
		kurtosis_core_rpc_api_bindings.KurtosisFeatureFlag_NO_INSTRUCTIONS_CACHING,
	}
	run2result, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, scriptToRun, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithExperimentalFeatureFlags(deactivateInstructionsCaching)))
	require.NoError(t, err)

	require.Nil(t, run2result.InterpretationError)
	require.Empty(t, run2result.ValidationErrors)
	require.Nil(t, run2result.ExecutionError)
	expectedRun2Output := `Script run twice
`
	require.Equal(t, expectedRun2Output, result)

	require.NoError(t, destroyEnclaveFunc())
}

func mustRunStarlarkScript(t *testing.T, enclaveCtx *enclaves.EnclaveContext, ctx context.Context, script string) string {
	result, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, script, starlark_run_config.NewRunStarlarkConfig())
	require.NoError(t, err)

	require.Nil(t, result.InterpretationError)
	require.Empty(t, result.ValidationErrors)
	require.Nil(t, result.ExecutionError)
	return string(result.RunOutput)
}
