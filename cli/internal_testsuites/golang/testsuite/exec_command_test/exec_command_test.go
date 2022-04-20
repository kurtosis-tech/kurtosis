package exec_command_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testName = "exec-command"
	isPartitioningEnabled = false

	execCmdTestImage      = "alpine:3.12.4"
	inputForLogOutputTest = "hello"
	expectedLogOutput     = "hello\n"
	testServiceId         = "test"

	successExitCode int32 = 0

)

var execCommandThatShouldWork = []string{
	"true",
}

var execCommandThatShouldHaveLogOutput = []string{
	"echo",
	inputForLogOutputTest,
}

var execCommandThatShouldFail = []string{
	"false",
}

func TestExecCommand(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfigSupplier := getContainerConfigSupplier()

	testServiceContext, err := enclaveCtx.AddService(testServiceId, containerConfigSupplier)
	require.NoError(t, err, "An error occurred starting service '%v'", testServiceId)

	// ------------------------------------- TEST RUN ----------------------------------------------
	logrus.Infof("Running exec command '%v' that should return a successful exit code...", execCommandThatShouldWork)
	shouldWorkExitCode, _, err := runExecCmd(testServiceContext, execCommandThatShouldWork)
	require.NoError(t, err, "An error occurred running exec command '%v'", execCommandThatShouldWork)
	require.Equal(t, successExitCode, shouldWorkExitCode, "Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldWork, shouldWorkExitCode)
	logrus.Info("Exec command returned successful exit code as expected")

	logrus.Infof("Running exec command '%v' that should return an error exit code...", execCommandThatShouldFail)
	shouldFailExitCode, _, err := runExecCmd(testServiceContext, execCommandThatShouldFail)
	require.NoError(t, err, "An error occurred running exec command '%v'", execCommandThatShouldFail)
	require.NotEqual(t, successExitCode, shouldFailExitCode, "Exec command '%v' should fail, but got successful exit code %v", execCommandThatShouldFail, successExitCode)

	logrus.Infof("Running exec command '%v' that should return log output...", execCommandThatShouldHaveLogOutput)
	shouldHaveLogOutputExitCode, logOutput, err := runExecCmd(testServiceContext, execCommandThatShouldHaveLogOutput)
	require.NoError(t, err, "An error occurred running exec command '%v'", execCommandThatShouldHaveLogOutput)
	require.Equal(t, successExitCode, shouldHaveLogOutputExitCode, "Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldHaveLogOutput, shouldHaveLogOutputExitCode)
	require.Equal(t, expectedLogOutput, logOutput, "Exec command '%v' should return %v, but got %v.", execCommandThatShouldHaveLogOutput, inputForLogOutputTest, logOutput)
	logrus.Info("Exec command returned error exit code as expected")
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		// We sleep because the only function of this container is to test Docker executing a command while it's running
		// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
		// args), but this provides a nice little regression test of the ENTRYPOINT overriding
		entrypointArgs := []string{"sleep"}
		cmdArgs := []string{"30"}

		containerConfig := services.NewContainerConfigBuilder(
			execCmdTestImage,
		).WithEntrypointOverride(
			entrypointArgs,
		).WithCmdOverride(
			cmdArgs,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}

func runExecCmd(serviceContext *services.ServiceContext, command []string) (int32, string, error) {
	exitCode, logOutput, err := serviceContext.ExecCommand(command)
	if err != nil {
		return 0, "", stacktrace.Propagate(
			err,
			"An error occurred executing command '%v'", command)
	}
	return exitCode, logOutput, nil
}
