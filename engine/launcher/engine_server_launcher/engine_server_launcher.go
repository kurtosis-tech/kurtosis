/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"context"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_interface"
	"net"
	"time"

	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!
	KurtosisEngineVersion = "1.17.5"
	// !!!!!!!!!!!!!!!!!! DO NOT MODIFY THIS! IT WILL BE UPDATED AUTOMATICALLY DURING THE RELEASE PROCESS !!!!!!!!!!!!!!!

	// TODO This should come from the same logic that builds the server image!!!!!
	containerImage = "kurtosistech/kurtosis-engine-server"

	dockerSocketFilepath = "/var/run/docker.sock"

	networkToStartEngineContainerIn = "bridge"

	// The engine server uses gRPC so MUST listen on TCP (no other protocols are supported)
	// This is the Docker constant indicating a TCP port
	grpcPortProtocol = "tcp"

	grpcProxyPortProtocol = "tcp"

	// The protocol string we use in the netstat command used to ensure the engine container is available
	netstatWaitForAvailabilityPortProtocol = "tcp"

	maxWaitForAvailabilityRetries         = 10
	timeBetweenWaitForAvailabilityRetries = 1 * time.Second

	availabilityWaitingExecCmdSuccessExitCode = 0

	publicPortNumParsingBase     = 10
	publicPortNumParsingUintBits = 16

	// The location where the engine data directory (on the Docker host machine) will be bind-mounted
	//  on the engine server
	EngineDataDirpathOnEngineServerContainer = "/engine-data"
)

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
	grpcProxyListenPortNum uint16, // Envoy proxy port that will forward grpc-web calls to the engine
	engineDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultPublicIpAddr net.IP,
	resultPublicGrpcPortNum uint16,
	// NOTE: We can return a resultPublicGrpcProxyPortNum here if we ever need it
	resultErr error,
) {
	publicIpAddr, publicGrpcPortNum, err := launcher.LaunchWithCustomVersion(
		ctx,
		KurtosisEngineVersion,
		logLevel,
		grpcListenPortNum,
		grpcProxyListenPortNum,
		engineDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred launching the engine server container with default version tag '%v'", KurtosisEngineVersion)
	}
	return publicIpAddr, publicGrpcPortNum, nil
}

func (launcher *EngineServerLauncher) LaunchWithCustomVersion(
	ctx context.Context,
	imageVersionTag string,
	logLevel logrus.Level,
	grpcListenPortNum uint16, // The port that the engine server will listen on AND the port that it should be bound to on the host machine
	grpcProxyListenPortNum uint16, // Envoy proxy port that will forward grpc-web calls to the engine
	engineDataDirpathOnHostMachine string,
	metricsUserID string,
	didUserAcceptSendingMetrics bool,
) (
	resultPublicIpAddr net.IP,
	resultPublicGrpcPortNum uint16,
	resultErr error,
) {
	argsObj, err := args.NewEngineServerArgs(
		grpcListenPortNum,
		grpcProxyListenPortNum,
		logLevel.String(),
		imageVersionTag,
		engineDataDirpathOnHostMachine,
		metricsUserID,
		didUserAcceptSendingMetrics,
	)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred creating the engine server args")
	}

	envVars, err := args.GetEnvFromArgs(argsObj)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred generating the engine server's environment variables")
	}

	engine, err := launcher.kurtosisBackend.CreateEngine(ctx, containerImage, imageVersionTag, grpcListenPortNum, grpcProxyListenPortNum, engineDataDirpathOnHostMachine, envVars)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "An error occurred launching the engine server container with environment variables '%+v'", envVars)
	}
	return engine.GetPublicIPAddress(), engine.GetPublicGRPCPort().GetNumber(), nil
}
