package main

import (
	"context"
	"fmt"
	portal_api "github.com/kurtosis-tech/kurtosis/api/golang/portal/kurtosis_portal_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/portal/daemon/server"
	minimal_grpc_server "github.com/kurtosis-tech/minimal-grpc-server/golang/server"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

const (
	successExitCode = 0
	failureExitCode = 1

	grpcServerStopGracePeriod = 5 * time.Second

	forceColors   = true
	fullTimestamp = true

	logMethodAlongWithLogLine = true
	functionPathSeparator     = "."
	emptyFunctionName         = ""
)

func main() {
	ctx := context.Background()
	logrus.SetLevel(logrus.DebugLevel)
	// This allows the filename & function to be reported
	logrus.SetReportCaller(logMethodAlongWithLogLine)
	// NOTE: we'll want to change the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:               forceColors,
		DisableColors:             false,
		ForceQuote:                false,
		DisableQuote:              false,
		EnvironmentOverrideColors: false,
		DisableTimestamp:          false,
		FullTimestamp:             fullTimestamp,
		TimestampFormat:           "",
		DisableSorting:            false,
		SortingFunc:               nil,
		DisableLevelTruncation:    false,
		PadLevelText:              false,
		QuoteEmptyFields:          false,
		FieldMap:                  nil,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fullFunctionPath := strings.Split(f.Function, functionPathSeparator)
			functionName := fullFunctionPath[len(fullFunctionPath)-1]
			_, filename := path.Split(f.File)
			return emptyFunctionName, formatFilenameFunctionForLogs(filename, functionName)
		},
	})

	err := runDaemon(ctx)
	if err != nil {
		logrus.Errorf("An error occurred when running the main function:")
		fmt.Fprintln(logrus.StandardLogger().Out, err)
		os.Exit(failureExitCode)
	}
	os.Exit(successExitCode)
}

func runDaemon(ctx context.Context) error {

	portalServer := server.NewPortalServer()
	defer portalServer.Close()

	kurtosisPortalDaemonRegistrationFunc := func(grpcServer *grpc.Server) {
		portal_api.RegisterKurtosisPortalDaemonServer(grpcServer, portalServer)
	}
	portalServerDaemon := minimal_grpc_server.NewMinimalGRPCServer(
		server.PortalServerGrpcPort,
		grpcServerStopGracePeriod,
		[]func(*grpc.Server){
			kurtosisPortalDaemonRegistrationFunc,
		},
	)

	logrus.Infof("Kurtosis Portal Daemon Server running and listening on port %d", server.PortalServerGrpcPort)
	if err := portalServerDaemon.RunUntilStopped(ctx.Done()); err != nil {
		return stacktrace.Propagate(err, "An error occurred running the Kurtosis Portal Daemon Server")
	}
	return nil
}

func formatFilenameFunctionForLogs(filename string, functionName string) string {
	var output strings.Builder
	output.WriteString("[")
	output.WriteString(filename)
	output.WriteString(":")
	output.WriteString(functionName)
	output.WriteString("]")
	return output.String()
}
