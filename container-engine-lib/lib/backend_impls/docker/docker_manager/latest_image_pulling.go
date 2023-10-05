/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package docker_manager

const (
	Never   LatestImagePulling = "never"
	Always                     = "always"
	Missing                    = "missing"
)

type LatestImagePulling string
