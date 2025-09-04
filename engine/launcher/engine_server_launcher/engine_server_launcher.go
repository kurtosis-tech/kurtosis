/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"context"
	"net"
	"os"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_collector"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/port_spec"
	"github.com/kurtosis-tech/kurtosis/engine/launcher/args"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/kurtosis/metrics-library/golang/lib/metrics_client"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// Default engine image, can be overridden via KURTOSIS_ENGINE_IMAGE env var
	defaultContainerImage = "kurtosistech/engine"
	engineImageEnvVar     = "KURTOSIS_ENGINE_IMAGE"
)

// containerImage is the actual image to use, checking env var override first
var containerImage = getImageWithEnvOverride(engineImageEnvVar, defaultContainerImage)

func getImageWithEnvOverride(envVar, defaultImage string) string {
	if override := os.Getenv(envVar); override != "" {
		logrus.Infof("Using custom engine image from %s: %s", envVar, override)
		return override
	}
	return defaultImage
}

type EngineServerLauncher struct {
	kurtosisBackend backend_interface.KurtosisBackend
}

func NewEngineServerLauncher(kurtosisBackend backend_interface.KurtosisBackend) *EngineServerLauncher {
	return &EngineServerLauncher{kurtosisBackend: kurtosisBackend}
}

func (launcher *EngineServerLauncher) LaunchWithDefaultVersion(
	ctx context.Context,
	logLevel logrus.Level,
	grpcListenPortNum uint16, // The port that the engine server will listen on AND the port that it should be bound to on the host machine
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	backendConfigSupplier KurtosisBackendConfigSupplier,
	onBastionHost bool,
	poolSize uint8,
	enclaveEnvVars string,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
	allowedCORSOrigins *[]string,
	shouldStartInDebugMode bool,
	githubAuthToken string,
	restartAPIContainers bool,
	domain string,
	logRetentionPeriod string,
	sinks logs_aggregator.Sinks,
	shouldEnablePersistentVolumeLogsCollection bool,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
) (
	resultPublicIpAddr net.IP,
	resultPublicGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	publicIpAddr, publicGrpcPortSpec, err := launcher.LaunchWithCustomVersion(
		ctx,
		kurtosis_version.KurtosisVersion,
		logLevel,
		grpcListenPortNum,
		metricsUserID,
		didUserAcceptSendingMetrics,
		backendConfigSupplier,
		onBastionHost,
		poolSize,
		enclaveEnvVars,
		isCI,
		cloudUserID,
		cloudInstanceID,
		allowedCORSOrigins,
		shouldStartInDebugMode,
		githubAuthToken,
		restartAPIContainers,
		domain,
		logRetentionPeriod,
		sinks,
		shouldEnablePersistentVolumeLogsCollection,
		logsCollectorFilters,
		logsCollectorParsers,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred launching the engine server container with default version tag '%v'", kurtosis_version.KurtosisVersion)
	}
	return publicIpAddr, publicGrpcPortSpec, nil
}

func (launcher *EngineServerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	grpcListenPortNum uint16, // The port that the engine server will listen on AND the port that it should be bound to on the host machine
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
	backendConfigSupplier KurtosisBackendConfigSupplier,
	onBastionHost bool,
	poolSize uint8,
	enclaveEnvVars string,
	isCI bool,
	cloudUserID metrics_client.CloudUserID,
	cloudInstanceID metrics_client.CloudInstanceID,
	allowedCORSOrigins *[]string,
	shouldStartInDebugMode bool,
	githubAuthToken string,
	restartAPIContainers bool,
	domain string,
	logRetentionPeriod string,
	sinks logs_aggregator.Sinks,
	shouldEnablePersistentVolumeLogsCollection bool,
	logsCollectorFilters []logs_collector.Filter,
	logsCollectorParsers []logs_collector.Parser,
) (
	resultPublicIpAddr net.IP,
	resultPublicGrpcPortSpec *port_spec.PortSpec,
	resultErr error,
) {
	kurtosisBackendType, kurtosisBackendConfig := backendConfigSupplier.getKurtosisBackendConfig()

	argsObj, err := args.NewEngineServerArgs(
		grpcListenPortNum,
		logLevel.String(),
		imageVersionTag,
		metricsUserID,
		didUserAcceptSendingMetrics,
		kurtosisBackendType,
		kurtosisBackendConfig,
		onBastionHost,
		poolSize,
		enclaveEnvVars,
		isCI,
		cloudUserID,
		cloudInstanceID,
		allowedCORSOrigins,
		restartAPIContainers,
		domain,
		logRetentionPeriod,
		logsCollectorFilters,
		logsCollectorParsers,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred creating the engine server args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	engine, err := launcher.kurtosisBackend.CreateEngine(
		ctx,
		containerImage,
		imageVersionTag,
		grpcListenPortNum,
		envVars,
		shouldStartInDebugMode,
		githubAuthToken,
		sinks,
		shouldEnablePersistentVolumeLogsCollection,
		logsCollectorFilters,
		logsCollectorParsers,
	)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "An error occurred launching the engine server container")
	}
	return engine.GetPublicIPAddress(), engine.GetPublicGRPCPort(), nil
}
