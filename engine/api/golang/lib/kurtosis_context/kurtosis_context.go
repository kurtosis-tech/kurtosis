package kurtosis_context

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/kurtosis_api_version_const"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis-engine-server/api/golang/kurtosis_engine_rpc_api_consts"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	localHostIPAddressStr = "0.0.0.0"

	// TODO even the org-and-repo should come from Kurt Core
	apiContainerImage = "kurtosistech/kurtosis-core_api:" + kurtosis_api_version_const.KurtosisApiVersion

	shouldPublishAllPorts = true
)
var apiContainerLogLevel = logrus.InfoLevel

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
type KurtosisContext struct {
	client kurtosis_engine_rpc_api_bindings.EngineServiceClient
}

// NewKurtosisContextFromLocalEngine
// Attempts to create a KurtosisContext connected to a Kurtosis engine running locally
func NewKurtosisContextFromLocalEngine() (*KurtosisContext, error) {
	kurtosisEngineSocketStr := fmt.Sprintf("%v:%v", localHostIPAddressStr, kurtosis_engine_rpc_api_consts.ListenPort)

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

	kurtosisContext := &KurtosisContext{
		client: engineServiceClient,
	}

	return kurtosisContext, nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
func (kurtosisCtx *KurtosisContext) CreateEnclave(
	ctx context.Context,
	enclaveId enclaves.EnclaveID,
	isPartitioningEnabled bool,
) (*enclaves.EnclaveContext, error) {

	createEnclaveArgs := &kurtosis_engine_rpc_api_bindings.CreateEnclaveArgs{
		EnclaveId: string(enclaveId),
		ApiContainerImage: apiContainerImage,
		ApiContainerLogLevel: apiContainerLogLevel.String(),
		IsPartitioningEnabled: isPartitioningEnabled,
		ShouldPublishAllPorts: shouldPublishAllPorts,
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
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

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
func (kurtosisCtx *KurtosisContext) StopEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	stopEnclaveArgs := &kurtosis_engine_rpc_api_bindings.StopEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.StopEnclave(ctx, stopEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred stopping enclave with ID '%v'", enclaveId)
	}

	return nil
}

// Docs available at https://docs.kurtosistech.com/kurtosis-engine-api-lib/lib-documentation
func (kurtosisCtx *KurtosisContext) DestroyEnclave(ctx context.Context, enclaveId enclaves.EnclaveID) error {
	destroyEnclaveArgs := &kurtosis_engine_rpc_api_bindings.DestroyEnclaveArgs{
		EnclaveId: string(enclaveId),
	}

	if _, err := kurtosisCtx.client.DestroyEnclave(ctx, destroyEnclaveArgs); err != nil {
		return stacktrace.Propagate(err, "An error occurred destroying enclave with ID '%v'", enclaveId)
	}

	return nil
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
