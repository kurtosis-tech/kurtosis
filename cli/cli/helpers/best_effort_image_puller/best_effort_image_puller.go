/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package best_effort_image_puller

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_impls/docker/docker_manager"
	"github.com/sirupsen/logrus"
)

func PullImageBestEffort(ctx context.Context, dockerManager *docker_manager.DockerManager, image string) {
	if err := dockerManager.PullImage(ctx, image); err != nil {
		logrus.Warnf("Failed to pull the latest version of image '%v'; you may be running an out-of-date version", image)
	}
}
