package main

import server "github.com/kurtosis-tech/kurtosis/enclave-manager"

const (
	enforceAuth = false
)

func main() {
	server.RunEnclaveManagerApiServer(enforceAuth)
}
