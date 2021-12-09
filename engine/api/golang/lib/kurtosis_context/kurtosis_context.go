package kurtosis_context

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/Masterminds/semver/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// NOTE: This needs to be 127.0.0.1 rather than 0.0.0.0, because Windows machines don't translate 0.0.0.0 -> 127.0.0.1
	localHostIPAddressStr = "127.0.0.1"

	shouldPublishAllPorts = true

	DefaultKurtosisEngineServerPortNum = uint16(9710)

	// Blank tells the engine server to use the default
	defaultApiContainerVersionTag = ""
)

var apiContainerLogLevel = logrus.InfoLevel

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
type KurtosisContext struct {
	client kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// NewKurtosisContextFromLocalEngine
// Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
func NewKurtosisContextFromLocalEngine() (*KurtosisContext, error) {
	ctx := context.Background()
	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, DefaultKurtosisEngineServerPortNum)

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

	getEngineInfoResponse, err := engineServiceClient.GetEngineInfo(ctx, &emptypb.Empty{})
	if err != nil {
		errorStr := "An error occurred getting engine info"
		grpcErrorCode := status.Code(err)
		if grpcErrorCode == codes.Unavailable {
			errorStr = "The Kurtosis Engine Server is unavailable, probably it is not running, you should start it before executing any request"
		}
		return nil, stacktrace.Propagate(err, errorStr)
	}
	runningEngineVersionStr := getEngineInfoResponse.GetEngineVersion()

	runningEngineSemver, err := semver.StrictNewVersion(runningEngineVersionStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing running engine version string '%v' to semantic version", runningEngineVersionStr)
	}

	runningEngineMajorVersion := runningEngineSemver.Major()
	runningEngineMinorVersion := runningEngineSemver.Minor()

	libraryEngineSemver, err := semver.StrictNewVersion(kurtosis_engine_version.KurtosisEngineVersion)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing library engine version string '%v' to semantic version", kurtosis_engine_version.KurtosisEngineVersion)
	}
	libraryEngineMajorVersion := libraryEngineSemver.Major()
	libraryEngineMinorVersion := libraryEngineSemver.Minor()

	doApiVersionsMatch := libraryEngineMajorVersion == runningEngineMajorVersion && libraryEngineMinorVersion == runningEngineMinorVersion
	if !doApiVersionsMatch {
		return nil, stacktrace.NewError(
			"An API version mismatch was detected between the running engine version '%v' and the engine version the library expects, '%v'",
			runningEngineSemver.String(),
			libraryEngineSemver.String(),
		)
	}

	kurtosisContext := &KurtosisContext{
		client: engineServiceClient,
	}

	return kurtosisContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
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
		ShouldPublishAllPorts:  shouldPublishAllPorts,
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
func (kurtosisCtx *KurtosisContext) StopEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	stopEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.StopEnclave(ctx, stopEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave with ID '%v'", enclaveId)
	}

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
func (kurtosisCtx *KurtosisContext) DestroyEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave with ID '%v'", enclaveId)
	}

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-server/lib-documentation
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

// ====================================================================================================
// 									   Private helper methods
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
		apiContainerHostMachineInfo.PortOnHostMachine,
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
		enclaveInfo.EnclaveDataDirpathOnHostMachine,
	)

	return result, nil
}
