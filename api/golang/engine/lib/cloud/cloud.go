package cloud

import (
	"crypto/tls"
	"crypto/x509"
	api "github.com/kurtosis-tech/kurtosis-cloud-backend/api/golang/kurtosis_backend_server_rpc_api_bindings"
	"github.com/kurtosis-tech/stacktrace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

func CreateCloudClient(connectionStr string) (api.KurtosisCloudBackendServerClient, error) {
	caChain := "/var/tmp/fullchain.pem" // CA cert that signed the proxy
	f, err := os.ReadFile(caChain)

	p := x509.NewCertPool()
	p.AppendCertsFromPEM(f)

	tlsConfig := &tls.Config{RootCAs: p}
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
