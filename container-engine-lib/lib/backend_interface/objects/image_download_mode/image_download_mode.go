/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package image_download_mode

import "github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"

const (
	Always  = "always"
	Missing = "missing"
)

type ImageDownloadMode string

func FromAPI(api_mode kurtosis_core_rpc_api_bindings.ImageDownloadMode) ImageDownloadMode {
	switch kurtosis_core_rpc_api_bindings.ImageDownloadMode_name[int32(api_mode)] {
	case "always":
		return Always
	case "missing":
		return Missing
	default:
		panic("Invalid value from API")
	}
}
