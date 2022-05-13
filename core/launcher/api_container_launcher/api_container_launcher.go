/* * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"context"
	"fmt"
	"github.com/google/martian/log"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/api_container"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface/objects/enclave"
	"github.com/kurtosis-tech/kurtosis-core/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"net"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	DefaultVersion = "1.46.2"
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
	networkId string,
	subnetMask string,
	grpcListenPort uint16,
	grpcProxyListenPort uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {
	resultApiContainer, err := launcher.LaunchWithCustomVersion(
		ctx,
		DefaultVersion,
		logLevel,
		enclaveId,
		networkId,
		subnetMask,
		grpcListenPort,
		grpcProxyListenPort,
		gatewayIpAddr,
		apiContainerIpAddr,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
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
	networkId string,
	subnetMask string,
	grpcPortNum uint16,
	grpcProxyPortNum uint16,
	gatewayIpAddr net.IP,
	apiContainerIpAddr net.IP,
	isPartitioningEnabled bool,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultApiContainer *api_container.APIContainer,
	resultErr error,
) {

	takenIpAddrStrSet := map[string]bool{
		gatewayIpAddr.String():      true,
		apiContainerIpAddr.String(): true,
	}
	argsObj, err := args.NewAPIContainerArgs(
		imageVersionTag,
		logLevel.String(),
		grpcPortNum,
		grpcProxyPortNum,
		string(enclaveId),
		networkId,
		subnetMask,
		apiContainerIpAddr.String(),
		takenIpAddrStrSet,
		isPartitioningEnabled,
		metricsUserID,
		didUserAcceptSendingMetrics,
		enclaveDataVolumeDirpath,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the API container args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred generating the API container's environment variables")
	}

	containerImageAndTag := fmt.Sprintf(
		"%v:%v",
		containerImage,
		imageVersionTag,
	)

	log.Debugf("Launching Kurtosis API container...")
	apiContainer, err := launcher.kurtosisBackend.CreateAPIContainer(
		ctx,
		containerImageAndTag,
		enclaveId,
		apiContainerIpAddr,
		grpcPortNum,
		grpcProxyPortNum,
		enclaveDataVolumeDirpath,
		envVars,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred launching the API container")
	}

	return apiContainer, nil
}
