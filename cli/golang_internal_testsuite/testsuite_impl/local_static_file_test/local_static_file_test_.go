/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package local_static_file_test

import (
	"github.com/kurtosis-tech/kurtosis-cli/golang_internal_testsuite/static_files_consts"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/kurtosis-testsuite-api-lib/golang/lib/testsuite"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
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
	builder.WithSetupTimeoutSeconds(
		60,
	).WithRunTimeoutSeconds(
		60,
	)
}

func (l LocalStaticFileTest) Setup(networkCtx *networks.NetworkContext) (networks.Network, error) {

	containerConfigSupplier := getContainerConfigSupplier()

	_, _, err := networkCtx.AddService(testService, containerConfigSupplier)
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

	expectedTestFilesContent := []string{expectedTestFile1Contents, expectedTestFile2Contents}
	for staticFileNameKey, staticFileName := range static_files_consts.StaticFilesNames {
		testFileFilePath := serviceCtx.GetSharedDirectory().GetChildPath(staticFileName)

		catStaticFileCmd := []string{
			"cat",
			testFileFilePath.GetAbsPathOnServiceContainer(),
		}

		exitCode, logOutput, err := serviceCtx.ExecCommand(catStaticFileCmd)
		if err != nil {
			return stacktrace.Propagate(err, "An error occurred executing command '%+v' to cat the static file '%v' contents", catStaticFileCmd, staticFileName)
		}
		if exitCode != execCommandSuccessExitCode {
			return stacktrace.NewError("Command '%+v' to cat the static file '%v' exited with non-successful exit code '%v'", catStaticFileCmd, staticFileName, exitCode)
		}
		fileContents := logOutput
		if fileContents != expectedTestFilesContent[staticFileNameKey] {
			return stacktrace.NewError("Static file contents '%v' don't match expected static file '%v' contents '%v'", fileContents, staticFileName, expectedTestFilesContent[staticFileNameKey])
		}
		logrus.Infof("Static file '%v' contents were '%v' as expected", staticFileName, expectedTestFilesContent[staticFileNameKey])
	}

	return nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func getContainerConfigSupplier() func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {
	containerConfigSupplier  := func(ipAddr string, sharedDirectory *services.SharedPath) (*services.ContainerConfig, error) {

		//Copy static files from the static_files folder in testsuite container to the service's folder in the service container
		if err := copyStaticFilesInServiceContainer(static_files_consts.StaticFilesNames, static_files_consts.StaticFilesDirpathOnTestsuiteContainer, sharedDirectory); err != nil{
			return nil, stacktrace.Propagate(err, "An error occurred copying static files into the service's folder in the service container")
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

func copyStaticFilesInServiceContainer(staticFilesNames []string, staticFilesFolder string, sharedDirectory *services.SharedPath) error {
	for _, staticFileName := range staticFilesNames {
		if err := copyStaticFileInServiceContainer(staticFileName, staticFilesFolder, sharedDirectory); err != nil {
			return stacktrace.Propagate(err, "An error occurred copying file with filename '%v' to service's folder in service container", staticFileName)
		}
	}
	return nil
}

func copyStaticFileInServiceContainer(staticFileName string, staticFilesFolder string,sharedDirectory *services.SharedPath) error {
	testStaticFileFilePath := sharedDirectory.GetChildPath(staticFileName)

	testStaticFilepath := filepath.Join(staticFilesFolder, staticFileName)

	testStaticFileContent, err := ioutil.ReadFile(testStaticFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred reading file from '%v'", testStaticFilepath)
	}

	err = ioutil.WriteFile(testStaticFileFilePath.GetAbsPathOnThisContainer(), testStaticFileContent, os.ModePerm)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing file '%v'", testStaticFileFilePath.GetAbsPathOnThisContainer())
	}
	return nil
}
