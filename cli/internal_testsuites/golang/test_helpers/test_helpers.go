package test_helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

const (
	configFilename                = "config.json"
	configMountpathOnApiContainer = "/config"

	datastoreImage  = "kurtosistech/example-datastore-server"
	apiServiceImage = "kurtosistech/example-api-server"

	datastorePortId string = "rpc"
	apiPortId       string = "rpc"

	datastoreWaitForStartupMaxPolls          = 10
	datastoreWaitForStartupDelayMilliseconds = 1000

	apiWaitForStartupMaxPolls          = 10
	apiWaitForStartupDelayMilliseconds = 1000

	defaultPartitionId = ""
)

var datastorePortSpec = services.NewPortSpec(
	datastore_rpc_api_consts.ListenPort,
	services.PortProtocol_TCP,
)
var apiPortSpec = services.NewPortSpec(
	example_api_server_rpc_api_consts.ListenPort,
	services.PortProtocol_TCP,
)

type GrpcAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort uint16 `json:"datastorePort"`
}

func AddDatastoreService(
	ctx context.Context,
	serviceId services.ServiceID,
	enclaveCtx *enclaves.EnclaveContext,
) (
	resultServiceCtx *services.ServiceContext,
	resultClient datastore_rpc_api_bindings.DatastoreServiceClient,
	resultClientCloseFunc func(),
	resultErr error,
) {
	containerConfigSupplier := getDatastoreContainerConfigSupplier()

	serviceCtx, err := enclaveCtx.AddService(serviceId, containerConfigSupplier)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[datastorePortId]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No datastore public port found for port ID '%v'", datastorePortId)
	}

	publicIp := serviceCtx.GetMaybePublicIPAddress()
	publicPortNum := publicPort.GetNumber()
	client, clientCloseFunc, err := CreateDatastoreClient(publicIp, publicPortNum)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the datastore client for IP '%v' and port '%v'",
			publicIp,
			publicPortNum,
		)
	}

	if err := WaitForHealthy(ctx, client, datastoreWaitForStartupMaxPolls, datastoreWaitForStartupDelayMilliseconds); err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the datastore service to become available")
	}
	return serviceCtx, client, clientCloseFunc, nil
}

func CreateDatastoreClient(ipAddr string, portNum uint16) (datastore_rpc_api_bindings.DatastoreServiceClient, func(), error) {
	url := fmt.Sprintf("%v:%v", ipAddr, portNum)
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred connecting to datastore service on URL '%v'", url)
	}
	clientCloseFunc := func() {
		if err := conn.Close(); err != nil {
			logrus.Warnf("We tried to close the datastore client, but doing so threw an error:\n%v", err)
		}
	}
	client := datastore_rpc_api_bindings.NewDatastoreServiceClient(conn)
	return client, clientCloseFunc, nil
}

func AddAPIService(ctx context.Context, serviceId services.ServiceID, enclaveCtx *enclaves.EnclaveContext, datastorePrivateIp string) (*services.ServiceContext, example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func(), error) {
	serviceCtx, client, clientCloseFunc, err := AddAPIServiceToPartition(ctx, serviceId, enclaveCtx, datastorePrivateIp, defaultPartitionId)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding API service to default partition")
	}
	return serviceCtx, client, clientCloseFunc, nil
}

func AddAPIServiceToPartition(ctx context.Context, serviceId services.ServiceID, enclaveCtx *enclaves.EnclaveContext, datastorePrivateIp string, partitionId enclaves.PartitionID) (*services.ServiceContext, example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func(), error) {
	configFilepath, err := createApiConfigFile(datastorePrivateIp)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the datastore config file")
	}
	datastoreConfigArtifactUuid, err := enclaveCtx.UploadFiles(configFilepath)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred uploading the datastore config file")
	}

	containerConfigSupplier := getApiServiceContainerConfigSupplier(datastoreConfigArtifactUuid)

	serviceCtx, err := enclaveCtx.AddServiceToPartition(serviceId, partitionId, containerConfigSupplier)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[apiPortId]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No API service public port found for port ID '%v'", apiPortId)
	}

	url := fmt.Sprintf("%v:%v", serviceCtx.GetMaybePublicIPAddress(), publicPort.GetNumber())
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred connecting to API service on URL '%v'", url)
	}
	clientCloseFunc := func() {
		if err := conn.Close(); err != nil {
			logrus.Warnf("We tried to close the API service client, but doing so threw an error:\n%v", err)
		}
	}
	client := example_api_server_rpc_api_bindings.NewExampleAPIServerServiceClient(conn)

	if err := WaitForHealthy(context.Background(), client, apiWaitForStartupMaxPolls, apiWaitForStartupDelayMilliseconds); err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred waiting for the API service to become available")
	}
	return serviceCtx, client, clientCloseFunc, nil
}

func WaitForHealthy(ctx context.Context, client GrpcAvailabilityChecker, retries uint32, retriesDelayMilliseconds uint32) error {
	var (
		emptyArgs = &empty.Empty{}
		err       error
	)

	for i := uint32(0); i < retries; i++ {
		_, err = client.IsAvailable(ctx, emptyArgs)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(retriesDelayMilliseconds) * time.Millisecond)
	}

	if err != nil {
		return stacktrace.Propagate(
			err,
			"The service didn't return a success code, even after %v retries with %v milliseconds in between retries",
			retries,
			retriesDelayMilliseconds,
		)
	}

	return nil
}

// ====================================================================================================
//                                      Private Helper Methods
// ====================================================================================================
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

func getApiServiceContainerConfigSupplier(apiConfigArtifactUuid services.FilesArtifactUUID) func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		startCmd := []string{
			"./example-api-server.bin",
			"--config",
			path.Join(configMountpathOnApiContainer, configFilename),
		}

		containerConfig := services.NewContainerConfigBuilder(
			apiServiceImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			apiPortId: apiPortSpec,
		}).WithFiles(map[services.FilesArtifactUUID]string{
			apiConfigArtifactUuid: configMountpathOnApiContainer,
		}).WithCmdOverride(startCmd).Build()

		return containerConfig, nil
	}

	return containerConfigSupplier
}

func createApiConfigFile(datastoreIP string) (string, error) {
	tempDirpath, err := ioutil.TempDir("", "")
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating a temporary directory to house the datastore config file")
	}
	tempFilepath := path.Join(tempDirpath, configFilename)

	configObj := datastoreConfig{
		DatastoreIp:   datastoreIP,
		DatastorePort: datastore_rpc_api_consts.ListenPort,
	}
	configBytes, err := json.Marshal(configObj)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the config to JSON")
	}

	if err := ioutil.WriteFile(tempFilepath, configBytes, os.ModePerm); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing the serialized config JSON to file")
	}

	return tempFilepath, nil
}

func SkipTestInKubernetes(t *testing.T) {
	if IsInKubernetes() {
		t.Skip("Building testsuite Against Kubernetes, Skipping this test as functionality is not expected in Kubernetes")
	}
}

// Returns true if the testsuite is being built against Kubernetes
func IsInKubernetes() bool {
	return len(os.Getenv("TESTING_KUBERNETES")) > 0
}
