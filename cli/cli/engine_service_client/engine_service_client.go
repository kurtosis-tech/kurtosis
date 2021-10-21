package engine_service_client

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-api-lib/golang/kurtosis_engine_rpc_api_consts"
	"github.com/palantir/stacktrace"
	"google.golang.org/grpc"
)

const (
	localHostIPAddressStr = "0.0.0.0"
)

func NewEngineServiceClient() (kurtosis_engine_rpc_api_bindings.EngineServiceClient, func() error, error) {
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