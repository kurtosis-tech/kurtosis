package shared_starlark_calls

import (
	"context"
	"fmt"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/enclaves"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/services"
	"github.com/dzobbe/PoTE-kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	stopServiceStarlarkScript = `
def run(plan, args):
	plan.stop_service(name=args["service_name"])
`
)

func StopServiceStarlarkCommand(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName) error {
	params := fmt.Sprintf(`{"service_name": "%s"}`, serviceName)
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, stopServiceStarlarkScript, starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(params)))
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error occurred on Starlark for stopping service")
	}
	if runResult.ExecutionError != nil {
		return stacktrace.NewError("An error occurred during Starlark script execution for stopping service: %s", runResult.ExecutionError.GetErrorMessage())
	}
	if runResult.InterpretationError != nil {
		return stacktrace.NewError("An error occurred during Starlark script interpretation for stopping service: %s", runResult.InterpretationError.GetErrorMessage())
	}
	if len(runResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error occurred during Starlark script validation for stopping service: %v", runResult.ValidationErrors)
	}
	return nil
}
