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
)

var DefaultApiContainerLogLevel = logrus.DebugLevel
var DefaultEngineLogLevel = logrus.DebugLevel
