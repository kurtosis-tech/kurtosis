package cloud

import (
	api "github.com/kurtosis-tech/kurtosis-cloud-backend/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CreateCloudClient(connectionStr string) (api.KurtosisCloudBackendServerClient, error) {
	conn, err := grpc.Dial(connectionStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Cloud server at '%v'",
			connectionStr,
		)
	}
	client := api.NewKurtosisCloudBackendServerClient(conn)
	return client, nil
}
