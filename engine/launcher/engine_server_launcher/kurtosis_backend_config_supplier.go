/*
 * Copyright (c) 2022 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package engine_server_launcher

import (
	"github.com/kurtosis-tech/kurtosis-engine-server/launcher/args"
)

type KurtosisBackendConfigSupplier interface {
	// Public because both the launcher and the enclave manager need to call this
	GetKurtosisBackendConfig() (args.KurtosisBackendType, interface{})
}
