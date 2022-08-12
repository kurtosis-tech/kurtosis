package basic_datastore_and_api_test

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "bulk-api"
	isPartitioningEnabled = false

	datastoreImage                        = "kurtosistech/example-datastore-server"
	datastorePortId    string             = "rpc"
	datastoreServiceId services.ServiceID = "datastore"
)

var datastorePortSpec = services.NewPortSpec(
	datastore_rpc_api_consts.ListenPort,
	services.PortProtocol_TCP,
)

func TestAddingDatastoreServicesInBulk(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, destroyEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer destroyEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	datastoreServiceConfigSuppliers := map[services.ServiceID]func(string) (*services.ContainerConfig, error){}
	for i := 0; i < 3; i++ {
		serviceID := fmt.Sprintf("%v-%v", datastoreServiceId, i)
		datastoreServiceConfigSuppliers[services.ServiceID(serviceID)] = getDatastoreContainerConfigSupplier()
	}

	logrus.Infof("Adding three datastore services simultaneously...")
	successfulServiceCtx, failedServiceErrs, err := enclaveCtx.AddServices(datastoreServiceConfigSuppliers)
	require.NoError(t, err, "An error occurred adding the datastore services to the enclave")
	logrus.Infof("Added datastore service")
	// ------------------------------------- TEST RUN ----------------------------------------------
	require.Equal(t, len(successfulServiceCtx), 3)
	require.Equal(t, len(failedServiceErrs), 0)
}

func getDatastoreContainerConfigSupplier() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			datastorePortId: datastorePortSpec,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}