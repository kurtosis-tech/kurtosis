package fluentbit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	//We use this image and version because we already are using this in other projects so there is a high probability
	//that the image is already downloaded in the local machine
	configuratorContainerImage = "alpine:3.12.4"
	configuratorContainerName = "kurtosis-fluentbit-configurator"

	shBinaryFilepath = "/bin/sh"
	shCmdFlag        = "-c"
	printfCmdName    = "printf"

	configFileCreationSuccessExitCode = 0

	configFileCreationCmdMaxRetries     = 2
	configFileCreationCmdDelayInRetries = 200 * time.Millisecond
)

func (fluent *FluentbitContainerConfigProvider) runConfigurator(
	ctx context.Context,
	targetNetworkId string,
	volumeMounts map[string]string,
	dockerManager *docker_manager.DockerManager,
) error {

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		configuratorContainerImage,
		configuratorContainerName,
		targetNetworkId,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Fluenbit configurator container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the Fluenbit configurator container with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the container with ID '%v'!!!!!!", containerId)
		}
	}()

	if err := fluent.createFluentbitConfigFileInVolume(
		ctx,
		dockerManager,
		containerId,
		configFileCreationCmdMaxRetries,
		configFileCreationCmdDelayInRetries,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating the Fluentbit config file into the volume")
	}

	return nil
}

func (fluent *FluentbitContainerConfigProvider)  createFluentbitConfigFileInVolume(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	containerId string,
	maxRetries uint,
	timeBetweenRetries time.Duration,
) error {

	configFileContentStr, err := fluent.getConfigFileContent()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the Fluentbit config file content")
	}

	commandStr := fmt.Sprintf(
		"%v '%v' > %v",
		printfCmdName,
		configFileContentStr,
		configFilepathInContainer,
	)

	execCmd := []string{
		shBinaryFilepath,
		shCmdFlag,
		commandStr,
	}
	for i := uint(0); i < maxRetries; i++ {
		outputBuffer := &bytes.Buffer{}
		exitCode, err := dockerManager.RunExecCommand(ctx, containerId, execCmd, outputBuffer)
		if err == nil {
			if exitCode == configFileCreationSuccessExitCode {
				logrus.Debugf("The Fluentbit config file with content '%v' was successfully added into the volume", configFileContentStr)
				return nil
			}
			logrus.Debugf(
				"Fluentbit config file creation command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				configFileCreationSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"Fluentbit config file creation command '%v' experienced a Docker error:\n%v",
				commandStr,
				err,
			)
		}

		// Tiny optimization to not sleep if we're not going to run the loop again
		if i < maxRetries {
			time.Sleep(timeBetweenRetries)
		}
	}

	return stacktrace.NewError(
		"The Fluentbit config file creation didn't success (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}
