package kurtosis_context

import (
	"context"
	"fmt"
	"io"

	"github.com/Masterminds/semver/v3"
	portal_constructors "github.com/kurtosis-tech/kurtosis-portal/api/golang/constructors"
	portal_api "github.com/kurtosis-tech/kurtosis-portal/api/golang/generated"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/contexts-config-store/store"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// NOTE: This needs to be 127.0.0.1 rather than 0.0.0.0, because Windows machines don't translate 0.0.0.0 -> 127.0.0.1
	localHostIPAddressStr = "127.0.0.1"

	DefaultGrpcEngineServerPortNum = uint16(9710)

	// Blank tells the engine server to use the default
	defaultApiContainerVersionTagStr = ""

	serviceLogsStreamContentChanBufferSize = 5

	grpcStreamCancelContextErrorMessage = "rpc error: code = Canceled desc = context canceled"

	validUuidMatchesAllowed = 1

	portalIsRequired = true

	defaultShouldAPICRunInDebugMode = false
	runAPICInDebugMode              = true
)

var (
	defaultApiContainerLogLevelStr = logrus.DebugLevel.String()

	apicPortTransportProtocol = portal_api.TransportProtocol_TCP

	EnginePortTransportProtocol = portal_api.TransportProtocol_TCP

	ApicRemoteEndpointType   = portal_api.RemoteEndpointType_Apic
	EngineRemoteEndpointType = portal_api.RemoteEndpointType_Engine
	UserServiceEndpointType  = portal_api.RemoteEndpointType_UserService

	ForwardPortWaitUntilReady      = true
	ForwardPortDoNotWaitUntilReady = false
)

// Docs available at https://docs.kurtosis.com/sdk#kurtosiscontext
type KurtosisContext struct {
	engineClient kurtosis_engine_rpc_api_bindings.EngineServiceClient
	portalClient portal_api.KurtosisPortalClientClient
}

// NewKurtosisContextFromLocalEngine
// Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
func NewKurtosisContextFromLocalEngine() (*KurtosisContext, error) {
	ctx := context.Background()
	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, DefaultGrpcEngineServerPortNum)

	// TODO SECURITY: Use HTTPS to ensure we're connecting to the real Kurtosis API servers
	conn, err := grpc.Dial(kurtosisEngineSocketStr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred creating a connection to the Kurtosis Engine Server at '%v'",
			kurtosisEngineSocketStr,
		)
	}

	engineServiceClient := kurtosis_engine_rpc_api_bindings.NewEngineServiceClient(conn)
	if err = validateEngineApiVersion(ctx, engineServiceClient); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating the Kurtosis engine API version")
	}

	var portalClient portal_api.KurtosisPortalClientClient
	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err == nil {
		if store.IsRemote(currentContext) {
			portalClient, err = CreatePortalDaemonClient(portalIsRequired)
			if err != nil {
				return nil, stacktrace.Propagate(err, "Error building client for Kurtosis Portal daemon")
			}
		}
	} else {
		logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
	}

	kurtosisContext := &KurtosisContext{
		engineClient: engineServiceClient,
		portalClient: portalClient,
	}

	return kurtosisContext, nil
}

// Docs available at https://docs.kurtosis.com/sdk#createenclaveenclaveid-enclaveid-boolean-issubnetworkingenabled---enclavecontextenclavecontext-enclavecontext
func (kurtosisCtx *KurtosisContext) CreateEnclave(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := newCreateEnclaveArgsWithDefaultValues(enclaveName)

	response, err := kurtosisCtx.engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave with name '%v'", enclaveName)
	}

	enclaveContext, err := newEnclaveContextFromEnclaveInfo(ctx, kurtosisCtx.portalClient, response.EnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from a newly-created enclave; this should never happen")
	}

	return enclaveContext, nil
}

func (kurtosisCtx *KurtosisContext) CreateEnclaveWithDebugEnabled(
	ctx context.Context,
	enclaveName string,
) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := newCreateEnclaveArgsWithDefaultValuesForDebugging(enclaveName)

	response, err := kurtosisCtx.engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave with name '%v'", enclaveName)
	}

	enclaveContext, err := newEnclaveContextFromEnclaveInfo(ctx, kurtosisCtx.portalClient, response.EnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from a newly-created enclave; this should never happen")
	}

	return enclaveContext, nil
}

// Docs available at https://docs.kurtosis.com/sdk#createenclaveenclaveid-enclaveid-boolean-issubnetworkingenabled---enclavecontextenclavecontext-enclavecontext
func (kurtosisCtx *KurtosisContext) CreateProductionEnclave(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := newCreateProductionEnclaveWithDefaultValues(enclaveName)

	response, err := kurtosisCtx.engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave with name '%v'", enclaveName)
	}

	enclaveContext, err := newEnclaveContextFromEnclaveInfo(ctx, kurtosisCtx.portalClient, response.EnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from a newly-created enclave; this should never happen")
	}

	return enclaveContext, nil
}

func (kurtosisCtx *KurtosisContext) CreateProductionEnclaveWithDebugEnabled(ctx context.Context, enclaveName string) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := newCreateProductionEnclaveWithDefaultValuesForDebugging(enclaveName)

	response, err := kurtosisCtx.engineClient.CreateEnclave(ctx, createEnclaveArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave with name '%v'", enclaveName)
	}

	enclaveContext, err := newEnclaveContextFromEnclaveInfo(ctx, kurtosisCtx.portalClient, response.EnclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from a newly-created enclave; this should never happen")
	}

	return enclaveContext, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#getenclavecontextstring-enclaveidentifier---enclavecontextenclavecontext-enclavecontext
func (kurtosisCtx *KurtosisContext) GetEnclaveContext(ctx context.Context, enclaveIdentifier string) (*enclaves.EnclaveContext, error) {
	enclaveInfo, err := kurtosisCtx.GetEnclave(ctx, enclaveIdentifier)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while getting enclave with identifier '%v'", enclaveIdentifier)
	}

	enclaveCtx, err := newEnclaveContextFromEnclaveInfo(ctx, kurtosisCtx.portalClient, enclaveInfo)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating an enclave context from the returned enclave info")
	}

	return enclaveCtx, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getenclaves---enclaves-enclaves
func (kurtosisCtx *KurtosisContext) GetEnclaves(ctx context.Context) (*Enclaves, error) {
	response, err := kurtosisCtx.engineClient.GetEnclaves(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting enclaves",
		)
	}

	enclavesByUuid := map[string]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	enclavesByName := map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	enclavesByShortenedUuid := map[string][]*kurtosis_engine_rpc_api_bindings.EnclaveInfo{}
	for enclaveUuid, enclaveInfo := range response.EnclaveInfo {
		enclavesByUuid[enclaveUuid] = response.EnclaveInfo[enclaveUuid]
		enclavesByName[enclaveInfo.Name] = append(enclavesByShortenedUuid[enclaveInfo.GetName()], response.EnclaveInfo[enclaveUuid])
		enclavesByShortenedUuid[enclaveInfo.ShortenedUuid] = append(enclavesByShortenedUuid[enclaveInfo.ShortenedUuid], response.EnclaveInfo[enclaveUuid])
	}

	return &Enclaves{
		enclavesByUuid:          enclavesByUuid,
		enclavesByName:          enclavesByName,
		enclavesByShortenedUuid: enclavesByShortenedUuid,
	}, nil
}

// Docs available at https://docs.kurtosis.com/sdk/#getenclavestring-enclaveidentifier---enclaveinfo-enclaveinfo
func (kurtosisCtx *KurtosisContext) GetEnclave(ctx context.Context, enclaveIdentifier string) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {
	enclaves, err := kurtosisCtx.GetEnclaves(ctx)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred getting enclave for identifier '%v'",
			enclaveIdentifier,
		)
	}

	if enclaveInfo, found := enclaves.enclavesByUuid[enclaveIdentifier]; found {
		return enclaveInfo, nil
	}

	if enclaveInfos, found := enclaves.enclavesByShortenedUuid[enclaveIdentifier]; found {
		if len(enclaveInfos) == validUuidMatchesAllowed {
			return enclaveInfos[0], nil
		} else if len(enclaveInfos) > validUuidMatchesAllowed {
			return nil, stacktrace.NewError("Found multiple enclaves '%v' matching shortened uuid '%v'. Please use a uuid to be more specific", enclaveInfos, enclaveIdentifier)
		}
	}

	if enclaveInfos, found := enclaves.enclavesByName[enclaveIdentifier]; found {
		if len(enclaveInfos) == validUuidMatchesAllowed {
			return enclaveInfos[0], nil
		} else if len(enclaveInfos) > validUuidMatchesAllowed {
			return nil, stacktrace.NewError("Found multiple enclaves '%v' matching name '%v'. Please use a uuid to be more specific", enclaveInfos, enclaveIdentifier)
		}
	}

	return nil, stacktrace.NewError("Couldn't find an enclave for identifier '%v'", enclaveIdentifier)
}

// Docs available at https://docs.kurtosis.com/sdk/#stopenclavestring-enclaveidentifier
func (kurtosisCtx *KurtosisContext) StopEnclave(ctx context.Context, enclaveIdentifier string) error {
	stopEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
		EnclaveIdentifier: enclaveIdentifier,
	}

	if _, err := kurtosisCtx.engineClient.StopEnclave(ctx, stopEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave with identifier '%v'", enclaveIdentifier)
	}

	return nil
}

// Docs available at https://docs.kurtosis.com/sdk/#destroyenclavestring-enclaveidentifier
func (kurtosisCtx *KurtosisContext) DestroyEnclave(ctx context.Context, enclaveIdentifier string) error {
	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{
		EnclaveIdentifier: enclaveIdentifier,
	}

	if _, err := kurtosisCtx.engineClient.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave with identifier '%v'", enclaveIdentifier)
	}

	return nil
}

// Docs available at https://docs.kurtosis.com/sdk#cleanboolean-shouldcleanall---enclavenameanduuid-removedenclavenameanduuids
func (kurtosisCtx *KurtosisContext) Clean(ctx context.Context, shouldCleanAll bool) ([]*kurtosis_engine_rpc_api_bindings.EnclaveNameAndUuid, error) {
	cleanArgs := &kurtosis_engine_rpc_api_bindings.CleanArgs{
		ShouldCleanAll: &shouldCleanAll,
	}
	cleanResponse, err := kurtosisCtx.engineClient.Clean(ctx, cleanArgs)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred when trying to perform a clean with the clean-all arg set to '%v'", shouldCleanAll)
	}

	return cleanResponse.RemovedEnclaveNameAndUuids, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getservicelogsstring-enclaveidentifier-setserviceuuid-serviceuuids-boolean-shouldfollowlogs-loglinefilter-loglinefilter---servicelogsstreamcontent-servicelogsstreamcontent
func (kurtosisCtx *KurtosisContext) GetServiceLogs(
	ctx context.Context,
	enclaveIdentifier string,
	userServiceUuids map[services.ServiceUUID]bool,
	shouldFollowLogs bool,
	shouldReturnAllLogs bool,
	numLogLines uint32,
	logLineFilter *LogLineFilter,
) (
	chan *serviceLogsStreamContent,
	func(),
	error,
) {

	ctxWithCancel, cancelCtxFunc := context.WithCancel(ctx)
	shouldCancelCtx := true
	defer func() {
		if shouldCancelCtx {
			cancelCtxFunc()
		}
	}()

	//this is a buffer channel for the case that users could be consuming this channel in a process and
	//this process could take much time until the next channel pull, so we could be filling the buffer during that time to not let the servers thread idled
	serviceLogsStreamContentChan := make(chan *serviceLogsStreamContent, serviceLogsStreamContentChanBufferSize)

	getServiceLogsArgs, err := newGetServiceLogsArgs(enclaveIdentifier, userServiceUuids, shouldFollowLogs, shouldReturnAllLogs, numLogLines, logLineFilter)
	if err != nil {
		return nil, nil, stacktrace.Propagate(
			err,
			"An error occurred creating the service logs arguments with enclave identifier '%v', user service UUID '%+v', should follow logs value '%v' and with these conjunctive log line filters '%+v'",
			enclaveIdentifier,
			userServiceUuids,
			shouldFollowLogs,
			logLineFilter,
		)
	}

	stream, err := kurtosisCtx.engineClient.GetServiceLogs(ctxWithCancel, getServiceLogsArgs)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred streaming service logs using args '%+v'", getServiceLogsArgs)
	}

	go runReceiveStreamLogsFromTheServerRoutine(
		cancelCtxFunc,
		enclaveIdentifier,
		userServiceUuids,
		serviceLogsStreamContentChan,
		stream,
	)

	//This is an async operation, so we don't want to cancel the context if the connection is established and data is flowing
	shouldCancelCtx = false
	return serviceLogsStreamContentChan, cancelCtxFunc, nil
}

// Docs available at https://docs.kurtosis.com/sdk#getexistingandhistoricalenclaveidentifiers---enclaveidentifiers-enclaveidentifiers
func (kurtosisCtx *KurtosisContext) GetExistingAndHistoricalEnclaveIdentifiers(ctx context.Context) (*EnclaveIdentifiers, error) {
	historicalEnclaveIdentifiers, err := kurtosisCtx.engineClient.GetExistingAndHistoricalEnclaveIdentifiers(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while fetching existing and historical enclave identifiers")
	}

	return newEnclaveIdentifiers(historicalEnclaveIdentifiers.AllIdentifiers), nil
}

// ====================================================================================================
//
//	Private helper methods
//
// ====================================================================================================
func runReceiveStreamLogsFromTheServerRoutine(
	cancelCtxFunc context.CancelFunc,
	enclaveIdentifier string,
	requestedServiceUuids map[services.ServiceUUID]bool,
	serviceLogsStreamContentChan chan *serviceLogsStreamContent,
	stream kurtosis_engine_rpc_api_bindings.EngineService_GetServiceLogsClient,
) {

	//Closing all the open resources at the end
	defer func() {
		cancelCtxFunc()
		close(serviceLogsStreamContentChan)
	}()

	for {
		//this is a blocking call, and the only way to unblock it from our side is to cancel the context that it was created with
		getServiceLogsResponse, errReceivingStream := stream.Recv()
		//stream ends case
		if errReceivingStream == io.EOF {
			logrus.Debug("Received an 'EOF' error from the user service logs GRPC stream")
			return
		}
		if errReceivingStream != nil {

			//context canceled case
			if errReceivingStream.Error() == grpcStreamCancelContextErrorMessage {
				logrus.Debug("Received a 'context canceled' error from the user service logs GRPC stream")
				return
			}
			//error during stream case
			logrus.Errorf("An error occurred receiving user service logs stream for user services '%+v' in enclave '%v'. Error:\n%v", requestedServiceUuids, enclaveIdentifier, errReceivingStream)
			return
		}

		serviceLogsStreamContentObj := newServiceLogsStreamContentFromGrpcStreamResponse(requestedServiceUuids, getServiceLogsResponse)

		serviceLogsStreamContentChan <- serviceLogsStreamContentObj
	}
}

func newEnclaveContextFromEnclaveInfo(
	ctx context.Context,
	portalClient portal_api.KurtosisPortalClientClient,
	enclaveInfo *kurtosis_engine_rpc_api_bindings.EnclaveInfo,
) (*enclaves.EnclaveContext, error) {
	// for remote contexts, we need to tunnel the APIC port to the local machine
	currentContext, err := store.GetContextsConfigStore().GetCurrentContext()
	if err == nil {
		// for remote contexts, we need to tunnel the APIC port to the local machine
		if store.IsRemote(currentContext) && portalClient != nil {
			apicGrpcPort := enclaveInfo.GetApiContainerHostMachineInfo().GetGrpcPortOnHostMachine()
			forwardApicPortArgs := portal_constructors.NewForwardPortArgs(apicGrpcPort, apicGrpcPort, ApicRemoteEndpointType, &apicPortTransportProtocol, &ForwardPortWaitUntilReady)
			if _, err := portalClient.ForwardPort(ctx, forwardApicPortArgs); err != nil {
				return nil, stacktrace.Propagate(err, "Unable to forward remote API container port to the local machine")
			}
		}
	} else {
		logrus.Warnf("Unable to retrieve current Kurtosis context. This is not critical, it will assume using Kurtosis default context for now.")
	}

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
	apiContainerConn, err := grpc.Dial(apiContainerHostMachineUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred connecting to the API container on host machine URL '%v'", apiContainerHostMachineUrl)
	}
	apiContainerClient := kurtosis_core_rpc_api_bindings.NewApiContainerServiceClient(apiContainerConn)

	result := enclaves.NewEnclaveContext(
		apiContainerClient,
		enclaves.EnclaveUUID(enclaveInfo.EnclaveUuid),
		enclaveInfo.GetName(),
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
				"  1) upgrade your Kurtosis CLI to latest using the instructions at https://docs.kurtosis.com/upgrade\n"+
				"  2) use the Kurtosis CLI to restart your engine via 'kurtosis engine restart'\n"+
				"  3) upgrade your Kurtosis SDK library using the instructions at https://github.com/kurtosis-tech/kurtosis-engine-api-lib\n",
			runningEngineSemver.String(),
			libraryEngineSemver.String(),
		)
	}

	return nil
}

func newGetServiceLogsArgs(
	enclaveIdentifier string,
	userServiceUUIDs map[services.ServiceUUID]bool,
	shouldFollowLogs bool,
	shouldReturnAllLogs bool,
	numLogLines uint32,
	logLineFilter *LogLineFilter,
) (*kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs, error) {
	userServiceUuuidSet := make(map[string]bool, len(userServiceUUIDs))

	for userServiceUUID, isUserServiceInSet := range userServiceUUIDs {
		userServiceUUIDStr := string(userServiceUUID)
		userServiceUuuidSet[userServiceUUIDStr] = isUserServiceInSet
	}

	grpcConjunctiveFilters, err := newGRPCConjunctiveFilters(logLineFilter)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the GRPC conjunctive log line filters '%+v'", logLineFilter)
	}

	getUserServiceLogsArgs := &kurtosis_engine_rpc_api_bindings.GetServiceLogsArgs{
		EnclaveIdentifier:  enclaveIdentifier,
		ServiceUuidSet:     userServiceUuuidSet,
		FollowLogs:         &shouldFollowLogs,
		ConjunctiveFilters: grpcConjunctiveFilters,
		ReturnAllLogs:      &shouldReturnAllLogs,
		NumLogLines:        &numLogLines,
	}

	return getUserServiceLogsArgs, nil
}

// Even though the backend is prepared for receiving a list of conjunctive filters
// We allow users to send only one filter so far, because it covers the current supported use cases
func newGRPCConjunctiveFilters(
	logLineFilter *LogLineFilter,
) ([]*kurtosis_engine_rpc_api_bindings.LogLineFilter, error) {

	grpcLogLineFilters := []*kurtosis_engine_rpc_api_bindings.LogLineFilter{}

	if logLineFilter == nil {
		return grpcLogLineFilters, nil
	}

	var grpcOperator kurtosis_engine_rpc_api_bindings.LogLineOperator

	switch logLineFilter.operator {
	case logLineOperator_DoesContainText:
		grpcOperator = kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_TEXT
	case logLineOperator_DoesNotContainText:
		grpcOperator = kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_TEXT
	case logLineOperator_DoesContainMatchRegex:
		grpcOperator = kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_CONTAIN_MATCH_REGEX
	case logLineOperator_DoesNotContainMatchRegex:
		grpcOperator = kurtosis_engine_rpc_api_bindings.LogLineOperator_LogLineOperator_DOES_NOT_CONTAIN_MATCH_REGEX
	default:
		return nil, stacktrace.NewError("Unrecognized log line filter operator '%v' in filter '%v'; this is a bug in Kurtosis", logLineFilter.operator, logLineFilter)
	}
	grpcLogLineFilter := &kurtosis_engine_rpc_api_bindings.LogLineFilter{
		TextPattern: logLineFilter.textPattern,
		Operator:    grpcOperator,
	}

	grpcLogLineFilters = append(grpcLogLineFilters, grpcLogLineFilter)

	return grpcLogLineFilters, nil
}

func newServiceLogsStreamContentFromGrpcStreamResponse(
	requestedServiceUuids map[services.ServiceUUID]bool,
	getServiceLogResponse *kurtosis_engine_rpc_api_bindings.GetServiceLogsResponse,
) *serviceLogsStreamContent {
	serviceLogsByServiceUuidMap := map[services.ServiceUUID][]*ServiceLog{}

	receivedServiceLogsByServiceUuid := getServiceLogResponse.GetServiceLogsByServiceUuid()

	for serviceUuid := range requestedServiceUuids {
		serviceUuidStr := string(serviceUuid)
		serviceLogs := []*ServiceLog{}
		serviceLogLine, found := receivedServiceLogsByServiceUuid[serviceUuidStr]
		if found {
			for _, logLineContent := range serviceLogLine.Line {
				serviceLog := newServiceLog(logLineContent)
				serviceLogs = append(serviceLogs, serviceLog)
			}
		}
		serviceLogsByServiceUuidMap[serviceUuid] = serviceLogs
	}

	notFoundServiceUuidSet := getServiceLogResponse.NotFoundServiceUuidSet

	notFoundServiceUuids := make(map[services.ServiceUUID]bool, len(notFoundServiceUuidSet))

	for notFoundServiceUuidStr := range notFoundServiceUuidSet {
		notFoundServiceUuid := services.ServiceUUID(notFoundServiceUuidStr)
		notFoundServiceUuids[notFoundServiceUuid] = true
	}

	newServiceLogsStreamContentObj := newServiceLogsStreamContent(serviceLogsByServiceUuidMap, notFoundServiceUuids)

	return newServiceLogsStreamContentObj
}

func newCreateEnclaveArgsWithDefaultValues(enclaveName string) *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs {

	defaultApiContainerVersionTag := defaultApiContainerVersionTagStr
	defaultApiContainerLogLevel := defaultApiContainerLogLevelStr
	defaultEnclaveMode := kurtosis_engine_rpc_api_bindings.EnclaveMode_TEST
	shouldApicRunInDebugMode := defaultShouldAPICRunInDebugMode

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:              &enclaveName,
		ApiContainerVersionTag:   &defaultApiContainerVersionTag,
		ApiContainerLogLevel:     &defaultApiContainerLogLevel,
		Mode:                     &defaultEnclaveMode,
		ShouldApicRunInDebugMode: &shouldApicRunInDebugMode,
	}

	return createEnclaveArgs
}

func newCreateEnclaveArgsWithDefaultValuesForDebugging(enclaveName string) *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs {

	defaultApiContainerVersionTag := defaultApiContainerVersionTagStr
	defaultApiContainerLogLevel := defaultApiContainerLogLevelStr
	defaultEnclaveMode := kurtosis_engine_rpc_api_bindings.EnclaveMode_TEST
	shouldApicRunInDebugMode := runAPICInDebugMode

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:              &enclaveName,
		ApiContainerVersionTag:   &defaultApiContainerVersionTag,
		ApiContainerLogLevel:     &defaultApiContainerLogLevel,
		Mode:                     &defaultEnclaveMode,
		ShouldApicRunInDebugMode: &shouldApicRunInDebugMode,
	}

	return createEnclaveArgs
}

func newCreateProductionEnclaveWithDefaultValues(enclaveName string) *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs {

	defaultApiContainerVersionTag := defaultApiContainerVersionTagStr
	defaultApiContainerLogLevel := defaultApiContainerLogLevelStr
	defaultEnclaveMode := kurtosis_engine_rpc_api_bindings.EnclaveMode_PRODUCTION
	shouldApicRunInDebugMode := defaultShouldAPICRunInDebugMode

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:              &enclaveName,
		ApiContainerVersionTag:   &defaultApiContainerVersionTag,
		ApiContainerLogLevel:     &defaultApiContainerLogLevel,
		Mode:                     &defaultEnclaveMode,
		ShouldApicRunInDebugMode: &shouldApicRunInDebugMode,
	}

	return createEnclaveArgs
}

func newCreateProductionEnclaveWithDefaultValuesForDebugging(enclaveName string) *kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs {

	defaultApiContainerVersionTag := defaultApiContainerVersionTagStr
	defaultApiContainerLogLevel := defaultApiContainerLogLevelStr
	defaultEnclaveMode := kurtosis_engine_rpc_api_bindings.EnclaveMode_PRODUCTION
	shouldApicRunInDebugMode := runAPICInDebugMode

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveName:              &enclaveName,
		ApiContainerVersionTag:   &defaultApiContainerVersionTag,
		ApiContainerLogLevel:     &defaultApiContainerLogLevel,
		Mode:                     &defaultEnclaveMode,
		ShouldApicRunInDebugMode: &shouldApicRunInDebugMode,
	}

	return createEnclaveArgs
}
