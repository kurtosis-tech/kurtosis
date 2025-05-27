package benchmark_test

import (
	"context"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/util"
	command_args_run "github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/run"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	testName = "benchmark-test"
)

func TestStartosis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, err := util.GetEnclave(ctx, testName)
	require.NoError(t, err)
	logrus.Infof("Retrieved enclave %v", testName)

	// enclaveCtx, _, _, err := test_helpers.CreateEnclave(t, ctx, testName)
	// require.NoError(t, err)
	// logrus.Infof("Created enclave %v", testName)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Ethereum package benchmark...")

	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams("{}"))

	starlarkResponseLineChan, cancelCtxFunc, err := enclaveCtx.RunStarlarkRemotePackage(ctx, "github.com/ethpandaops/ethereum-package", starlarkRunConfig)
	require.NoError(t, err)

	err = run.ReadAndPrintResponseLinesUntilClosed(starlarkResponseLineChan, cancelCtxFunc, command_args_run.Verbosity(0), false)
	require.NoError(t, err)
}
