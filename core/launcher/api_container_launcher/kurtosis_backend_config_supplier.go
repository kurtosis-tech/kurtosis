/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package api_container_launcher

import (
	"github.com/kurtosis-tech/kurtosis/core/launcher/args"
)

type KurtosisBackendConfigSupplier interface {
	// Private because only the launcher should call it
	getKurtosisBackendConfig() (args.KurtosisBackendType, interface{})
}
