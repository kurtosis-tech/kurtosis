/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package main

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/kurtosis_core_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/api_container_availability_waiter/api_container_availability_waiter_consts"
	"net"
	"os"
	"time"
)

const (
	apiContainerServerAddress             = "localhost"
	maxWaitForAvailabilityRetries         = 5
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	errorExitCode = 1
)

// We don't want to return an enclave to the user until the API container gRPC server is actually available
// However, when we create an enclave, we can't assume that any container will be inside the enclave besides the API container
// We therefore can't use the gRPC server endpoints to verify that the API container is up
// Instead, we Docker exec this CLI which will block until the API container becomes available or a timeout is reached
func main() {
	dialUrl := fmt.Sprintf("%v:%v", apiContainerServerAddress, kurtosis_core_rpc_api_consts.ListenPort)
	for i := 0; i < maxWaitForAvailabilityRetries; i++ {
		conn, err := net.Dial(kurtosis_core_rpc_api_consts.ListenProtocol, dialUrl)
		if err == nil {
			conn.Close()
			os.Exit(api_container_availability_waiter_consts.SuccessExitCode)
		}
		time.Sleep(timeBetweenWaitForAvailabilityRetries)
	}
	fmt.Println(fmt.Sprintf(
		"The API container at %v didn't become available even after retrying %v times with %v between retries",
		dialUrl,
		maxWaitForAvailabilityRetries,
		timeBetweenWaitForAvailabilityRetries,
	))
	os.Exit(errorExitCode)
}


