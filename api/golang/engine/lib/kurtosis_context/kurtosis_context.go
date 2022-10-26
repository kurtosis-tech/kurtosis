package kurtosis_context

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/user_service_logs_read_closer"
	"github.com/kurtosis-tech/kurtosis/api/golang/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"time"
)

const (
	// NOTE: This needs to be 127.0.0.1 rather than 0.0.0.0, because Windows machines don't translate 0.0.0.0 -> 127.0.0.1
	localHostIPAddressStr = "127.0.0.1"

	DefaultGrpcEngineServerPortNum = uint16(9710)

	DefaultGrpcProxyEngineServerPortNum = uint16(9711)

	// Blank tells the engine server to use the default
	defaultApiContainerVersionTag = ""

	//Left the connection open from the client-side for 4 days
	maxAllowedWebsocketConnectionDurationOnClientSide = 96 * time.Hour
)

var apiContainerLogLevel = logrus.DebugLevel

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
type KurtosisContext struct {
	client kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// NewKurtosisContextFromLocalEngine
// Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
func NewKurtosisContextFromLocalEngine() (*KurtosisContext, error) {
	ctx := context.Background()
	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, DefaultGrpcEngineServerPortNum)

	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	conn, err := grpc.Dial(kurtosisEngineSocketStr, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Engine Server at '%v'",
			kurtosisEngineSocketStr,
		)
	}

	engineServiceClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	if err := validateEngineApiVersion(ctx, engineServiceClient); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the Kurtosis engine API version")
	}

	kurtosisContext := &KurtosisContext{
		client: engineServiceClient,
	}

	return kurtosisContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) CreateEnclave(
	ctx context.Context,
	enclaveId enclaves.EnclaveID,
	isPartitioningEnabled bool,
) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId:              string(enclaveId),
		ApiContainerVersionTag: defaultApiContainerVersionTag,
		ApiContainerLogLevel:   apiContainerLogLevel.String(),
		IsPartitioningEnabled:  isPartitioningEnabled,
	}

	response, err := kurtosisCtx.client.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave with ID '%v'", enclaveId)
	}

	enclaveContext, err := newEnclaveContextFromEnclaveInfo(response.EnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from a newly-created enclave; this should never happen")
	}

	return enclaveContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) GetEnclaveContext(ctx context.Context, enclaveId enclaves.EnclaveID) (*enclaves.EnclaveContext, error) {
	response, err := kurtosisCtx.client.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting enclaves",
		)
	}

	allEnclaveInfo := response.EnclaveInfo
	enclaveInfo, found := allEnclaveInfo[string(enclaveId)]
	if !found {
		return nil, stacktrace.Propagate(err, "No enclave with ID '%v' found", enclaveId)
	}

	enclaveCtx, err := newEnclaveContextFromEnclaveInfo(enclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from the returned enclave info")
	}

	return enclaveCtx, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) GetEnclaves(ctx context.Context) (map[enclaves.EnclaveID]bool, error) {
	response, err := kurtosisCtx.client.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting enclaves",
		)
	}

	result := map[enclaves.EnclaveID]bool{}
	for enclaveId := range response.EnclaveInfo {
		result[enclaves.EnclaveID(enclaveId)] = true
	}

	return result, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) StopEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	stopEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.StopEnclave(ctx, stopEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave with ID '%v'", enclaveId)
	}

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) DestroyEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave with ID '%v'", enclaveId)
	}

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) Clean(ctx context.Context, shouldCleanAll bool) (map[string]bool, error) {
	cleanArgs := &kurtosis_engine_rpc_api_bindings.CleanArgs{
		ShouldCleanAll: shouldCleanAll,
	}
	cleanResponse, err := kurtosisCtx.client.Clean(ctx, cleanArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when trying to perform a clean with the clean-all arg set to '%v'", shouldCleanAll)
	}

	return cleanResponse.RemovedEnclaveIds, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis/engine-lib-documentation
func (kurtosisCtx *KurtosisContext) GetUserServiceLogs(
	ctx context.Context,
	enclaveID enclaves.EnclaveID,
	userServiceGUIDs map[services.ServiceGUID]bool,
) (map[services.ServiceGUID][]string, error) {

	userServiceLogsByUserServiceGUID := map[services.ServiceGUID][]string{}

	getUserServiceLogsArgs := newGetUserServiceLogsArgs(enclaveID, userServiceGUIDs)

	gutUserServiceLogsResponse, err := kurtosisCtx.client.GetUserServiceLogs(ctx, getUserServiceLogsArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting user service logs using args '%+v'", getUserServiceLogsArgs)
	}

	for userServiceGUIDStr, userServiceLogLine := range gutUserServiceLogsResponse.UserServiceLogsByUserServiceGuid {
		userServiceGUID := services.ServiceGUID(userServiceGUIDStr)
		userServiceLogsByUserServiceGUID[userServiceGUID] = userServiceLogLine.Line
	}

	return userServiceLogsByUserServiceGUID, nil
}

func (kurtosisCtx *KurtosisContext) StreamUserServiceLogs(
	ctx context.Context,
	enclaveID enclaves.EnclaveID,
	userServiceGUIDs map[services.ServiceGUID]bool,
)(
	map[services.ServiceGUID]io.ReadCloser,
	error,
) {

	ctxWithCancel, cancelCtxFunc := context.WithTimeout(ctx, maxAllowedWebsocketConnectionDurationOnClientSide)
	shouldCancelCtx := false
	defer func() {
		if shouldCancelCtx {
			cancelCtxFunc()
		}
	}()

	userServiceReadCloserLogsByServiceGuid := newUserServiceReadCloserLogsByServiceGuid(userServiceGUIDs)

	getUserServiceLogsArgs := newGetUserServiceLogsArgs(enclaveID, userServiceGUIDs)

	stream, err := kurtosisCtx.client.StreamUserServiceLogs(ctxWithCancel, getUserServiceLogsArgs)
	if err != nil {
		shouldCancelCtx = true
		return nil, stacktrace.Propagate(err, "An error occurred streaming user service logs using args '%+v'", getUserServiceLogsArgs)
	}

	go func() {
		receiveStreamLogsLoop:
			for {
				getUserServiceLogResponse, errReceivingStream := stream.Recv()
				if errReceivingStream != nil {
					if errReceivingStream == io.EOF {
						break receiveStreamLogsLoop
					}
					logrus.Errorf("An error occurred receveing user service logs stream for user services '%+v' in enclave '%v'. Error:\n%v", userServiceGUIDs, enclaveID, errReceivingStream)
					if errCloseSend := stream.CloseSend(); errCloseSend != nil {
						logrus.Errorf("Streaming user service logs has thrown an error, so we tried to close the send direction of the stream, but an error was thrown:\n%v", errCloseSend)
					}
					break receiveStreamLogsLoop
				}
				for userServiceGuidStr, userServiceLogLine := range getUserServiceLogResponse.UserServiceLogsByUserServiceGuid {
					userServiceGuid := services.ServiceGUID(userServiceGuidStr)
					 if err := streamNewUserServiceLogLines(userServiceGuid, userServiceReadCloserLogsByServiceGuid, userServiceLogLine); err != nil {
						 logrus.Errorf("An error occurred streaming new user service log lines '%+v' for user service with GUID '%v'. Error:\n%v", userServiceLogLine, userServiceGuid, err)
						 break receiveStreamLogsLoop
					 }
				}
			}

		//Closing all the open resources
		cancelCtxFunc()
		closeAllUserServiceReadCloserLogs(userServiceReadCloserLogsByServiceGuid)
	}()

	return userServiceReadCloserLogsByServiceGuid,  nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func newEnclaveContextFromEnclaveInfo(
	enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) (*enclaves.EnclaveContext, error) {

	enclaveContainersStatus := enclaveInfo.GetContainersStatus()
	if enclaveContainersStatus != kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING {
		return nil, stacktrace.NewError(
			"Enclave containers status was '%v', but we can't create an enclave context from a non-running enclave",
			enclaveContainersStatus,
		)
	}

	enclaveApiContainerStatus := enclaveInfo.GetApiContainerStatus()
	if enclaveApiContainerStatus != kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING {
		return nil, stacktrace.NewError(
			"Enclave API container status was '%v', but we can't create an enclave context without a running API container",
			enclaveApiContainerStatus,
		)
	}

	apiContainerInfo := enclaveInfo.GetApiContainerInfo()
	if apiContainerInfo == nil {
		return nil, stacktrace.NewError("API container was listed as running, but no API container info exists")
	}
	apiContainerHostMachineInfo := enclaveInfo.GetApiContainerHostMachineInfo()
	if apiContainerHostMachineInfo == nil {
		return nil, stacktrace.NewError("API container was listed as running, but no API container host machine info exists")
	}

	apiContainerHostMachineUrl := fmt.Sprintf(
		"%v:%v",
		apiContainerHostMachineInfo.IpOnHostMachine,
		apiContainerHostMachineInfo.GrpcPortOnHostMachine,
	)
	// TODO SECURITY: use HTTPS!
	apiContainerConn, err := grpc.Dial(apiContainerHostMachineUrl, grpc.WithInsecure())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred connecting to the API container on host machine URL '%v'", apiContainerHostMachineUrl)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(apiContainerConn)

	result := enclaves.NewEnclaveContext(
		apiContainerClient,
		enclaves.EnclaveID(enclaveInfo.EnclaveId),
	)

	return result, nil
}

func validateEngineApiVersion(ctx context.Context, engineServiceClient kurtosis_engine_rpc_api_bindings.EngineServiceClient) error {
	getEngineInfoResponse, err := engineServiceClient.GetEngineInfo(ctx, &emptypb.Empty{})
	if err != nil {
		errorStr := "An error occurred getting engine info"
		grpcErrorCode := status.Code(err)
		if grpcErrorCode == codes.Unavailable {
			errorStr = "The Kurtosis Engine Server is unavailable and is probably not running; you will need to start it using the Kurtosis CLI before you can create a connection to it"
		}
		return stacktrace.Propagate(err, errorStr)
	}
	runningEngineVersionStr := getEngineInfoResponse.GetEngineVersion()

	runningEngineSemver, err := semver.StrictNewVersion(runningEngineVersionStr)
	if err != nil {
		logrus.Warnf("We expected the running engine version to match format X.Y.Z, but instead got '%v'; "+
			"this means that we can't verify the API library and engine versions match so you may encounter runtime errors", runningEngineVersionStr)
		return nil
	}

	libraryEngineSemver, err := semver.StrictNewVersion(kurtosis_version.KurtosisVersion)
	if err != nil {
		logrus.Warnf("We expected the API library version to match format X.Y.Z, but instead got '%v'; "+
			"this means that we can't verify the API library and engine versions match so you may encounter runtime errors", kurtosis_version.KurtosisVersion)
		return nil
	}

	runningEngineMajorVersion := runningEngineSemver.Major()
	runningEngineMinorVersion := runningEngineSemver.Minor()

	libraryEngineMajorVersion := libraryEngineSemver.Major()
	libraryEngineMinorVersion := libraryEngineSemver.Minor()

	doApiVersionsMatch := libraryEngineMajorVersion == runningEngineMajorVersion && libraryEngineMinorVersion == runningEngineMinorVersion

	if !doApiVersionsMatch {
		return stacktrace.NewError(
			"An API version mismatch was detected between the running engine version '%v' and the engine version this Kurtosis SDK library expects, '%v'. You should:\n"+
				"  1) upgrade your Kurtosis CLI to latest using the instructions at https://docs.kurtosistech.com/installation.html\n"+
				"  2) use the Kurtosis CLI to restart your engine via 'kurtosis engine restart'\n"+
				"  3) upgrade your Kurtosis SDK library using the instructions at https://github.com/kurtosis-tech/kurtosis-engine-api-lib\n",
			runningEngineSemver.String(),
			libraryEngineSemver.String(),
		)
	}

	return nil
}

func newGetUserServiceLogsArgs(
	enclaveID enclaves.EnclaveID,
	userServiceGUIDs map[services.ServiceGUID]bool,
) *kurtosis_engine_rpc_api_bindings.GetUserServiceLogsArgs {
	userServiceGUIDStrSet := make(map[string]bool, len(userServiceGUIDs))

	for userServiceGUID, isUserServiceInSet := range userServiceGUIDs {
		userServiceGUIDStr := string(userServiceGUID)
		userServiceGUIDStrSet[userServiceGUIDStr] = isUserServiceInSet
	}

	getUserServiceLogsArgs := &kurtosis_engine_rpc_api_bindings.GetUserServiceLogsArgs{
		EnclaveId:      string(enclaveID),
		ServiceGuidSet: userServiceGUIDStrSet,
	}

	return getUserServiceLogsArgs
}

func newUserServiceReadCloserLogsByServiceGuid(
	userServiceGUIDs map[services.ServiceGUID]bool,
) map[services.ServiceGUID]io.ReadCloser {
	userServiceLogsByUserServiceGUID := map[services.ServiceGUID]io.ReadCloser{}

	for userServiceGUID := range userServiceGUIDs {
		userServiceReadCloser := user_service_logs_read_closer.NewUserServiceLogsReadCloser()
		userServiceLogsByUserServiceGUID[userServiceGUID] = userServiceReadCloser
	}
	return userServiceLogsByUserServiceGUID
}

func streamNewUserServiceLogLines(
	userServiceGuid services.ServiceGUID,
	userServiceReadCloserLogsByServiceGuid map[services.ServiceGUID]io.ReadCloser,
	userServiceLogLines *kurtosis_engine_rpc_api_bindings.LogLine,
) error {

	readCloser, found := userServiceReadCloserLogsByServiceGuid[userServiceGuid]
	if !found {
		return stacktrace.NewError("Expected to find a read closer logs for user service with GUID '%v' on user service logs map '%+v' but was not found; this should never happen, and is a bug in Kurtosis", userServiceGuid, userServiceReadCloserLogsByServiceGuid)
	}
	userServiceReadCloser, ok := readCloser.(*user_service_logs_read_closer.UserServiceLogsReadCloser)
	if !ok {
		return stacktrace.NewError("An error occurred downcasting user-service-logs-read-closer '%+v'; this should never happen, and is a bug with Kurtosis!", readCloser)
	}
	for _, userServiceLogLineStr := range userServiceLogLines.Line {
		userServiceReadCloser.AddLine(userServiceLogLineStr)
	}
	return nil
}

func closeAllUserServiceReadCloserLogs(userServiceReadCloserLogsByServiceGuid map[services.ServiceGUID]io.ReadCloser) {
	for userServiceGuid, userServiceReadCloser := range userServiceReadCloserLogsByServiceGuid {
		if err := userServiceReadCloser.Close(); err != nil {
			logrus.Errorf("Streaming user service logs has finished, so we tried to close user service logs read closer for service with GUID '%v', but an error was thrown:\n%v", userServiceGuid, err)
		}
	}
}
