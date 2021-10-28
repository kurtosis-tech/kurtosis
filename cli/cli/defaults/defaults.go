/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

import (
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/kurtosis-tech/kurtosis-core/commons/kurtosis_core_version"
	"github.com/kurtosis-tech/kurtosis-engine-server/engine/kurtosis_engine_server_version"
	"github.com/sirupsen/logrus"
)

const (
	kurtosisDockerOrg = "kurtosistech"

	DefaultJavascriptReplImage = kurtosisDockerOrg + "/javascript-interactive-repl:" + kurtosis_cli_version.KurtosisCLIVersion

	DefaultApiContainerImage = kurtosisDockerOrg + "/kurtosis-core_api:" + kurtosis_core_version.KurtosisCoreVersion

	DefaultEngineImage = kurtosisDockerOrg + "/kurtosis-engine-server:" + kurtosis_engine_server_version.KurtosisEngineServerVersion
)
var DefaultEngineLogLevel = logrus.InfoLevel
