/* * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	enclaveDataVolumeDirpath = "/kurtosis-data"

	// TODO This should come from the same logic that builds the server image!!!!!
	containerImage = "kurtosistech/core"
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
	enclaveId enclave.EnclaveUUID,
	grpcListenPort uint16,
	backendConfigSupplier KurtosisBackendConfigSupplier,
	enclaveEnvVars string,
	isProductionEnclave bool,
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
	resultApiContainer, err := launcher.LaunchWithCustomVersion(
		ctx,
		kurtosis_version.KurtosisVersion,
		logLevel,
		enclaveId,
		grpcListenPort,
		backendConfigSupplier,
		enclaveEnvVars,
		isProductionEnclave,
		metricsUserID,
		didUserAcceptSendingMetrics,
		isCI,
		cloudUserID,
		cloudInstanceID,
		shouldStartInDebugMode,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container with default version tag '%v'", kurtosis_version.KurtosisVersion)
	}
	return resultApiContainer, nil
}

func (launcher ApiContainerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	enclaveUuid enclave.EnclaveUUID,
	grpcPortNum uint16,
	backendConfigSupplier KurtosisBackendConfigSupplier,
	enclaveEnvVars string,
	isProductionEnclave bool,
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
	kurtosisBackendType, kurtosisBackendConfig := backendConfigSupplier.getKurtosisBackendConfig()
	argsObj, err := args.NewAPIContainerArgs(
		imageVersionTag,
		logLevel.String(),
		grpcPortNum,
		string(enclaveUuid),
		enclaveDataVolumeDirpath,
		kurtosisBackendType,
		kurtosisBackendConfig,
		enclaveEnvVars,
		isProductionEnclave,
		metricsUserID,
		didUserAcceptSendingMetrics,
		isCI,
		cloudUserID,
		cloudInstanceID,
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
		enclaveUuid,
		grpcPortNum,
		enclaveDataVolumeDirpath,
		ownIpAddressEnvvar,
		envVars,
		shouldStartInDebugMode,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}

	return apiContainer, nil
}
