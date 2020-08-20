package main

import (
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
	"log"
	"net/http"
	"os"
)

func main() {
	// NOTE: we'll want to chnage the ForceColors to false if we ever want structured logging
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})


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
	flag.Parse()

	logLevelPtr := logging.LevelFromString(*logLevelArg)
	if logLevelPtr == nil {
		// It's a little goofy that we're logging an error before we've set the loglevel, but we do so at the highest
		//  level so that whatever the default the user should see it
		logrus.Fatalf("Invalid initializer log level %v", *logLevelArg)
		os.Exit(1)
	}
	logrus.SetLevel(*logLevelPtr)

	err := initializeAndRun(
		*networkIdArg,
		*subnetMaskArg,
		*gatewayIpArg,
		*containerIpArg)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

/*
Args:
	networkId: The network ID of the Docker network that the API container is running inside
	networkSubnetMask: The subnet mask of the Docker network that the API container is running inside
	gatewayIp: The IP of the gateway inside the Docker network the API container is running in
 */
func initializeAndRun(networkId string, networkSubnetMask string, gatewayIp string, apiContainerIp string) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		networkSubnetMask,
		map[string]bool{
			gatewayIp: true,
			apiContainerIp: true,
		})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	kurtosisApi := api.NewKurtosisAPI(
		dockerManager,
		networkId,
		freeIpAddrTracker,
	)

	logrus.Info("Launching server...")

	server := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterService(kurtosisApi, "")

	http.Handle("/jsonrpc", server)
	log.Fatal(http.ListenAndServe(":8080", nil))

	return nil
}
