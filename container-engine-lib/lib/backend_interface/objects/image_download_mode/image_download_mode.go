/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package image_download_mode

const (
	Never   ImageDownloadMode = "never"
	Always                    = "always"
	Missing                   = "missing"
)

type ImageDownloadMode string
