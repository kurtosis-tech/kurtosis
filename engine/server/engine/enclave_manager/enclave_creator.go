package enclave_manager

import (
	"context"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/engine/server/engine/types"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	defaultHttpLogsCollectorPortNum = uint16(9712)
	defaultTcpLogsCollectorPortNum  = uint16(9713)
)

type EnclaveCreator struct {
	kurtosisBackend                           backend_interface.KurtosisBackend
	apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier
}

func newEnclaveCreator(
	kurtosisBackend backend_interface.KurtosisBackend,
	apiContainerKurtosisBackendConfigSupplier api_container_launcher.KurtosisBackendConfigSupplier,
) *EnclaveCreator {

	return &EnclaveCreator{
		kurtosisBackend: kurtosisBackend,
		apiContainerKurtosisBackendConfigSupplier: apiContainerKurtosisBackendConfigSupplier,
	}
}

func (creator *EnclaveCreator) CreateEnclave(
	setupCtx context.Context,
	// If blank, will use the default
	apiContainerImageVersionTag string,
	apiContainerLogLevel logrus.Level,
	//If blank, will use a random one
	enclaveName string,
	enclaveEnvVars string,
	isProduction bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
	kurtosisBackendType args.KurtosisBackendType,
	shouldAPICRunInDebugMode bool,
) (*types.EnclaveInfo, error) {

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating UUID for enclave with supplied name '%v'", enclaveName)
	}
	enclaveUuid := enclave.EnclaveUUID(uuid)

	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled
	// Create Enclave with kurtosisBackend

	newEnclave, err := creator.kurtosisBackend.CreateEnclave(setupCtx, enclaveUuid, enclaveName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating enclave with name `%v` and uuid '%v'", enclaveName, enclaveUuid)
	}
	shouldDestroyEnclave := true
	defer func() {
		if shouldDestroyEnclave {
			_, destroyEnclaveErrs, err := creator.kurtosisBackend.DestroyEnclaves(teardownCtx, getEnclaveByEnclaveIdFilter(enclaveUuid))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the backend and destroy enclave '%v', but an error occurred:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
				return
			}
			for enclaveUuid, err := range destroyEnclaveErrs {
				logrus.Errorf("Expected to be able to cleanup the enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
			}
		}
	}()

	// only create log collector for backend as
	shouldDeleteLogsCollector := true
	if kurtosisBackendType == args.KurtosisBackendType_Docker {
		// TODO the logs collector has a random private ip address in the enclave network that must be tracked
		if _, err := creator.kurtosisBackend.CreateLogsCollectorForEnclave(setupCtx, enclaveUuid, defaultTcpLogsCollectorPortNum, defaultHttpLogsCollectorPortNum); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating the logs collector with TCP port number '%v' and HTTP port number '%v'", defaultTcpLogsCollectorPortNum, defaultHttpLogsCollectorPortNum)
		}
		defer func() {
			if shouldDeleteLogsCollector {
				err = creator.kurtosisBackend.DestroyLogsCollectorForEnclave(teardownCtx, enclaveUuid)
				if err != nil {
					logrus.Errorf("Couldn't cleanup logs collector for enclave '%v' as the following error was thrown:\n%v", enclaveUuid, err)
				}
			}
		}()
	}

	apiContainer, err := creator.launchApiContainer(setupCtx,
		apiContainerImageVersionTag,
		apiContainerLogLevel,
		enclaveUuid,
		apiContainerListenGrpcPortNumInsideNetwork,
		enclaveEnvVars,
		isProduction,
		metricsUserID,
		didUserAcceptSendingMetrics,
		isCI,
		cloudUserID,
		cloudInstanceID,
		shouldAPICRunInDebugMode,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}
	shouldStopApiContainer := true
	defer func() {
		if shouldStopApiContainer {
			_, destroyApiContainerErrs, err := creator.kurtosisBackend.DestroyAPIContainers(teardownCtx, getApiContainerByEnclaveIdFilter(enclaveUuid))
			manualActionRequiredStrFmt := "ACTION REQUIRED: You'll need to manually destroy the API Container for enclave '%v'!!!!!!"
			if err != nil {
				logrus.Errorf("Expected to be able to call the backend and destroy the API container for enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
				return
			}
			for enclaveUuid, err := range destroyApiContainerErrs {
				logrus.Errorf("Expected to be able to cleanup the API Container in enclave '%v', but an error was thrown:\n%v", enclaveUuid, err)
				logrus.Errorf(manualActionRequiredStrFmt, enclaveUuid)
			}
		}
	}()

	var apiContainerHostMachineInfo *types.EnclaveAPIContainerHostMachineInfo
	if apiContainer.GetPublicIPAddress() != nil &&
		apiContainer.GetPublicGRPCPort() != nil {

		apiContainerHostMachineInfo = &types.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:       apiContainer.GetPublicIPAddress().String(),
			GrpcPortOnHostMachine: uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
		}
	}

	creationTimestamp, err := getEnclaveCreationTimestamp(newEnclave)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the creation timestamp for enclave with UUID '%v'", newEnclave.GetUUID())
	}
	newEnclaveUuid := newEnclave.GetUUID()
	newEnclaveUuidStr := string(newEnclaveUuid)
	shortenedUuid := uuid_generator.ShortenedUUIDString(newEnclaveUuidStr)

	bridgeIpAddr := ""
	if apiContainer.GetBridgeNetworkIPAddress() != nil {
		bridgeIpAddr = apiContainer.GetBridgeNetworkIPAddress().String()
	}

	mode := types.EnclaveMode_TEST
	if newEnclave.IsProductionEnclave() {
		mode = types.EnclaveMode_PRODUCTION
	}

	newEnclaveInfo := &types.EnclaveInfo{
		EnclaveUuid:        newEnclaveUuidStr,
		Name:               newEnclave.GetName(),
		ShortenedUuid:      shortenedUuid,
		EnclaveStatus:      types.EnclaveStatus_RUNNING,
		ApiContainerStatus: types.ContainerStatus_RUNNING,
		ApiContainerInfo: &types.EnclaveAPIContainerInfo{
			ContainerId:           "",
			IpInsideEnclave:       apiContainer.GetPrivateIPAddress().String(),
			GrpcPortInsideEnclave: uint32(apiContainerListenGrpcPortNumInsideNetwork),
			BridgeIpAddress:       bridgeIpAddr,
		},
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		CreationTime:                *creationTimestamp,
		Mode:                        mode,
	}

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldStopApiContainer = false
	shouldDeleteLogsCollector = false
	shouldDestroyEnclave = false
	return newEnclaveInfo, nil
}

func (creator *EnclaveCreator) launchApiContainer(
	ctx context.Context,
	apiContainerImageVersionTag string,
	logLevel logrus.Level,
	enclaveUuid enclave.EnclaveUUID,
	grpcListenPort uint16,
	enclaveEnvVars string,
	isProduction bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
	shouldStartInDebugMode bool,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {
	apiContainerLauncher := api_container_launcher.NewApiContainerLauncher(
		creator.kurtosisBackend,
	)
	if apiContainerImageVersionTag != "" {
		apiContainer, err := apiContainerLauncher.LaunchWithCustomVersion(
			ctx,
			apiContainerImageVersionTag,
			logLevel,
			enclaveUuid,
			grpcListenPort,
			creator.apiContainerKurtosisBackendConfigSupplier,
			enclaveEnvVars,
			isProduction,
			metricsUserID,
			didUserAcceptSendingMetrics,
			isCI,
			cloudUserID,
			cloudInstanceID,
			shouldStartInDebugMode,
		)
		if err != nil {
			return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with custom version '%v', but an error occurred", enclaveUuid, apiContainerImageVersionTag)
		}
		return apiContainer, nil
	}
	apiContainer, err := apiContainerLauncher.LaunchWithDefaultVersion(
		ctx,
		logLevel,
		enclaveUuid,
		grpcListenPort,
		creator.apiContainerKurtosisBackendConfigSupplier,
		enclaveEnvVars,
		isProduction,
		metricsUserID,
		didUserAcceptSendingMetrics,
		isCI,
		cloudUserID,
		cloudInstanceID,
		shouldStartInDebugMode,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with the default version, but an error occurred", enclaveUuid)
	}
	return apiContainer, nil
}
