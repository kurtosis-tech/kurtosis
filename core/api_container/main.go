package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/kurtosis-tech/kurtosis/api_container/api"
	"github.com/kurtosis-tech/kurtosis/api_container/logging"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
)

func main() {
	// NOTE: we'll want to chnage the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})

	testSuiteContainerIdArg := flag.String(
		"test-suite-container-id",
		"",
		"ID of the Docker container running the test suite",
	)

	networkIdArg := flag.String(
		"network-id",
		"",
		"ID of the Docker network that the API container is running in, and in which all services should be started",
	)

	subnetMaskArg := flag.String(
		"subnet-mask",
		"",
		"Subnet mask of the Docker network that the API container is running in",
	)

	gatewayIpArg := flag.String(
		"gateway-ip",
		"",
		"IP address of the gateway address on the Docker network that the test controller is running in",
	)

	containerIpArg := flag.String(
		"container-ip",
		"",
		"IP address of the Docker container running the API container",
	)

	logLevelArg := flag.String(
		"log-level",
		"info",
		fmt.Sprintf("Log level to use for the API container (%v)", logging.GetAcceptableStrings()),
	)

	// TODO add a flag to write output to both STDOUT and a file using io.MultiWriter

	flag.Parse()

	logLevelPtr := logging.LevelFromString(*logLevelArg)
	if logLevelPtr == nil {
		// It's a little goofy that we're logging an error before we've set the loglevel, but we do so at the highest
		//  level so that whatever the default the user should see it
		logrus.Fatalf("Invalid initializer log level %v", *logLevelArg)
		os.Exit(1)
	}
	logrus.SetLevel(*logLevelPtr)

	// A value on this channel indicates that a test was registered and it either completed or timed out, and the
	//  bool value will be "true" if the test execution ended before timeout and "false" if the timeout was hit
	testExecutionEndedBeforeTimeoutChan := make(chan bool, 1)

	server, err := createServer(
		testExecutionEndedBeforeTimeoutChan,
		*testSuiteContainerIdArg,
		*networkIdArg,
		*subnetMaskArg,
		*gatewayIpArg,
		*containerIpArg)
	if err != nil {
		logrus.Error("Failed to create a server with the following error:")
		fmt.Fprint(logrus.StandardLogger().Out, err)
		os.Exit(1)
	}

	go func(){
		server.ListenAndServe()
	}()

	// Docker will send SIGTERM to end the process, and we need to catch it to stop gracefully
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	logrus.Info("Waiting for stop signal or test completion...")
	var exitCode int
	select {
	case signal := <- signalChan:
		logrus.Infof("Received signal %v; server will shut down", signal)
		exitCode = 0
	case testExecutionEndedBeforeTimeout := <- testExecutionEndedBeforeTimeoutChan:
		if testExecutionEndedBeforeTimeout {
			logrus.Info("Test execution ended before timeout as expected")
			exitCode = 0
		} else {
			logrus.Error("Test execution hit timeout")
			exitCode = 1
		}
	}

	// NOTE: Might need to kick off a timeout thread to separately close the context if it's taking too long or if
	//  the server hangs forever trying to shutdown
	logrus.Info("Shutting down JSON RPC server...")
	server.Shutdown(context.Background())
	logrus.Info("JSON RPC server shut down")

	os.Exit(exitCode)
}

func createServer(
		testExecutionEndedBeforeTimeoutChan chan bool,
		testSuiteContainerId string,
		networkId string,
		networkSubnetMask string,
		gatewayIp string,
		apiContainerIp string) (*http.Server, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		networkSubnetMask,
		map[string]bool{
			gatewayIp:      true,
			apiContainerIp: true,
		})
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	kurtosisApi := api.NewKurtosisAPI(
		testSuiteContainerId,
		testExecutionEndedBeforeTimeoutChan,
		dockerManager,
		networkId,
		freeIpAddrTracker,
	)

	logrus.Info("Launching server...")

	httpHandler := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	httpHandler.RegisterCodec(jsonCodec, "application/json")
	httpHandler.RegisterService(kurtosisApi, "")
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", api.KurtosisAPIContainerPort),
		Handler: httpHandler,
	}

	return server, nil
}
