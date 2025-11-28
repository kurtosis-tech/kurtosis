package main

import (
	server "github.com/kurtosis-tech/kurtosis/enclave-manager"
	"github.com/sirupsen/logrus"
)

const (
	enforceAuth = false
)

func main() {
	logrus.Info("Running the enclave manager from the enclave manager main package.")
	server.RunEnclaveManagerApiServer(enforceAuth)
}
