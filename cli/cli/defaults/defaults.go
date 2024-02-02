/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

import (
	"github.com/sirupsen/logrus"
)

const (
	// If this version is passed to the engine, the engine will use its default version
	DefaultAPIContainerVersion = ""
	// TODO perhaps move this to the metrics library
	SendMetricsByDefault = true

	// engine-enclave-pool-size = 0 means that enclave pool feat will be disabled
	DefaultEngineEnclavePoolSize uint8 = 0

	DefaultGitAuthTokenOverride = ""
)

var DefaultApiContainerLogLevel = logrus.DebugLevel
var DefaultEngineLogLevel = logrus.DebugLevel
