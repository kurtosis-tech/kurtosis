package main

import "github.com/kurtosis-tech/kurtosis/enclave-manager/server"

const (
	enforceAuth = false
)

func main() {
	server.RunEnclaveManagerApiServer(enforceAuth)
}
