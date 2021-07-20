package local_static_file_test

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/networks"
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis/internal_testsuite/testsuite_impl/static_file_consts"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	dockerImage                    = "alpine:3.12.4"
	testService services.ServiceID = "test-service"

	execCommandSuccessExitCode = 0
	expectedTestFile1Contents  = "This is a test static file"
	expectedTestFile2Contents  = "This is another test static file"
)

type LocalStaticFileTest struct{}

func (l LocalStaticFileTest) Configure(builder *testsuite.TestConfigurationBuilder) {
	builder.WithSetupTimeoutSeconds(60).WithRunTimeoutSeconds(60)
}

func (l LocalStaticFileTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {

	containerCreationConfig, runConfigFunc := getServiceConfigurations()

	_, _, err := networkCtx.AddService(testService, containerCreationConfig, runConfigFunc)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred adding the file server service")
	}
	return networkCtx, nil
}

func (l LocalStaticFileTest) Run(network networks.Network) error {
	castedNetwork, ok := network.(*networks.NetworkContext)
	if !ok {
		return stacktrace.NewError("An error occurred casting the uncasted network to a NetworkContext")
	}

	serviceCtx, err := castedNetwork.GetServiceContext(testService)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting service '%v'", testService)
	}

	staticFileAbsFilepaths, err := serviceCtx.LoadStaticFiles(map[services.StaticFileID]bool{
		static_file_consts.TestStaticFile1ID: true,
		static_file_consts.TestStaticFile2ID: true,
	})
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred loading the static file corresponding to key '%v'", static_file_consts.TestStaticFile1ID)
	}
	testFile1AbsFilepath, found := staticFileAbsFilepaths[static_file_consts.TestStaticFile1ID]
	if !found {
		return stacktrace.Propagate(err, "No filepath found for test file 1 key '%v'; this is a bug in Kurtosis!", static_file_consts.TestStaticFile1ID)
	}
	testFile2AbsFilepath, found := staticFileAbsFilepaths[static_file_consts.TestStaticFile2ID]
	if !found {
		return stacktrace.Propagate(err, "No filepath found for test file 2 key '%v'; this is a bug in Kurtosis!", static_file_consts.TestStaticFile2ID)
	}

	// Test file 1
	catStaticFile1Cmd := []string{
		"cat",
		testFile1AbsFilepath,
	}
	exitCode1, outputBytes1, err := serviceCtx.ExecCommand(catStaticFile1Cmd)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static test file 1 contents", catStaticFile1Cmd)
	}
	if exitCode1 != execCommandSuccessExitCode {
		return stacktrace.NewError("Command '%+v' to cat the static test file 1 exited with non-successful exit code '%v'", catStaticFile1Cmd, exitCode1)
	}
	file1Contents := string(*outputBytes1)
	if file1Contents != expectedTestFile1Contents {
		return stacktrace.NewError("Static file contents '%v' don't match expected test file 1 contents '%v'", file1Contents, expectedTestFile1Contents)
	}
	logrus.Infof("Static file 1 contents were '%v' as expected", expectedTestFile1Contents)

	// Test file 2
	catStaticFile2Cmd := []string{
		"cat",
		testFile2AbsFilepath,
	}
	exitCode2, outputBytes2, err := serviceCtx.ExecCommand(catStaticFile2Cmd)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static test file 2 contents", catStaticFile2Cmd)
	}
	if exitCode2 != execCommandSuccessExitCode {
		return stacktrace.NewError("Command '%+v' to cat the static test file 2 exited with non-successful exit code '%v'", catStaticFile2Cmd, exitCode2)
	}
	file2Contents := string(*outputBytes2)
	if file2Contents != expectedTestFile2Contents {
		return stacktrace.NewError("Static file contents '%v' don't match expected test file 2 contents '%v'", file2Contents, expectedTestFile2Contents)
	}
	logrus.Infof("Static file 2 contents were '%v' as expected", expectedTestFile2Contents)

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
	return services.NewContainerCreationConfigBuilder(dockerImage).Build()
}

func getRunConfigFunc() func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
	runConfigFunc := func(ipAddr string, generatedFileFilepaths map[string]string, staticFileFilepaths map[services.StaticFileID]string) (*services.ContainerRunConfig, error) {
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
	return runConfigFunc
}
