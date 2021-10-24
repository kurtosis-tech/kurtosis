package engine_client

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/start"
	"github.com/kurtosis-tech/kurtosis-cli/cli/commands/engine/stop"
	engine_status_retriever2 "github.com/kurtosis-tech/kurtosis-cli/cli/helpers/engine_status_retriever"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
	"os"
	"path"
)

const (
	localHostIPAddressStr = "0.0.0.0"
)

func NewEngineClientFromLocalEngine(ctx context.Context, dockerManager *docker_manager.DockerManager) (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
	// Check the engine status first so we can print a helpful message in case the engine isn't running
	status, _, err := engine_status_retriever2.RetrieveEngineStatus(ctx, dockerManager)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred retrieving the Kurtosis engine status, which is necessary for creating a connection to the engine")
	}
	binaryFilename := path.Base(os.Args[0])
	switch status {
	case engine_status_retriever2.EngineStatus_Stopped:
		return nil, nil, stacktrace.NewError(
			"No Kurtosis engine is running; you'll need to start one by running '%v %v %v'",
			binaryFilename,
			engine.CommandStr,
			start.CommandStr,
		)
	case engine_status_retriever2.EngineStatus_ContainerRunningButServerNotResponding:
		return nil, nil, stacktrace.NewError(
			"A Kurtosis engine container is running, but it's not responding; this shouldn't happen and you'll likely " +
				"want to restart the engine by running '%v %v %v && %v %v %v'",
			binaryFilename,
			engine.CommandStr,
			stop.CommandStr,
			binaryFilename,
			engine.CommandStr,
			start.CommandStr,
		)
	case engine_status_retriever2.EngineStatus_Running:
		// This is the happy case; nothing to do
	default:
		return nil, nil, stacktrace.NewError("Unrecognized engine status '%v'; this is a bug in Kurtosis", status)
	}

	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, kurtosis_engine_rpc_api_consts.ListenPort)

	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	conn, err := grpc.Dial(kurtosisEngineSocketStr, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Engine Server at '%v'",
			kurtosisEngineSocketStr,
		)
	}

	engineServiceClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)

	return engineServiceClient, conn.Close, nil
}
