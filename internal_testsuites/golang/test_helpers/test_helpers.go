package test_helpers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_bindings"
	"github.com/kurtosis-tech/example-api-server/api/golang/example_api_server_rpc_api_consts"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_bindings"
	"github.com/kurtosis-tech/example-datastore-server/api/golang/datastore_rpc_api_consts"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
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
	waitForStartupMaxRetries = 30

	// File server wait for availability configuration
	waitForFileServerTimeoutMilliseconds  = 45000
	waitForFileServerIntervalMilliseconds = 100

	userServiceMountPointForTestFilesArtifact = "/static"

	// datastore server dummy test values
	testDatastoreKey   = "my-key"
	testDatastoreValue = "test-value"

	waitForGetAvaliabilityStalarkScript = `
def run(plan, args):
	get_recipe = GetHttpRequestRecipe(
		port_id = args["port_id"],
		endpoint = args["endpoint"],
	)
	plan.wait(recipe=get_recipe, field="code", assertion="==", target_value=200, interval=args["interval"], timeout=args["timeout"], service_name=args["service_name"])
`
	waitForGetAvaliabilityStalarkScriptParams = `{ "service_name": "%s", "port_id": "%s", "endpoint": "/%s", "interval": "%dms", "timeout": "%dms"}`

	noExpectedLogLines = 0

	dockerGettingStartedImage = "docker/getting-started"

	emptyPrivateIpAddrPlaceholder  = ""
	emptyCpuAllocationMillicpus    = 0
	emptyMemoryAllocationMegabytes = 0
	emptyApplicationProtocol       = ""

	artifactNamePrefix = "artifact-uploaded-via-helper-%v"

	defaultWaitTimeoutForTest  = "2m"
	defaultShouldReturnAllLogs = true
	defaultNumLogLines         = 0 // bc return all logs is true, what this is set to doesn't matter
)

var (
	emptyPrivatePorts            = map[string]*kurtosis_core_rpc_api_bindings.Port{}
	emptyFileArtifactMountPoints = map[string]string{}
	emptyEntrypointArgs          = []string{}
	emptyCmdArgs                 = []string{}
	emptyEnvVars                 = map[string]string{}

	// skip flaky tests period
	skipFlakyTestStartDate = time.Date(2023, 11, 10, 0, 0, 0, 0, time.UTC)
	oneWeekDays            = 7
	oneWeekAfterStartDate  = skipFlakyTestStartDate.AddDate(0, 0, oneWeekDays)
)

var fileServerPortSpec = &kurtosis_core_rpc_api_bindings.Port{
	Number:                   fileServerPrivatePortNum,
	TransportProtocol:        kurtosis_core_rpc_api_bindings.Port_TCP,
	MaybeApplicationProtocol: emptyApplicationProtocol,
	MaybeWaitTimeout:         defaultWaitTimeoutForTest,
	Locked:                   nil,
	Alias:                    nil,
}
var datastorePortSpec = &kurtosis_core_rpc_api_bindings.Port{
	Number:                   uint32(datastore_rpc_api_consts.ListenPort),
	TransportProtocol:        kurtosis_core_rpc_api_bindings.Port_TCP,
	MaybeApplicationProtocol: emptyApplicationProtocol,
	MaybeWaitTimeout:         defaultWaitTimeoutForTest,
	Locked:                   nil,
	Alias:                    nil,
}
var apiPortSpec = &kurtosis_core_rpc_api_bindings.Port{
	Number:                   uint32(example_api_server_rpc_api_consts.ListenPort),
	TransportProtocol:        kurtosis_core_rpc_api_bindings.Port_TCP,
	MaybeApplicationProtocol: emptyApplicationProtocol,
	MaybeWaitTimeout:         defaultWaitTimeoutForTest,
	Locked:                   nil,
	Alias:                    nil,
}

type GrpcAvailabilityChecker interface {
	IsAvailable(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type datastoreConfig struct {
	DatastoreIp   string `json:"datastoreIp"`
	DatastorePort uint16 `json:"datastorePort"`
}

func AddService(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	serviceName services.ServiceName,
	serviceConfigStarlark string,
) (
	*services.ServiceContext, error,
) {
	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig()
	starlarkScript := fmt.Sprintf(`
def run(plan):
	plan.add_service(name = "%s", config = %s)
`, serviceName, serviceConfigStarlark)
	starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlarkRunConfig)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error has occurred when running Starlark to add service")
	}
	if len(starlarkRunResult.ValidationErrors) > 0 {
		return nil, stacktrace.NewError("An error has occurred when validating Starlark to add service: %s", starlarkRunResult.ValidationErrors)
	}
	if starlarkRunResult.InterpretationError != nil {
		return nil, stacktrace.NewError("An error has occurred when interpreting Starlark to add service: %s", starlarkRunResult.InterpretationError)
	}
	if starlarkRunResult.ExecutionError != nil {
		return nil, stacktrace.NewError("An error has occurred when executing Starlark to add service: %s", starlarkRunResult.ExecutionError)
	}
	serviceContext, err := enclaveCtx.GetServiceContext(string(serviceName))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error has occurred when getting service added by Starlark")
	}
	return serviceContext, nil
}

func RemoveService(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	serviceName services.ServiceName,
) error {
	starlarkScript := fmt.Sprintf(`
def run(plan):
	plan.remove_service(name = "%s")
`, serviceName)
	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig()
	starlarkRunResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, starlarkScript, starlarkRunConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An error has occurred when running Starlark to remove service")
	}
	if len(starlarkRunResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error has occurred when validating Starlark to remove service: %s", starlarkRunResult.ValidationErrors)
	}
	if starlarkRunResult.InterpretationError != nil {
		return stacktrace.NewError("An error has occurred when interpreting Starlark to remove service: %s", starlarkRunResult.InterpretationError)
	}
	if starlarkRunResult.ExecutionError != nil {
		return stacktrace.NewError("An error has occurred when executing Starlark to remove service: %s", starlarkRunResult.ExecutionError)
	}
	return nil
}

func AddDatastoreService(
	ctx context.Context,
	serviceName services.ServiceName,
	enclaveCtx *enclaves.EnclaveContext,
) (
	resultServiceCtx *services.ServiceContext,
	resultClient datastore_rpc_api_bindings.DatastoreServiceClient,
	resultClientCloseFunc func(),
	resultErr error,
) {
	serviceConfigStarlark := getDatastoreServiceConfigStarlark()

	serviceCtx, err := AddService(ctx, enclaveCtx, serviceName, serviceConfigStarlark)
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

func ValidateDatastoreServiceHealthy(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName, portId string) error {
	serviceCtx, err := enclaveCtx.GetServiceContext(string(serviceName))
	if err != nil {
		return stacktrace.Propagate(err, "Error retrieving service context for service '%s'", serviceName)
	}
	ipAddr := serviceCtx.GetMaybePublicIPAddress()

	publicPort, found := serviceCtx.GetPublicPorts()[portId]
	if !found {
		return stacktrace.NewError("No public port found for service '%s' and port ID '%s'", serviceName, portId)
	}

	datastoreClient, datastoreClientConnCloseFunc, err := createDatastoreClient(
		ipAddr,
		publicPort.GetNumber(),
	)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating a new datastore client for service with name '%v' and IP address '%v'", serviceName, ipAddr)
	}
	defer datastoreClientConnCloseFunc()

	err = WaitForHealthy(context.Background(), datastoreClient, waitForStartupMaxRetries, waitForStartupTimeBetweenPolls)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred waiting for datastore service '%v' to become available", serviceName)
	}

	upsertArgs := &datastore_rpc_api_bindings.UpsertArgs{
		Key:   testDatastoreKey,
		Value: testDatastoreValue,
	}
	_, err = datastoreClient.Upsert(ctx, upsertArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred adding the test key to datastore service '%v'", serviceName)
	}

	getArgs := &datastore_rpc_api_bindings.GetArgs{
		Key: testDatastoreKey,
	}
	getResponse, err := datastoreClient.Get(ctx, getArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test key from datastore service '%v'", serviceName)
	}

	actualValue := getResponse.GetValue()
	if testDatastoreValue != actualValue {
		return stacktrace.NewError("Datastore service '%v' is storing value '%v' for the test key '%v', which doesn't match the expected value '%v'", serviceName, actualValue, testDatastoreKey, testDatastoreValue)
	}
	return nil
}

func AddAPIService(ctx context.Context, serviceName services.ServiceName, enclaveCtx *enclaves.EnclaveContext, datastorePrivateIp string) (*services.ServiceContext, example_api_server_rpc_api_bindings.ExampleAPIServerServiceClient, func(), error) {
	configFilepath, err := createApiConfigFile(datastorePrivateIp)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred creating the datastore config file")
	}
	artifactName := fmt.Sprintf(artifactNamePrefix, time.Now().Unix())
	_, _, err = enclaveCtx.UploadFiles(configFilepath, artifactName)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred uploading the datastore config file")
	}

	serviceConfigStarlark := getApiServiceServiceConfigStarlark(artifactName)

	serviceCtx, err := AddService(ctx, enclaveCtx, serviceName, serviceConfigStarlark)
	if err != nil {
		return nil, nil, nil, stacktrace.Propagate(err, "An error occurred adding the API service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[apiPortId]
	if !found {
		return nil, nil, nil, stacktrace.NewError("No API service public port found for port ID '%v'", apiPortId)
	}

	url := fmt.Sprintf("%v:%v", serviceCtx.GetMaybePublicIPAddress(), publicPort.GetNumber())
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func RunScriptWithDefaultConfig(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, script string) (*enclaves.StarlarkRunResult, error) {
	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig()
	return enclaveCtx.RunStarlarkScriptBlocking(ctx, script, starlarkRunConfig)
}

func SetupSimpleEnclaveAndRunScript(t *testing.T, ctx context.Context, testName string, script string) (*enclaves.StarlarkRunResult, error) {

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, _, destroyEnclaveFunc, err := CreateEnclave(t, ctx, testName)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer func() { _ = destroyEnclaveFunc() }()

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Executing Startosis script...")
	logrus.Debugf("Startosis script content: \n%v", script)

	return RunScriptWithDefaultConfig(ctx, enclaveCtx, script)
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

func StartFileServer(ctx context.Context, fileServerServiceId services.ServiceName, filesArtifactUUID services.FilesArtifactUUID, pathToCheckOnFileServer string, enclaveCtx *enclaves.EnclaveContext) (string, uint16, error) {
	filesArtifactMountPoints := map[string]services.FilesArtifactUUID{
		userServiceMountPointForTestFilesArtifact: filesArtifactUUID,
	}
	fileServerServiceConfigStarlark := getFileServerServiceConfigStarlark(filesArtifactMountPoints)
	serviceCtx, err := AddService(ctx, enclaveCtx, fileServerServiceId, fileServerServiceConfigStarlark)
	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}

	publicPort, found := serviceCtx.GetPublicPorts()[fileServerPortId]
	if !found {
		return "", 0, stacktrace.NewError("Expected to find public port for ID '%v', but none was found", fileServerPortId)
	}

	fileServerPublicIp := serviceCtx.GetMaybePublicIPAddress()
	fileServerPublicPortNum := publicPort.GetNumber()

	err = waitForFileServerAvailability(
		ctx,
		enclaveCtx,
		fileServerServiceId,
		fileServerPortId,
		pathToCheckOnFileServer,
		waitForFileServerIntervalMilliseconds,
		waitForFileServerTimeoutMilliseconds,
	)

	if err != nil {
		return "", 0, stacktrace.Propagate(err, "An error occurred waiting for the file server service to become available.")
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

func GetLogsResponse(
	t *testing.T,
	ctx context.Context,
	timeout time.Duration,
	kurtosisCtx *kurtosis_context.KurtosisContext,
	enclaveIdentifier string,
	serviceUuids map[services.ServiceUUID]bool,
	expectedLogLinesByService map[services.ServiceUUID][]string,
	shouldFollowLogs bool,
	logLineFilter *kurtosis_context.LogLineFilter,
) (
	error,
	map[services.ServiceUUID][]string,
	map[services.ServiceUUID]bool,
) {

	if expectedLogLinesByService == nil {
		return stacktrace.NewError("The 'expectedLogLinesByService' can't be nil because it is needed for handling the retry strategy"), nil, nil
	}

	receivedLogLinesByService := map[services.ServiceUUID][]string{}
	receivedNotFoundServiceGuids := map[services.ServiceUUID]bool{}
	var testEvaluationErr error

	serviceLogsStreamContentChan, cancelStreamUserServiceLogsFunc, err := kurtosisCtx.GetServiceLogs(ctx, enclaveIdentifier, serviceUuids, shouldFollowLogs, defaultShouldReturnAllLogs, defaultNumLogLines, logLineFilter)
	defer cancelStreamUserServiceLogsFunc()
	require.NoError(t, err, "An error occurred getting user service logs from user services with UUIDs '%+v' in enclave '%v' and with follow logs value '%v'", serviceUuids, enclaveIdentifier, shouldFollowLogs)

	shouldContinueInTheLoop := true

	ticker := time.NewTicker(timeout)

	for shouldContinueInTheLoop {
		select {
		case <-ticker.C:
			testEvaluationErr = stacktrace.NewError("Receiving stream logs in the test has reached the '%v' time out", timeout.String())
			shouldContinueInTheLoop = false
			break
		case serviceLogsStreamContent, isChanOpen := <-serviceLogsStreamContentChan:
			if !isChanOpen {
				shouldContinueInTheLoop = false
				break
			}

			serviceLogsByGuid := serviceLogsStreamContent.GetServiceLogsByServiceUuids()
			receivedNotFoundServiceGuids = serviceLogsStreamContent.GetNotFoundServiceUuids()

			for serviceUuid, serviceLogLines := range serviceLogsByGuid {
				receivedLogLines := []string{}
				for _, serviceLogLine := range serviceLogLines {
					receivedLogLines = append(receivedLogLines, serviceLogLine.GetContent())
				}
				if _, found := receivedLogLinesByService[serviceUuid]; found {
					receivedLogLinesByService[serviceUuid] = append(receivedLogLinesByService[serviceUuid], receivedLogLines...)
				} else {
					receivedLogLinesByService[serviceUuid] = receivedLogLines
				}
			}

			for serviceUuid, expectedLogLines := range expectedLogLinesByService {
				receivedLogLines, found := receivedLogLinesByService[serviceUuid]
				if len(expectedLogLines) == noExpectedLogLines && !found {
					receivedLogLines = []string{}
				} else if !found {
					return stacktrace.NewError("Expected to receive log lines for service with UUID '%v' but none was found in the received log lines by service map '%+v'", serviceUuid, receivedLogLinesByService), nil, nil
				}
				if len(receivedLogLines) != len(expectedLogLines) {
					break
				}
				shouldContinueInTheLoop = false
			}

			if !shouldContinueInTheLoop {
				break
			}
		}
	}

	return testEvaluationErr, receivedLogLinesByService, receivedNotFoundServiceGuids
}

func AddServicesWithLogLines(
	ctx context.Context,
	enclaveCtx *enclaves.EnclaveContext,
	logLinesByServiceName map[services.ServiceName][]string,
) (map[services.ServiceName]*services.ServiceContext, error) {

	servicesAdded := make(map[services.ServiceName]*services.ServiceContext, len(logLinesByServiceName))
	for serviceName, logLines := range logLinesByServiceName {
		serviceConfigStarlark := getServiceWithLogLinesServiceConfigStarlark(logLines)
		serviceCtx, err := AddService(ctx, enclaveCtx, serviceName, serviceConfigStarlark)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred adding service with name %v", serviceName)
		}
		servicesAdded[serviceName] = serviceCtx
	}
	return servicesAdded, nil
}

func GenerateRandomTempFile(byteSize int, filePathOptional string) (string, func(), error) {
	fileCreationSuccessful := false
	content := make([]byte, byteSize)
	_, err := rand.Read(content)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error generating random content for file")
	}

	var file *os.File
	if filePathOptional == "" {
		file, err = os.CreateTemp("", "")
		if err != nil {
			return "", nil, stacktrace.Propagate(err, "Error creating temporary file")
		}
	} else {
		file, err = os.Create(filePathOptional)
	}
	cleanFileFunc := func() {
		if err = os.Remove(file.Name()); err != nil {
			logrus.Warnf("Error removing file '%s' after test has finished", file.Name())
		}
	}
	defer func() {
		if err = file.Close(); err != nil {
			logrus.Warnf("Unexpected error closing temporary random file after creating it")
		}
		if !fileCreationSuccessful {
			cleanFileFunc()
		}
	}()

	_, err = file.Write(content)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "Error writing content to temporary file")
	}
	fileCreationSuccessful = true
	return file.Name(), cleanFileFunc, nil
}

// ====================================================================================================
//
//	Private Helper Methods
//
// ====================================================================================================
func getDatastoreServiceConfigStarlark() string {
	return services.GetServiceConfigStarlark(datastoreImage, map[string]*kurtosis_core_rpc_api_bindings.Port{datastorePortId: datastorePortSpec}, emptyFileArtifactMountPoints, emptyEntrypointArgs, emptyCmdArgs, emptyEnvVars, emptyPrivateIpAddrPlaceholder, emptyCpuAllocationMillicpus, emptyMemoryAllocationMegabytes, 0, 0, "", nil, "", "")
}

func getApiServiceServiceConfigStarlark(apiConfigArtifactName string) string {
	startCmd := []string{
		"./example-api-server.bin",
		"--config",
		path.Join(configMountpathOnApiContainer, configFilename),
	}

	return services.GetServiceConfigStarlark(apiServiceImage, map[string]*kurtosis_core_rpc_api_bindings.Port{apiPortId: apiPortSpec}, map[string]string{configMountpathOnApiContainer: apiConfigArtifactName}, emptyEntrypointArgs, startCmd, emptyEnvVars, emptyPrivateIpAddrPlaceholder, emptyCpuAllocationMillicpus, emptyMemoryAllocationMegabytes, 0, 0, "", nil, "", "")
}

func createApiConfigFile(datastoreIP string) (string, error) {
	tempDirpath, err := os.MkdirTemp("", "")
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

	if err := os.WriteFile(tempFilepath, configBytes, os.ModePerm); err != nil {
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

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return "", stacktrace.Propagate(err,
			"An error occurred reading the response body when getting the contents of file '%v'", realtiveFilepath)
	}

	bodyStr := string(bodyBytes)
	return bodyStr, nil
}

func getFileServerServiceConfigStarlark(filesArtifactMountPoints map[string]services.FilesArtifactUUID) string {
	filesArtifactMountPointsStr := map[string]string{}
	for k, v := range filesArtifactMountPoints {
		filesArtifactMountPointsStr[k] = string(v)
	}

	return services.GetServiceConfigStarlark(fileServerServiceImage, map[string]*kurtosis_core_rpc_api_bindings.Port{fileServerPortId: fileServerPortSpec}, filesArtifactMountPointsStr, emptyEntrypointArgs, emptyCmdArgs, emptyEnvVars, emptyPrivateIpAddrPlaceholder, emptyCpuAllocationMillicpus, emptyMemoryAllocationMegabytes, 0, 0, "", nil, "", "")
}

func createDatastoreClient(ipAddr string, portNum uint16) (datastore_rpc_api_bindings.DatastoreServiceClient, func(), error) {
	url := fmt.Sprintf("%v:%v", ipAddr, portNum)
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

func waitForFileServerAvailability(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, serviceName services.ServiceName, portId string, endpoint string, initialDelayMilliseconds uint32, timeoutMilliseconds uint32) error {
	starlarkParams := fmt.Sprintf(waitForGetAvaliabilityStalarkScriptParams, serviceName, portId, endpoint, initialDelayMilliseconds, timeoutMilliseconds)
	starlarkRunConfig := starlark_run_config.NewRunStarlarkConfig(starlark_run_config.WithSerializedParams(starlarkParams))
	runResult, err := enclaveCtx.RunStarlarkScriptBlocking(ctx, waitForGetAvaliabilityStalarkScript, starlarkRunConfig)
	if err != nil {
		return stacktrace.Propagate(err, "An unexpected error has occurred getting endpoint availability using Starlark")
	}
	if runResult.ExecutionError != nil {
		return stacktrace.NewError("An error has occurred getting endpoint availability during Starlark due to execution error %s", runResult.ExecutionError.GetErrorMessage())
	}
	if runResult.InterpretationError != nil {
		return stacktrace.NewError("An error has occurred getting endpoint availability using Starlark due to interpretation error %s", runResult.InterpretationError.GetErrorMessage())
	}
	if len(runResult.ValidationErrors) > 0 {
		return stacktrace.NewError("An error has occurred getting endpoint availability during Starlark due to validation error %v", runResult.ValidationErrors)
	}
	return nil
}

func getServiceWithLogLinesServiceConfigStarlark(logLines []string) string {

	entrypointArgs := []string{"/bin/sh", "-c"}

	var logLinesWithQuotes []string
	for _, logLine := range logLines {
		logLineWithQuote := fmt.Sprintf("\"%s\"", logLine)
		logLinesWithQuotes = append(logLinesWithQuotes, logLineWithQuote)
	}

	logLineSeparator := " "
	logLinesStr := strings.Join(logLinesWithQuotes, logLineSeparator)
	echoLogLinesLoopCmdStr := fmt.Sprintf("for logLine in %s; do echo \"$logLine\"; done;sleep 10000s", logLinesStr)

	cmdArgs := []string{echoLogLinesLoopCmdStr}

	return services.GetServiceConfigStarlark(dockerGettingStartedImage, emptyPrivatePorts, emptyFileArtifactMountPoints, entrypointArgs, cmdArgs, emptyEnvVars, emptyPrivateIpAddrPlaceholder, emptyCpuAllocationMillicpus, emptyMemoryAllocationMegabytes, 0, 0, "", nil, "", "")
}

func SkipFlakyTest(t *testing.T, testName string) {
	now := time.Now()
	if now.Before(oneWeekAfterStartDate) {
		t.Skipf("Skipping %s, because it is too noisy, until %s or until we fix the flakyness", testName, oneWeekAfterStartDate)
	}
}
