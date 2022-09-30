//go:build !minikube
// +build !minikube

// We don't run this test in Kubernetes because, as of 2022-07-07, Kubernetes doesn't support network partitioning

package network_soft_partition_test

import (
	"context"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName              = "network-soft-partition"
	isPartitioningEnabled = true

	dockerGettingStartedImage                             = "docker/getting-started"
	exampleServiceId                   services.ServiceID = "docker-getting-started"
	kurtosisIpRoute2DockerImageName                       = "kurtosistech/iproute2"
	testService                        services.ServiceID = "test-service"
	exampleServicePortNumInsideNetwork                    = 80

	execCommandSuccessExitCode = int32(0)

	exampleServicePartitionID enclaves.PartitionID = "example"
	testServicePartitionID    enclaves.PartitionID = "test"

	exampleServiceMainPortID = "main"

	sleepCmd = "sleep"

	testServiceSleepMillisecondsStr = "300000"

	percentageSign                    = "%"
	zeroPacketLoss                    = float64(0)
	softPartitionPacketLossPercentage = float32(99)

	zeroElementsInMtrHubField = 0
)

type MtrReport struct {
	Report struct {
		Hubs []struct {
			Loss float64 `json:"Loss%"`
		} `json:"hubs"`
	} `json:"report"`
}

func TestNetworkSoftPartitions(t *testing.T) {
	ctx := context.Background()
	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	exampleServiceContainerConfig := getExampleServiceConfig()

	exampleServiceCtx, err := enclaveCtx.AddService(exampleServiceId, exampleServiceContainerConfig)
	require.NoError(t, err, "An error occurred adding the datastore service")
	logrus.Debugf("Example service IP: %v", exampleServiceCtx.GetPrivateIPAddress())

	testServiceContainerConfig := getTestServiceContainerConfig()

	testServiceCtx, err := enclaveCtx.AddService(testService, testServiceContainerConfig)
	require.NoError(t, err, "An error occurred adding the file server service")
	logrus.Debugf("Test service IP: %v", testServiceCtx.GetPrivateIPAddress())

	installMtrCmd := []string{
		"apk",
		"add",
		"mtr",
	}

	exitCode, _, err := testServiceCtx.ExecCommand(installMtrCmd)
	require.NoError(t, err, "An error occurred executing command '%+v' ", installMtrCmd)
	require.Equal(
		t,
		execCommandSuccessExitCode,
		exitCode,
		"Command '%+v' to install mtr cli exited with non-successful exit code '%v'",
		installMtrCmd,
		exitCode,
	)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Info("Executing mtr report to check there is no packet loss in services' communication before soft partition...")
	mtrReportCmd := []string{
		"mtr",
		exampleServiceCtx.GetPrivateIPAddress(),
		"--report",
		"--json",
		"--report-cycles",
		"2",        // We set report cycles to 2 to generate the report faster because default is 10
		"--no-dns", //No domain name resolution, also to improve velocity
	}

	exitCode, logOutput, err := testServiceCtx.ExecCommand(mtrReportCmd)
	require.NoError(t, err, "An error occurred executing command '%+v' ", mtrReportCmd)
	require.Equal(
		t,
		execCommandSuccessExitCode,
		exitCode,
		"Command '%+v' to run mtr report exited with non-successful exit code '%v'",
		mtrReportCmd,
		exitCode,
	)

	jsonStr := logOutput
	logrus.Debugf("MTR report before soft partition result:\n  %+v", jsonStr)

	mtrReportBeforeSoftPartition := new(MtrReport)
	err = json.Unmarshal([]byte(jsonStr), mtrReportBeforeSoftPartition)
	require.NoError(t, err, "An error occurred unmarshalling json string '%v' to mtr report struct ", jsonStr)
	require.Greaterf(t, len(mtrReportBeforeSoftPartition.Report.Hubs), zeroElementsInMtrHubField, "There isn't any element in the report hub field")
	require.Equal(t, zeroPacketLoss, mtrReportBeforeSoftPartition.Report.Hubs[0].Loss)
	logrus.Info("Report complete successfully, there was no packet loss between services during the test")

	logrus.Infof("Executing soft partition with packet loss %v%v...", softPartitionPacketLossPercentage, percentageSign)
	softPartitionConnection, err := enclaves.NewSoftPartitionConnection(softPartitionPacketLossPercentage)
	require.NoError(t, err, "An error occurred creating new soft partition with packet loss %v%v", softPartitionPacketLossPercentage, percentageSign)

	err = repartitionNetwork(enclaveCtx, softPartitionConnection)
	require.NoError(t, err, "An error occurred executing repartition network")
	logrus.Info("Partition complete")

	logrus.Info("Executing mtr report to check there is packet loss in services' communication after soft partition...")
	exitCode, logOutput, err = testServiceCtx.ExecCommand(mtrReportCmd)
	require.NoError(t, err, "An error occurred executing command '%+v' ", mtrReportCmd)
	require.Equal(
		t,
		execCommandSuccessExitCode,
		exitCode,
		"Command '%+v' to run mtr report exited with non-successful exit code '%v'",
		mtrReportCmd,
		exitCode,
	)

	jsonStr = logOutput
	logrus.Debugf("MTR report after soft partition result:\n  %+v", jsonStr)

	mtrReportAfterPartition := new(MtrReport)
	err = json.Unmarshal([]byte(jsonStr), mtrReportAfterPartition)
	require.NoError(t, err, "An error occurred unmarshalling json string '%v' to mtr report struct ", jsonStr)
	require.Equalf(t, zeroElementsInMtrHubField, len(mtrReportAfterPartition.Report.Hubs), "The absence of hub's elements means that all packets were lost, so shouldn't be any hub's elements on the report but it contains %v elements", len(mtrReportAfterPartition.Report.Hubs))
	logrus.Infof("Report complete successfully, no package was sent")

	logrus.Info("Executing repartition network to unblock partition and join services again...")
	unblockedPartitionConnection := enclaves.NewUnblockedPartitionConnection()
	err = repartitionNetwork(enclaveCtx, unblockedPartitionConnection)
	require.NoError(t, err, "An error occurred executing repartition network")
	logrus.Info("Partitions unblocked successfully")

	logrus.Info("Executing mtr report to check there is no packet loss in services' communication after unblocking partition...")
	exitCode, logOutput, err = testServiceCtx.ExecCommand(mtrReportCmd)
	require.NoError(t, err, "An error occurred executing command '%+v' ", mtrReportCmd)
	require.Equal(
		t,
		execCommandSuccessExitCode,
		exitCode,
		"Command '%+v' to run mtr report exited with non-successful exit code '%v'",
		mtrReportCmd,
		exitCode,
	)

	jsonStr = logOutput
	logrus.Debugf("MTR report after unblocking partition result:\n  %+v", jsonStr)

	mtrReportAfterUnblockedPartition := new(MtrReport)
	err = json.Unmarshal([]byte(jsonStr), mtrReportAfterUnblockedPartition)
	require.NoError(t, err, "An error occurred unmarshalling json string '%v' to mtr report struct ", jsonStr)
	require.Greaterf(t, len(mtrReportAfterUnblockedPartition.Report.Hubs), zeroElementsInMtrHubField, "There aren't any element in the report hub field")
	require.Equal(t, zeroPacketLoss, mtrReportAfterUnblockedPartition.Report.Hubs[0].Loss)
	logrus.Info("Report complete successfully, there was no packet loss between services during the test")
}

func repartitionNetwork(
	enclaveCtx *enclaves.EnclaveContext,
	partitionConnection enclaves.PartitionConnection,
) error {

	partitionServices := map[enclaves.PartitionID]map[services.ServiceID]bool{
		exampleServicePartitionID: {
			exampleServiceId: true,
		},
		testServicePartitionID: {
			testService: true,
		},
	}
	partitionConnections := map[enclaves.PartitionID]map[enclaves.PartitionID]enclaves.PartitionConnection{
		exampleServicePartitionID: {
			testServicePartitionID: partitionConnection,
		},
	}
	defaultPartitionConnection := partitionConnection
	if err := enclaveCtx.RepartitionNetwork(partitionServices, partitionConnections, defaultPartitionConnection); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred repartitioning the network with partition connection = %v",
			partitionConnection,
		)
	}
	return nil
}

func getExampleServiceConfig() *services.ContainerConfig {
	portSpec := services.NewPortSpec(exampleServicePortNumInsideNetwork, services.PortProtocol_TCP)
	containerConfig := services.NewContainerConfigBuilder(
		dockerGettingStartedImage,
	).WithUsedPorts(
		map[string]*services.PortSpec{exampleServiceMainPortID: portSpec},
	).Build()
	return containerConfig
}

func getTestServiceContainerConfig() *services.ContainerConfig {

	// We sleep because the only function of this container is to test Docker executing a command while it's running
	// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
	// args), but this provides a nice little regression test of the ENTRYPOINT overriding
	entrypointArgs := []string{sleepCmd}
	cmdArgs := []string{testServiceSleepMillisecondsStr}

	containerConfig := services.NewContainerConfigBuilder(
		kurtosisIpRoute2DockerImageName,
	).WithEntrypointOverride(
		entrypointArgs,
	).WithCmdOverride(
		cmdArgs,
	).Build()
	return containerConfig
}
