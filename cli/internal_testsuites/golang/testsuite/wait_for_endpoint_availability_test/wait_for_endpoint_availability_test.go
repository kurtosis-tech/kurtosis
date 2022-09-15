package wait_for_endpoint_availability_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "wait-for-endpoint-availability"
	isPartitioningEnabled = false

	dockerGettingStartedImage                       = "docker/getting-started"
	exampleServiceId             services.ServiceID = "docker-getting-started"
	exampleServicePortId = "http"
	exampleServicePrivatePortNum                    = 80
	healthCheckUrlSlug                              = ""
	healthyValue                                    = ""

	waitForStartupTimeBetweenPolls = 1
	waitForStartupMaxPolls         = 15
	waitInitialDelayMilliseconds   = 500
)
var exampleServicePrivatePortSpec = services.NewPortSpec(
	exampleServicePrivatePortNum,
	services.PortProtocol_TCP,
)

func TestWaitForEndpointAvailabilityFunction(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	configSupplier := getExampleServiceConfigSupplier()
	config, err := configSupplier()
	_, err = enclaveCtx.AddService(exampleServiceId, config)
	require.NoError(t, err, "An error occurred adding the datastore service")

	// ------------------------------------- TEST RUN ----------------------------------------------

	portUint32 := uint32(exampleServicePrivatePortNum)

	require.NoError(
		t,
		enclaveCtx.WaitForHttpGetEndpointAvailability(exampleServiceId, portUint32, healthCheckUrlSlug, waitInitialDelayMilliseconds, waitForStartupMaxPolls, waitForStartupTimeBetweenPolls, healthyValue),
		"An error occurred waiting for the datastore service to become available",
	)
	logrus.Infof("Service: %v is available", exampleServiceId)
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getExampleServiceConfigSupplier() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			dockerGettingStartedImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			exampleServicePortId: exampleServicePrivatePortSpec,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
