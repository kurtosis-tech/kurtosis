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
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"net/http"
	"os"
	"path"
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

	fileServerServiceImage   = "flashspys/nginx-static"
	fileServerPortId         = "http"
	fileServerPrivatePortNum = 80

	waitForStartupTimeBetweenPolls = 500
	/*
		NOTE: on 2022-05-16 this failed with the following error so we bumped the num polls to 20.

		time="2022-05-16T23:58:21Z" level=info msg="Sanity-checking that all 4 datastore services added via the module work as expected..."
		--- FAIL: TestModule (21.46s)
			module_test.go:81:
					Error Trace:	module_test.go:81
					Error:      	Received unexpected error:
									The service didn't return a success code, even after 15 retries with 1000 milliseconds in between retries
									 --- at /home/circleci/project/internal_testsuites/golang/test_helpers/test_helpers.go:179 (WaitForHealthy) ---
									Caused by: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:49188: connect: connection refused"
					Test:       	TestModule
					Messages:   	An error occurred waiting for the datastore service to become available

		NOTE: On 2022-05-21 this failed again at 20s. I opened the enclave logs and it's weird because nothing is failing and
		the datastore service is showing itself as up *before* we even start the check-if-available wait. We're in crunch mode
		so I'm going to bump this up to 30s, but I suspect there's some sort of nondeterministic underlying failure happening.
	*/
	waitForStartupMaxRetries     = 30
	waitInitialDelayMilliseconds = 0
	waitForAvailabilityBodyText  = ""

	userServiceMountPointForTestFilesArtifact = "/static"

	// datastore server dummy test values
	testDatastoreKey   = "my-key"
	testDatastoreValue = "test-value"
)

var fileServerPortSpec = services.NewPortSpec(
	fileServerPrivatePortNum,
	services.PortProtocol_TCP,
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
	containerConfig := getDatastoreContainerConfig()

	serviceCtx, err := enclaveCtx.AddService(serviceId, containerConfig)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the datastore service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[datastorePortId]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No datastore public port found for port ID '%v'", datastorePortId)
	}

	publicIp := serviceCtx.GetMaybePublicIPAddress()
	publicPortNum := publicPort.GetNumber()
	client, clientCloseFunc, err := createDatastoreClient(publicIp, publicPortNum)
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

func ValidateDatastoreServiceHealthy(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceId services.ServiceID, portId string) error {
	serviceCtx, err := enclaveCtx.GetServiceContext(serviceId)
	if err != nil {
		return stacktrace.Propagate(err, "Error retrieving service context for service '%s'", serviceId)
	}
	ipAddr := serviceCtx.GetMaybePublicIPAddress()

	publicPort, found := serviceCtx.GetPublicPorts()[portId]
	if !found {
		return stacktrace.Propagate(err, "No public port found for service '%s' and port ID '%s'", serviceId, portId)
	}

	datastoreClient, datastoreClientConnCloseFunc, err := createDatastoreClient(
		ipAddr,
		publicPort.GetNumber(),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with ID '%v' and IP address '%v'", serviceId, ipAddr)
	}
	defer datastoreClientConnCloseFunc()

	err = WaitForHealthy(context.Background(), datastoreClient, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for datastore service '%v' to become available", serviceId)
	}

	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testDatastoreKey,
		Value: testDatastoreValue,
	}
	_, err = datastoreClient.Upsert(ctx, upsertArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the test key to datastore service '%v'", serviceId)
	}

	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testDatastoreKey,
	}
	getResponse, err := datastoreClient.Get(ctx, getArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key from datastore service '%v'", serviceId)
	}

	actualValue := getResponse.GetValue()
	if testDatastoreValue != actualValue {
		return stacktrace.NewError("Datastore service '%v' is storing value '%v' for the test key '%v', which doesn't match the expected value '%v'", serviceId, actualValue, testDatastoreKey, testDatastoreValue)
	}
	return nil
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

	containerConfig := getApiServiceContainerConfig(datastoreConfigArtifactUuid)

	serviceCtx, err := enclaveCtx.AddServiceToPartition(serviceId, partitionId, containerConfig)
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

	if err := WaitForHealthy(ctx, client, apiWaitForStartupMaxPolls, apiWaitForStartupDelayMilliseconds); err != nil {
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

func StartFileServer(fileServerServiceId services.ServiceID, filesArtifactUUID services.FilesArtifactUUID, pathToCheckOnFileServer string, enclaveCtx *enclaves.EnclaveContext) (string, uint16, error) {
	filesArtifactMountPoints := map[services.FilesArtifactUUID]string{
		filesArtifactUUID: userServiceMountPointForTestFilesArtifact,
	}
	fileServerContainerConfig := getFileServerContainerConfig(filesArtifactMountPoints)
	serviceCtx, err := enclaveCtx.AddService(fileServerServiceId, fileServerContainerConfig)
	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	if !found {
		return "", 0, stacktrace.NewError("Expected to find public port for ID '%v', but none was found", fileServerPortId)
	}

	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()

	err = enclaveCtx.WaitForHttpGetEndpointAvailability(
		fileServerServiceId,
		fileServerPrivatePortNum,
		pathToCheckOnFileServer,
		waitInitialDelayMilliseconds,
		waitForStartupMaxRetries,
		waitForStartupTimeBetweenPolls,
		waitForAvailabilityBodyText,
	)

	if err != nil {
		return "", 0, stacktrace.NewError("An error occurred waiting for the file server service to become available.")
	}

	logrus.Infof("Added file server service with public IP '%v' and port '%v'", fileServerPublicIp,
		fileServerPublicPortNum)

	return fileServerPublicIp, fileServerPublicPortNum, nil
}

// Compare the file contents on the server against expectedContent and see if they match.
func CheckFileContents(serverIP string, port uint16, relativeFilepath string, expectedContents string) error {
	fileContents, err := getFileContents(serverIP, port, relativeFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting '%s' contents", relativeFilepath)
	}
	if expectedContents != fileContents {
		return stacktrace.NewError(
			"The contents of '%s' do not match the expected content '%s'",
			fileContents,
			expectedContents,
		)
	}
	return nil
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func getDatastoreContainerConfig() *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(
		datastoreImage,
	).WithUsedPorts(map[string]*services.PortSpec{
		datastorePortId: datastorePortSpec,
	}).Build()
	return containerConfig
}

func getApiServiceContainerConfig(apiConfigArtifactUuid services.FilesArtifactUUID) *services.ContainerConfig {
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

	return containerConfig
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

func getFileContents(ipAddress string, portNum uint16, realtiveFilepath string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v:%v/%v", ipAddress, portNum, realtiveFilepath))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the contents of file '%v'", realtiveFilepath)
	}
	body := resp.Body
	defer func() {
		if err := body.Close(); err != nil {
			logrus.Warnf("We tried to close the response body, but doing so threw an error:\n%v", err)
		}
	}()

	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err,
			"An error occurred reading the response body when getting the contents of file '%v'", realtiveFilepath)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}

func getFileServerContainerConfig(filesArtifactMountPoints map[services.FilesArtifactUUID]string) *services.ContainerConfig {
	containerConfig := services.NewContainerConfigBuilder(
		fileServerServiceImage,
	).WithUsedPorts(map[string]*services.PortSpec{
		fileServerPortId: fileServerPortSpec,
	}).WithFiles(
		filesArtifactMountPoints,
	).Build()
	return containerConfig
}

func createDatastoreClient(ipAddr string, portNum uint16) (datastore_rpc_api_bindings.DatastoreServiceClient, func(), error) {
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
