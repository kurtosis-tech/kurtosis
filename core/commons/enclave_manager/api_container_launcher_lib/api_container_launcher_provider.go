/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package lib

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/api_container_launcher_lib/api_container_launcher"
	"github.com/kurtosis-tech/kurtosis/commons/enclave_manager/api_container_launcher_lib/api_versions"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

// The entrypoint for getting API container launchers
func GetAPIContainerLauncherForLaunchAPIVersion(
	launchApiVersion uint,
	dockerManager *docker_manager.DockerManager,
	log *logrus.Logger,
	containerImage string,
	listenPort uint,
	listenProtocol string,
	logLevel logrus.Level,
) (api_container_launcher.APIContainerLauncher, error) {
	if int(launchApiVersion) >= len(api_versions.PerAPIVersionLauncherFactories) {
		return nil, stacktrace.NewError("Launch API version '%v' is newer than any version we know about", launchApiVersion)
	}
	factory := api_versions.PerAPIVersionLauncherFactories[launchApiVersion]
	launcher := factory(dockerManager, log, containerImage, listenPort, listenProtocol, logLevel)
	return launcher, nil
}
