/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
)

type KurtosisBackendConfigSupplier interface {
	// Private because only the launcher should call it
	getKurtosisBackendConfig() (args.KurtosisBackendType, interface{})
}
