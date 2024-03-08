/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package image_download_mode

import "fmt"

const (
	ImageDownloadMode_Always  = iota
	ImageDownloadMode_Missing = iota
)

type ImageDownloadMode int

func (mode ImageDownloadMode) String() string {
	switch mode {
	case ImageDownloadMode_Always:
		return "always"
	case ImageDownloadMode_Missing:
		return "missing"
	default:
		return fmt.Sprintf("ImageDownloadMode_value_%d", int(mode))
	}
}
