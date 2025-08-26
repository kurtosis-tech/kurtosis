package cloud

import (
	"crypto/tls"
	"crypto/x509"
	api "github.com/kurtosis-tech/kurtosis/cloud/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func CreateCloudClient(connectionStr string, caCertChain string) (api.KurtosisCloudBackendServerClient, error) {
	caCertChainBytes := []byte(caCertChain)
	p := x509.NewCertPool()
	p.AppendCertsFromPEM(caCertChainBytes)

	// nolint: exhaustruct
	tlsConfig := &tls.Config{
		RootCAs: p,
	}
	conn, err := grpc.Dial(connectionStr, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
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
