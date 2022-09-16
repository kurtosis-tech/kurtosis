/* * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	DefaultVersion = "1.59.2"
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!

	enclaveDataVolumeDirpath = "/kurtosis-data"

	// TODO This should come from the same logic that builds the server image!!!!!
	containerImage = "kurtosistech/kurtosis-core_api"
)

type ApiContainerLauncher struct {
	kurtosisBackend backend_interface.KurtosisBackend
}

func NewApiContainerLauncher(kurtosisBackend backend_interface.KurtosisBackend) *ApiContainerLauncher {
	return &ApiContainerLauncher{kurtosisBackend: kurtosisBackend}
}

func (launcher ApiContainerLauncher) LaunchWithDefaultVersion(
	ctx context.Context,
	logLevel logrus.Level,
	enclaveId enclave.EnclaveID,
	grpcListenPort uint16,
	grpcProxyListenPort uint16,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	backendConfigSupplier KurtosisBackendConfigSupplier,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {
	resultApiContainer, err := launcher.LaunchWithCustomVersion(
		ctx,
		DefaultVersion,
		logLevel,
		enclaveId,
		grpcListenPort,
		grpcProxyListenPort,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
		backendConfigSupplier,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container with default version tag '%v'", DefaultVersion)
	}
	return resultApiContainer, nil
}

func (launcher ApiContainerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	enclaveId enclave.EnclaveID,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	backendConfigSupplier KurtosisBackendConfigSupplier,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {
	kurtosisBackendType, kurtosisBackendConfig := backendConfigSupplier.getKurtosisBackendConfig()
	argsObj, err := args.NewAPIContainerArgs(
		imageVersionTag,
		logLevel.String(),
		grpcPortNum,
		grpcProxyPortNum,
		string(enclaveId),
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
		enclaveDataVolumeDirpath,
		kurtosisBackendType,
		kurtosisBackendConfig,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	envVars, ownIpAddressEnvvar, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	logrus.Debugf("Launching Kurtosis API container...")
	apiContainer, err := launcher.kurtosisBackend.CreateAPIContainer(
		ctx,
		containerImageAndTag,
		enclaveId,
		grpcPortNum,
		grpcProxyPortNum,
		enclaveDataVolumeDirpath,
		ownIpAddressEnvvar,
		envVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}

	return apiContainer, nil
}
