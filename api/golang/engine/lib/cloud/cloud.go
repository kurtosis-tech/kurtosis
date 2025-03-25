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

	tlsConfig := &tls.Config{
		Rand:                                nil,
		Time:                                nil,
		Certificates:                        nil,
		NameToCertificate:                   nil,
		GetCertificate:                      nil,
		GetClientCertificate:                nil,
		GetConfigForClient:                  nil,
		VerifyPeerCertificate:               nil,
		VerifyConnection:                    nil,
		RootCAs:                             nil,
		NextProtos:                          nil,
		ServerName:                          "",
		ClientAuth:                          0,
		ClientCAs:                           nil,
		InsecureSkipVerify:                  false,
		CipherSuites:                        nil,
		PreferServerCipherSuites:            false,
		SessionTicketsDisabled:              false,
		SessionTicketKey:                    [32]byte{},
		ClientSessionCache:                  nil,
		UnwrapSession:                       nil,
		WrapSession:                         nil,
		MinVersion:                          0,
		MaxVersion:                          0,
		CurvePreferences:                    nil,
		DynamicRecordSizingDisabled:         false,
		Renegotiation:                       0,
		KeyLogWriter:                        nil,
		EncryptedClientHelloConfigList:      nil,
		EncryptedClientHelloRejectionVerify: nil,
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
