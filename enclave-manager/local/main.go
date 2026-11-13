package main

import (
	"github.com/kurtosis-tech/kurtosis/api/golang/auth"
	server "github.com/kurtosis-tech/kurtosis/enclave-manager"
	"github.com/sirupsen/logrus"
)

const (
	enforceAuth = false
	noDomain    = ""
)

func main() {
	logrus.Info("Running the enclave manager from the enclave manager main package.")
	authToken, err := auth.GetServerToken()
	if err != nil {
		logrus.Fatal(err)
	}
	if err := server.RunEnclaveManagerApiServer(enforceAuth, noDomain, authToken); err != nil {
		logrus.Fatal(err)
	}
}
