package server

import (
	"github.com/kurtosis-tech/kurtosis/enclave-manager/api/golang/kurtosis_enclave_manager_api_bindings/kurtosis_enclave_manager_api_bindingsconnect"
	"time"
)

const (
	listenPort                = 8080
	grpcServerStopGracePeriod = 5 * time.Second
)

func RunEnclaveApiServer() {

	//srv := apiserver.NewWebserver()

	apiPath, handler := kurtosis_enclave_manager_api_bindingsconnect.New

	//logrus.Infof("Web server running and listening on port %d", listenPort)
	//apiServer := connect_server.NewConnectServer(
	//	listenPort,
	//	grpcServerStopGracePeriod,
	//	handler,
	//	apiPath,
	//)

}
