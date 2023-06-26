package enclave_manager

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/kurtosis_engine_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/uuid_generator"
	"github.com/kurtosis-tech/kurtosis/core/launcher/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/name_generator"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
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
	isPartitioningEnabled bool,
) (*kurtosis_engine_rpc_api_bindings.EnclaveInfo, error) {

	uuid, err := uuid_generator.GenerateUUIDString()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred while creating UUID for enclave with supplied name '%v'", enclaveName)
	}
	enclaveUuid := enclave.EnclaveUUID(uuid)

	allCurrentEnclaves, err := creator.kurtosisBackend.GetEnclaves(setupCtx, getAllEnclavesFilter())
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred checking for enclaves with name '%v'", enclaveName)
	}

	if enclaveName == autogenerateEnclaveNameKeyword {
		enclaveName = GetRandomEnclaveNameWithRetries(name_generator.GenerateNatureThemeNameForEnclave, allCurrentEnclaves, getRandomEnclaveIdRetries)
	}

	if isEnclaveNameInUse(enclaveName, allCurrentEnclaves) {
		return nil, stacktrace.NewError("Cannot create enclave '%v' because an enclave with that name already exists", enclaveName)
	}

	if err := validateEnclaveName(enclaveName); err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred validating enclave name '%v'", enclaveName)
	}

	teardownCtx := context.Background() // Separate context for tearing stuff down in case the input context is cancelled
	// Create Enclave with kurtosisBackend
	newEnclave, err := creator.kurtosisBackend.CreateEnclave(setupCtx, enclaveUuid, enclaveName, isPartitioningEnabled)
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

	apiContainer, err := creator.launchApiContainer(setupCtx,
		apiContainerImageVersionTag,
		apiContainerLogLevel,
		enclaveUuid,
		apiContainerListenGrpcPortNumInsideNetwork,
		isPartitioningEnabled,
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

	var apiContainerHostMachineInfo *kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo
	if apiContainer.GetPublicIPAddress() != nil &&
		apiContainer.GetPublicGRPCPort() != nil {

		apiContainerHostMachineInfo = &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerHostMachineInfo{
			IpOnHostMachine:       apiContainer.GetPublicIPAddress().String(),
			GrpcPortOnHostMachine: uint32(apiContainer.GetPublicGRPCPort().GetNumber()),
		}
	}

	creationTimestamp := getEnclaveCreationTimestamp(newEnclave)
	newEnclaveUuid := newEnclave.GetUUID()
	newEnclaveUuidStr := string(newEnclaveUuid)
	shortenedUuid := uuid_generator.ShortenedUUIDString(newEnclaveUuidStr)

	result := &kurtosis_engine_rpc_api_bindings.EnclaveInfo{
		EnclaveUuid:        newEnclaveUuidStr,
		Name:               newEnclave.GetName(),
		ShortenedUuid:      shortenedUuid,
		ContainersStatus:   kurtosis_engine_rpc_api_bindings.EnclaveContainersStatus_EnclaveContainersStatus_RUNNING,
		ApiContainerStatus: kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerStatus_EnclaveAPIContainerStatus_RUNNING,
		ApiContainerInfo: &kurtosis_engine_rpc_api_bindings.EnclaveAPIContainerInfo{
			ContainerId:           "",
			IpInsideEnclave:       apiContainer.GetPrivateIPAddress().String(),
			GrpcPortInsideEnclave: uint32(apiContainerListenGrpcPortNumInsideNetwork),
		},
		ApiContainerHostMachineInfo: apiContainerHostMachineInfo,
		CreationTime:                creationTimestamp,
	}

	// Everything started successfully, so the responsibility of deleting the enclave is now transferred to the caller
	shouldDestroyEnclave = false
	shouldStopApiContainer = false
	return result, nil
}

func (creator *EnclaveCreator) launchApiContainer(
	ctx context.Context,
	apiContainerImageVersionTag string,
	logLevel logrus.Level,
	enclaveUuid enclave.EnclaveUUID,
	grpcListenPort uint16,
	isPartitioningEnabled bool,
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
			isPartitioningEnabled,
			creator.apiContainerKurtosisBackendConfigSupplier,
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
		isPartitioningEnabled,
		creator.apiContainerKurtosisBackendConfigSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Expected to be able to launch api container for enclave '%v' with the default version, but an error occurred", enclaveUuid)
	}
	return apiContainer, nil
}
