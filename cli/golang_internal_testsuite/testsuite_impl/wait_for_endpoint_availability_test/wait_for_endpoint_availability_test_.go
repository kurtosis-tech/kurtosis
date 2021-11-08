/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package wait_for_endpoint_availability_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	dockerGettingStartedImage                    = "docker/getting-started"
	datastoreServiceId        services.ServiceID = "docker-getting-started"
	dockerGettingStartedPort                                = 80
	healthCheckUrlSlug                           = ""
	healthyValue                                 = ""

	waitForStartupTimeBetweenPolls = 1
	waitForStartupMaxPolls         = 15
	waitInitialDelayMilliseconds   = 500
)

type WaitForEndpointAvailabilityTest struct {
}

func NewWaitForEndpointAvailabilityTest() *WaitForEndpointAvailabilityTest {
	return &WaitForEndpointAvailabilityTest{}
}

func (test WaitForEndpointAvailabilityTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test WaitForEndpointAvailabilityTest) Setup(enclaveCtx *networks.NetworkContext) (networks.Network, error) {
	return enclaveCtx, nil
}

func (test WaitForEndpointAvailabilityTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetworkContext := network.(*networks.NetworkContext)

	datastoreContainerConfigSupplier := getDatastoreContainerConfigSupplier()

	_, _, err := castedNetworkContext.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	port := uint32(dockerGettingStartedPort)

	if err := castedNetworkContext.WaitForHttpGetEndpointAvailability(datastoreServiceId, port, healthCheckUrlSlug, waitInitialDelayMilliseconds, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Service: %v is available", datastoreServiceId)

	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getDatastoreContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			dockerGettingStartedImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", dockerGettingStartedPort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
