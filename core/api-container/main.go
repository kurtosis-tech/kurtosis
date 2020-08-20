package main

import (
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"os"
)





func main() {
	err := initializeAndRun()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func initializeAndRun() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return stacktrace.Propagate(err, "Could not initialize a Docker client from the environment")
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Docker manager")
	}

	// TODO make these parameterized
	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		"172.17.0.0/16",
		map[string]bool{
			"172.17.0.1": true,
		})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the free IP address tracker")
	}

	kurtosisApi := &KurtosisAPI{
		dockerManager: dockerManager,
		dockerNetworkId: "b453ce4bac01", // TODO make this parameterizable
		freeIpAddrTracker: freeIpAddrTracker,
	}

	logrus.Info("Launching server...")

	server := rpc.NewServer()
	jsonCodec := json2.NewCodec()
	server.RegisterCodec(jsonCodec, "application/json")
	server.RegisterService(kurtosisApi, "")

	http.Handle("/jsonrpc", server)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))

	return nil
}