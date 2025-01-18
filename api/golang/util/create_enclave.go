package util

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/v2/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
)

func CreateEmptyEnclave(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, func() error, func() error, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to get kurtosis context from local engine. Is the engine running? Try running 'kurtosis engine start'")
	}
	enclaveCtx, err := kurtosisCtx.CreateEnclave(ctx, enclaveName)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to create enclave.")
	}
	stopEnclaveFunc := func() error {
		return kurtosisCtx.StopEnclave(ctx, enclaveName)

	}
	destroyEnclaveFunc := func() error {
		return kurtosisCtx.DestroyEnclave(ctx, enclaveName)
	}

	return enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, nil
}

func combineErrors(starlarkResult *enclaves.StarlarkRunResult, err error) error {
	if err != nil {
		return stacktrace.Propagate(err, "Failed to run Starlark during setup")
	}
	if starlarkResult.InterpretationError != nil {
		return stacktrace.NewError("Failed to setup enclave using Starlark because of an interpretation error:\n%v", starlarkResult.InterpretationError)
	}
	if len(starlarkResult.ValidationErrors) > 0 {
		return stacktrace.NewError("Failed to setup enclave using Starlark because of an interpretation error:\n%v", starlarkResult.ValidationErrors)
	}
	if starlarkResult.ExecutionError != nil {
		return stacktrace.NewError("Failed to setup enclave using Starlark because of an execution error:\n%v", starlarkResult.ExecutionError)
	}
	return nil
}

func CreateEnclaveFromStarlarkScript(ctx context.Context, enclaveName string, serializedScript string, runConfig *starlark_run_config.StarlarkRunConfig) (*enclaves.EnclaveContext, func() error, func() error, error) {
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := CreateEmptyEnclave(ctx, enclaveName)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to create empty enclave")
	}
	starlarkResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, serializedScript, runConfig)
	err = combineErrors(starlarkResult, err)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to setup enclave using Starlark")
	}
	return enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, nil
}

func CreateEnclaveFromStarlarkRemotePackage(ctx context.Context, enclaveName string, packageId string, runConfig *starlark_run_config.StarlarkRunConfig) (*enclaves.EnclaveContext, func() error, func() error, error) {
	enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, err := CreateEmptyEnclave(ctx, enclaveName)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to create empty enclave")
	}
	starlarkResult, err := enclaveCtx.RunStarlarkRemotePackageBlocking(ctx, packageId, runConfig)
	err = combineErrors(starlarkResult, err)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "Failed to setup enclave using Starlark")
	}
	return enclaveCtx, stopEnclaveFunc, destroyEnclaveFunc, nil
}
