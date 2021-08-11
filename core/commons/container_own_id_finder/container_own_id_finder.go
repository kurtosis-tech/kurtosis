/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package container_own_id_finder

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/commons/docker_manager"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	maxTimesTryingToFindOwnContainerId = 5
	timeBetweenTryingToFindOwnContainerId = 1 * time.Second
)

func GetOwnContainerIdByName(ctx context.Context, dockerManager *docker_manager.DockerManager, nameFragment string) (string, error) {
	logrus.Debugf("Getting own container ID given container name fragment '%v'...")

	// For some reason, Docker very occasionally will report 0 containers matching a name fragment even
	//  though this container definitely has the right name, so we therefore retry a couple times as a workaround
	// See: https://github.com/moby/moby/issues/42354)
	timesTried := 0
	for timesTried < maxTimesTryingToFindOwnContainerId {
		matchingIds, err := dockerManager.GetContainerIdsByName(ctx, nameFragment)
		if err != nil {
			logrus.Debugf("Got an error while trying to get own container ID: %v", err)
		} else if len(matchingIds) != 1 {
			logrus.Debugf("Expected exactly 1 container ID matching name fragment '%v' but got %v", nameFragment, len(matchingIds))
		} else {
			result := matchingIds[0]
			logrus.Debugf("Got own container ID: %v", result)
			return result, nil
		}
		timesTried = timesTried + 1
		if timesTried < maxTimesTryingToFindOwnContainerId {
			logrus.Debugf("Sleeping for %v before trying to get own container ID again...", timeBetweenTryingToFindOwnContainerId)
			time.Sleep(timeBetweenTryingToFindOwnContainerId)
		}
	}
	return "", stacktrace.NewError(
		"Couldn't get own container ID despite trying %v times with %v between tries",
		maxTimesTryingToFindOwnContainerId,
		timeBetweenTryingToFindOwnContainerId,
	)
}
