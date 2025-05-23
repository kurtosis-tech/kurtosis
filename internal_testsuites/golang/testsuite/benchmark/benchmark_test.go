package benchmark_test

import (
	"context"
	"time"
	"github.com/kurtosis-tech/kurtosis/benchmark"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/run"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "benchmark-test"
)

func TestStartosis(t *testing.T) {
	ctx := context.Background()

	benchmark := benchmark.RunBenchmark{
		TimeToRun: time.Duration(0),
		TimeToCreateEnclave: time.Duration(0),
		TimeToUploadStarlarkPackage: time.Duration(0),
		TimeToExecuteStarlark: time.Duration(0),
	}
	beforeRun := time.Now()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	beforeCreateEnclave := time.Now()
	enclaveCtx, _, _, err := test_helpers.CreateEnclave(t, ctx, testName)
	benchmark.TimeToCreateEnclave = time.Since(beforeCreateEnclave)

	require.NoError(t, err, "An error occurred creating an enclave")
	// defer func() {
	// 	err = destroyEnclaveFunc()
	// 	require.NoError(t, err, "An error occurred destroying the enclave after the test finished")
	// }()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Ethereum package benchmark...")

	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(""))

	beforeRunStarlarkPackage := time.Now()
	starlarkResponseLineChan, cancelCtxFunc, err := enclaveCtx.RunStarlarkRemotePackage(ctx, "github.com/ethpandaops/ethereum-package", starlarkRunConfig)
	require.NoError(t, err)

	err = run.ReadAndPrintResponseLinesUntilClosed(starlarkResponseLineChan, cancelCtxFunc, nil, false)
	require.NoError(t, err)

	benchmark.TimeToExecuteStarlark = time.Since(beforeRunStarlarkPackage)

	benchmark.TimeToRun = time.Since(beforeRun)
	benchmark.Print()
}
