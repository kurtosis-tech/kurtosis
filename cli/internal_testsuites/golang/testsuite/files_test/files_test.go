package local_static_file_test

import (
	"context"
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/test_helpers"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const (
	testName = "files"
	isPartitioningEnabled = false

	dockerImage                    = "alpine:3.12.4"
	testService services.ServiceID = "test-service"

	execCommandSuccessExitCode = int32(0)
	expectedTestFile1Contents  = "This is a test file"
	expectedTestFile2Contents  = "This is another test file"

	generatedFilePermBits = 0644
)
// Mapping of filepath_rel_to_shared_dir_root -> contents
var generatedFileRelPathsAndContents = map[string]string{
	"test1.txt": expectedTestFile1Contents,
	"test2.txt": expectedTestFile2Contents,
}

func TestFiles(t *testing.T) {
	ctx := context.Background()

	// ------------------------------------- ENGINE SETUP ----------------------------------------------
	enclaveCtx, stopEnclaveFunc, _, err := test_helpers.CreateEnclave(t, ctx, testName, isPartitioningEnabled)
	require.NoError(t, err, "An error occurred creating an enclave")
	defer stopEnclaveFunc()

	// ------------------------------------- TEST SETUP ----------------------------------------------
	containerConfigSupplier := getContainerConfigSupplier()

	serviceCtx, err := enclaveCtx.AddService(testService, containerConfigSupplier)
	require.NoError(t, err, "An error occurred adding the file server service")

	// ------------------------------------- TEST RUN ----------------------------------------------
	for relativeFilepath, expectedContents := range generatedFileRelPathsAndContents {
		sharedFilepath := serviceCtx.GetSharedDirectory().GetChildPath(relativeFilepath)

		catStaticFileCmd := []string{
			"cat",
			sharedFilepath.GetAbsPathOnServiceContainer(),
		}

		exitCode, logOutput, err := serviceCtx.ExecCommand(catStaticFileCmd)
		require.NoError(t, err, "An error occurred executing command '%+v' to cat the static file '%v' contents", catStaticFileCmd, relativeFilepath)
		require.Equal(
			t,
			execCommandSuccessExitCode,
			exitCode,
			"Command '%+v' to cat the static file '%v' exited with non-successful exit code '%v'",
			catStaticFileCmd,
			relativeFilepath,
			exitCode,
		)
		actualContents := logOutput
		require.Equal(
			t,
			expectedContents,
			actualContents,
			"Static file contents '%v' don't match expected static file '%v' contents '%v'",
			actualContents,
			relativeFilepath,
			expectedContents,
		)
		logrus.Infof("Static file '%v' contents were '%v' as expected", relativeFilepath, expectedContents)
	}
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		for relFilepath, contents := range generatedFileRelPathsAndContents {
			if err := generateFileInServiceContainer(relFilepath, contents, sharedDirectory); err != nil {
				return nil, stacktrace.Propagate(err, "An error occurred generating file with relative filepath '%v'", relFilepath)
			}
		}

		// We sleep because the only function of this container is to test Docker executing a command while it's running
		// NOTE: We could just as easily combine this into a single array (rather than splitting between ENTRYPOINT and CMD
		// args), but this provides a nice little regression test of the ENTRYPOINT overriding
		entrypointArgs := []string{"sleep"}
		cmdArgs := []string{"30"}

		containerConfig := services.NewContainerConfigBuilder(
			dockerImage,
		).WithEntrypointOverride(
			entrypointArgs,
		).WithCmdOverride(
			cmdArgs,
		).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}


func generateFileInServiceContainer(relativePath string, contents string, sharedDirectory *services.SharedPath) error {
	sharedFilepath := sharedDirectory.GetChildPath(relativePath)
	absFilepathOnThisContainer := sharedFilepath.GetAbsPathOnThisContainer()
	if err := ioutil.WriteFile(sharedFilepath.GetAbsPathOnThisContainer(), []byte(contents), generatedFilePermBits); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred writing contents '%v' to file '%v' with perms '%v'",
			contents,
			absFilepathOnThisContainer,
			generatedFilePermBits,
		)
	}
	return nil
}

