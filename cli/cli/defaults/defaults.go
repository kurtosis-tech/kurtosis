/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

import (
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/logs_aggregator"
	"github.com/sirupsen/logrus"
)

const (
	// If this version is passed to the engine, the engine will use its default version
	DefaultAPIContainerVersion = ""
	// TODO perhaps move this to the metrics library
	SendMetricsByDefault = true

	// engine-enclave-pool-size = 0 means that enclave pool feat will be disabled
	DefaultEngineEnclavePoolSize uint8 = 0

	// This is the persistent flag key used, accroos all the CLI commands, to determine wheter to run in debug mode
	DebugModeFlagKey                             = "debug-mode"
	DefaultEnableDebugMode                       = false
	DefaultKurtosisContainerDebugImageNameSuffix = "debug"

	DefaultGitHubAuthTokenOverride = ""

	DefaultDomain = ""

	DefaultLogRetentionPeriod = "168h"
)

var DefaultSinks = logs_aggregator.Sinks{}
var DefaultApiContainerLogLevel = logrus.DebugLevel
var DefaultEngineLogLevel = logrus.DebugLevel
