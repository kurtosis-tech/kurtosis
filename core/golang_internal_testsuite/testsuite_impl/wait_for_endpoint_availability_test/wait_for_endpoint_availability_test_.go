/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package wait_for_endpoint_availability_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	datastoreImage                        = "kurtosistech/example-microservices_datastore"
	datastoreServiceId services.ServiceID = "datastore"
	datastorePort                         = 1323
	healthCheckUrlSlug                    = "health"
	healthyValue                          = "healthy"

	waitForStartupTimeBetweenPolls = 1
	waitForStartupMaxPolls         = 15
	waitInitialDelayMilliseconds   = 500
)

type WaitForEndpointAvailabilityTest struct {
	datastoreImage string
}

func NewWaitForEndpointAvailabilityTest(datastoreImage string) *WaitForEndpointAvailabilityTest {
	return &WaitForEndpointAvailabilityTest{datastoreImage: datastoreImage}
}

func (test WaitForEndpointAvailabilityTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (test WaitForEndpointAvailabilityTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	return networkCtx, nil
}

func (test WaitForEndpointAvailabilityTest) Run(network networks.Network) error {
	// Necessary because Go doesn't have generics
	castedNetworkContext := network.(*networks.NetworkContext)

	datastoreContainerConfigSupplier := getDatastoreContainerConfigSupplier()

	_, _, err := castedNetworkContext.AddService(datastoreServiceId, datastoreContainerConfigSupplier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	port := uint32(datastorePort)

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
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(
			map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
