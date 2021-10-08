/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher_lib

import (
	"github.com/kurtosis-tech/container-engine-lib/lib/docker_manager"
	api_container_launcher2 "github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib/api_container_launcher"
	api_versions2 "github.com/kurtosis-tech/kurtosis-core/commons/api_container_launcher_lib/api_versions"
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
) (api_container_launcher2.APIContainerLauncher, error) {
	if int(launchApiVersion) >= len(api_versions2.PerAPIVersionLauncherFactories) {
		return nil, stacktrace.NewError("Launch API version '%v' is newer than any version we know about", launchApiVersion)
	}
	factory := api_versions2.PerAPIVersionLauncherFactories[launchApiVersion]
	launcher := factory(dockerManager, log, containerImage, listenPort, listenProtocol, logLevel)
	return launcher, nil
}
