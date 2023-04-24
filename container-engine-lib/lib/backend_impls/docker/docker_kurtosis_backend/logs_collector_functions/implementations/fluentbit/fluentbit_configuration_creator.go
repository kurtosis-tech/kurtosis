package fluentbit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"text/template"
	"time"
)

const (
	// We use this image and version because we already are using this in other projects so there is a high probability
	// that the image is in the local machine's cache
	configuratorContainerImage = "alpine:3.17"
	configuratorContainerName  = "kurtosis-fluentbit-configurator"

	shBinaryFilepath = "/bin/sh"
	shCmdFlag        = "-c"
	printfCmdName    = "printf"

	configFileCreationSuccessExitCode = 0

	configFileCreationCmdMaxRetries     = 2
	configFileCreationCmdDelayInRetries = 200 * time.Millisecond

	sleepSeconds = 1800
)

type fluentbitConfigurationCreator struct {
	config *FluentbitConfig
}

func newFluentbitConfigurationCreator(config *FluentbitConfig) *fluentbitConfigurationCreator {
	return &fluentbitConfigurationCreator{config: config}
}

func (fluent *fluentbitConfigurationCreator) CreateConfiguration(
	ctx context.Context,
	targetNetworkId string,
	volumeName string,
	dockerManager *docker_manager.DockerManager,
) error {

	entrypointArgs := []string{
		shBinaryFilepath,
		shCmdFlag,
		fmt.Sprintf("sleep %v", sleepSeconds),
	}

	volumeMounts := map[string]string{
		volumeName: configDirpathInContainer,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		configuratorContainerImage,
		configuratorContainerName,
		targetNetworkId,
	).WithEntrypointArgs(
		entrypointArgs,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the Fluentbit configurator container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if err = dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the Fluentbit configurator container with container ID '%v' didn't complete successfully so we "+
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

func (fluent *fluentbitConfigurationCreator) createFluentbitConfigFileInVolume(
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
		"The Fluentbit config file creation didn't return success (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}

func (fluent *fluentbitConfigurationCreator) getConfigFileContent() (string, error) {

	cngFileTemplate, err := template.New(configFileTemplateName).Parse(configFileTemplate)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing Fluentbit config template '%v'", configFileTemplate)
	}

	templateStrBuffer := &bytes.Buffer{}

	if err := cngFileTemplate.Execute(templateStrBuffer, fluent.config); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing the Fluentbit config file template")
	}

	templateStr := templateStrBuffer.String()

	return templateStr, nil
}
