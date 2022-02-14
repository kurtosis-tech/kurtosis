/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/sirupsen/logrus"
)

const (
	kurtosisDockerOrg = "kurtosistech"

	DefaultJavascriptReplImage = kurtosisDockerOrg + "/javascript-interactive-repl:" + kurtosis_cli_version.KurtosisCLIVersion

	// If this version is passed to the engine, the engine will use its default version
	DefaultAPIContainerVersion = ""
)
var DefaultApiContainerLogLevel = logrus.DebugLevel
var DefaultEngineLogLevel = logrus.DebugLevel