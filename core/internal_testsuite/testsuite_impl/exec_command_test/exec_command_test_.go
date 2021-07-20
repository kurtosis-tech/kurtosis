package exec_command_test

import (
	"fmt"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	execCmdTestImage      = "alpine:3.12.4"
	inputForLogOutputTest = "hello"
	expectedLogOutput     = "hello\n"
	testServiceId         = "test"

	successExitCode int32 = 0

	waitForStartupTimeBetweenPolls = 1 * time.Second
	waitForStartupMaxPolls         = 10
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

type ExecCommandTest struct{}

func (e ExecCommandTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(30).WithRunTimeoutSeconds(30)
}

func (e ExecCommandTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {
	containerCreationConfig, runConfigFunc := getServiceConfigurations()

	_, _, err := networkCtx.AddService(testServiceId, containerCreationConfig, runConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred starting service '%v'",
			testServiceId)
	}
	return networkCtx, nil
}

func (e ExecCommandTest) Run(uncastedNetwork networks.Network) error {
	// Necessary because Go doesn't have generics
	network := uncastedNetwork.(*networks.NetworkContext)

	testServiceContext, err := network.GetServiceContext(testServiceId)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the test service context")
	}

	logrus.Infof("Running exec command '%v' that should return a successful exit code...", execCommandThatShouldWork)
	shouldWorkExitCode, _, err := runExecCmd(testServiceContext, execCommandThatShouldWork)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldWork)
	}
	if shouldWorkExitCode != successExitCode {
		return stacktrace.NewError("Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldWork, shouldWorkExitCode)
	}
	logrus.Info("Exec command returned successful exit code as expected")

	logrus.Infof("Running exec command '%v' that should return an error exit code...", execCommandThatShouldFail)
	shouldFailExitCode, _, err := runExecCmd(testServiceContext, execCommandThatShouldFail)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldFail)
	}
	if shouldFailExitCode == successExitCode {
		return stacktrace.NewError("Exec command '%v' should fail, but got successful exit code %v", execCommandThatShouldFail, successExitCode)
	}

	logrus.Infof("Running exec command '%v' that should return log output...", execCommandThatShouldHaveLogOutput)
	shouldHaveLogOutputExitCode, logOutput, err := runExecCmd(testServiceContext, execCommandThatShouldHaveLogOutput)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred running exec command '%v'", execCommandThatShouldHaveLogOutput)
	}
	if shouldHaveLogOutputExitCode != successExitCode {
		return stacktrace.NewError("Exec command '%v' should work, but got unsuccessful exit code %v", execCommandThatShouldHaveLogOutput, shouldHaveLogOutputExitCode)
	}
	logOutputStr := fmt.Sprintf("%s", *logOutput)
	if logOutputStr != expectedLogOutput {
		return stacktrace.NewError("Exec command '%v' should return %v, but got %v.", execCommandThatShouldHaveLogOutput, inputForLogOutputTest, logOutputStr)
	}
	logrus.Info("Exec command returned error exit code as expected")

	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================

func getServiceConfigurations() (*services.ContainerCreationConfig, func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error)) {
	containerCreationConfig := getContainerCreationConfig()

	runConfigFunc := getRunConfigFunc()
	return containerCreationConfig, runConfigFunc
}

func getContainerCreationConfig() *services.ContainerCreationConfig {
	return services.NewContainerCreationConfigBuilder(execCmdTestImage).Build()
}

func getRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	return func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
		// We sleep because the only function of this container is to test Docker exec'ing a command while it's running
		// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
		// args), but this provides a nice little regression test of the ENTRYPOINT overriding
		entrypointArgs := []string{"sleep"}
		cmdArgs := []string{"30"}
		result := services.NewContainerRunConfigBuilder().WithEntrypointOverride(
			entrypointArgs,
		).WithCmdOverride(
			cmdArgs,
		).Build()
		return result, nil
	}
}

func runExecCmd(serviceContext *services.ServiceContext, command []string) (int32, *[]byte, error) {
	exitCode, logOutput, err := serviceContext.ExecCommand(command)
	if err != nil {
		return 0, nil, stacktrace.Propagate(
			err,
			"An error occurred executing command '%v'", command)
	}
	return exitCode, logOutput, nil
}

