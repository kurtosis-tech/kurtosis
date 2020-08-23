package main

import (
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kurtosis-tech/kurtosis/commons/docker"
	"github.com/kurtosis-tech/kurtosis/commons/networks"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
)

const (
	testNamesFilepathArg = "TEST_NAMES_FILEPATH"

	testNamesFilepath = "/shared/test-names-filepath"
)

func main() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker client: %v", err)
		os.Exit(1)
	}

	// Create the tempfile that the testsuite image will write test names to
	tempFp, err := ioutil.TempFile("", "test-names")
	if err != nil {
		logrus.Errorf("An error occurred creating the temp filepath: %v", err)
		os.Exit(1)
	}
	tempFp.Close()

	freeIpAddrTracker, err := networks.NewFreeIpAddrTracker(
		logrus.StandardLogger(),
		"172.17.0.0/16",
		map[string]bool{
			"172.17.0.1": true,	// gateway IP
		})
	if err != nil {
		logrus.Errorf("An error occurred creating the free IP address tracker: %v", err)
		os.Exit(1)
	}

	testSuiteContainerIp, err := freeIpAddrTracker.GetFreeIpAddr()
	if err != nil {
		logrus.Errorf("An error occurred getting a free IP address for the testsuite container: %v", err)
		os.Exit(1)
	}

	dockerManager, err := docker.NewDockerManager(logrus.StandardLogger(), dockerClient)
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker manager: %v", err)
		os.Exit(1)
	}

	containerId, err := dockerManager.CreateAndStartContainer(
		context.Background(),
		// TODO parameterize these
		"kevin-test",
		"b453ce4bac01",
		testSuiteContainerIp,
		map[nat.Port]bool{},
		nil,
		map[string]string{
			testNamesFilepathArg: testNamesFilepath,
		},
		map[string]string{
			tempFp.Name(): testNamesFilepath,
		},
		map[string]string{})
	if err != nil {
		logrus.Errorf("An error occurred creating the Docker container to run the test suite: %v", err)
		os.Exit(1)
	}

	exitCode, err := dockerManager.WaitForExit(
		context.Background(),
		containerId)
	if err != nil {
		logrus.Errorf("An error occurred waiting for the testsuite container to exit: %v", err)
		os.Exit(1)
	}
	if exitCode != 0 {
		logrus.Error("The testsuite container exited with a nonzero exit code")
		os.Exit(1)
	}

	tempFpReader, err := os.Open(tempFp.Name())
	if err != nil {
		logrus.Errorf("An error occurred opening the temp filename containing test names for reading: %v", err)
		os.Exit(1)
	}
	defer tempFpReader.Close()

	io.Copy(logrus.StandardLogger().Out, tempFpReader)
}