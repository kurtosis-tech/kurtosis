/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
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
	waitInitialDelaySeconds        = 1
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

	containerCreationConfig, runConfigFunc := getDatastoreServiceConfigurations()

	_, _, err := castedNetworkContext.AddService(datastoreServiceId, containerCreationConfig, runConfigFunc)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	port := uint32(datastorePort)

	if err := castedNetworkContext.WaitForEndpointAvailability(datastoreServiceId, port, healthCheckUrlSlug, waitInitialDelaySeconds, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue); err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	logrus.Infof("Service: %v is available", datastoreServiceId)

	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getDatastoreServiceConfigurations() (*services.ContainerCreationConfig, func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error)) {
	containerCreationConfig := getDatastoreServiceContainerCreationConfig()

	runConfigFunc := getDatastoreServiceRunConfigFunc()
	return containerCreationConfig, runConfigFunc
}

func getDatastoreServiceContainerCreationConfig() *services.ContainerCreationConfig {
	containerCreationConfig := services.NewContainerCreationConfigBuilder(
		datastoreImage,
	).WithUsedPorts(
		map[string]bool{fmt.Sprintf("%v/tcp", datastorePort): true},
	).Build()
	return containerCreationConfig
}

func getDatastoreServiceRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	runConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		return services.NewContainerRunConfigBuilder().Build(), nil
	}
	return runConfigFunc
}
