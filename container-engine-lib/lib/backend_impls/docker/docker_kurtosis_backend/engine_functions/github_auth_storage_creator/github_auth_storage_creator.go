package github_auth_storage_creator

import (
	"bytes"
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_kurtosis_backend/consts"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

const (
	// We use this image and version because we already are using this in other projects so there is a high probability
	// that the image is in the local machine's cache
	creatorContainerImage = "alpine:3.17"
	creatorContainerName  = "kurtosis-github-auth-storage-creator"

	shBinaryFilepath = "/bin/sh"
	shCmdFlag        = "-c"
	printfCmdName    = "printf"

	authStorageCreationSuccessExitCode = 0

	authStorageCreationCmdMaxRetries     = 2
	authStorageCreationCmdDelayInRetries = 200 * time.Millisecond

	sleepSeconds = 1800
)

type GitHubAuthStorageCreator struct {
	token string
}

func NewGitHubAuthStorageCreator(token string) *GitHubAuthStorageCreator {
	return &GitHubAuthStorageCreator{token: token}
}

func (creator *GitHubAuthStorageCreator) CreateGitHubAuthStorage(
	ctx context.Context,
	targetNetworkId string,
	volumeName string,
	githubAuthStorageDirPath string,
	dockerManager *docker_manager.DockerManager,
) error {
	entrypointArgs := []string{
		shBinaryFilepath,
		shCmdFlag,
		fmt.Sprintf("sleep %v", sleepSeconds),
	}

	volumeMounts := map[string]string{
		volumeName: githubAuthStorageDirPath,
	}

	createAndStartArgs := docker_manager.NewCreateAndStartContainerArgsBuilder(
		creatorContainerImage,
		creatorContainerName,
		targetNetworkId,
	).WithEntrypointArgs(
		entrypointArgs,
	).WithVolumeMounts(
		volumeMounts,
	).Build()

	containerId, _, err := dockerManager.CreateAndStartContainer(ctx, createAndStartArgs)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred starting the GitHub Auth Storage Creator container with these args '%+v'", createAndStartArgs)
	}
	//The killing step has to be executed always in the success and also in the failed case
	defer func() {
		if err = dockerManager.RemoveContainer(context.Background(), containerId); err != nil {
			logrus.Errorf(
				"Launching the GitHub Auth Storage Creator container with container ID '%v' didn't complete successfully so we "+
					"tried to remove the container we started, but doing so exited with an error:\n%v",
				containerId,
				err)
			logrus.Errorf("ACTION REQUIRED: You'll need to manually remove the container with ID '%v'!!!!!!", containerId)
		}
	}()

	if err := creator.storeTokenInVolume(
		ctx,
		dockerManager,
		containerId,
		authStorageCreationCmdMaxRetries,
		authStorageCreationCmdDelayInRetries,
		githubAuthStorageDirPath,
	); err != nil {
		return stacktrace.Propagate(err, "An error occurred creating  GitHub auth storage in volume.")
	}

	return nil
}

// GetBestEffortGitHubAuthToken Returns empty string if no token found in [githubAuthTokenFile] or [githubAuthTokenFile] doesn't exist
func GetBestEffortGitHubAuthToken() string {
	tokenBytes, err := os.ReadFile(path.Join(consts.GitHubAuthStorageDirPath, consts.GithubAuthStorageToken))
	if err != nil {
		return ""
	}
	return string(tokenBytes)
}

func (creator *GitHubAuthStorageCreator) storeTokenInVolume(
	ctx context.Context,
	dockerManager *docker_manager.DockerManager,
	containerId string,
	maxRetries uint,
	timeBetweenRetries time.Duration,
	githubAuthStorageDirPath string,
) error {
	commandStr := fmt.Sprintf(
		"%v '%v' > %v",
		printfCmdName,
		creator.token,
		fmt.Sprintf("%s/%s", githubAuthStorageDirPath, consts.GithubAuthStorageToken),
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
			if exitCode == authStorageCreationSuccessExitCode {
				logrus.Debugf("The GitHub auth token was successfully added into the volume.")
				return nil
			}
			logrus.Debugf(
				"GitHub auth storage creation command '%v' returned without a Docker error, but exited with non-%v exit code '%v' and logs:\n%v",
				commandStr,
				authStorageCreationSuccessExitCode,
				exitCode,
				outputBuffer.String(),
			)
		} else {
			logrus.Debugf(
				"GitHub auth storage creation command '%v' experienced a Docker error:\n%v",
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
		"The GitHub auth storage creation didn't return success (as measured by the command '%v') even after retrying %v times with %v between retries",
		commandStr,
		maxRetries,
		timeBetweenRetries,
	)
}
