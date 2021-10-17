/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

import "github.com/kurtosis-tech/kurtosis-core/commons/kurtosis_core_version"

const (
	// vvvv WARNING: DO NOT MODIFY THIS MANUALLY! IT WILL BE UPDATED DURING THE RELEASE PROCESS vvvvv
	// Own version, so that we can start the proper Javascript REPL Docker image
	ownVersion = "0.4.1"
	// ^^^^ WARNING: DO NOT MODIFY THIS MANUALLY! IT WILL BE UPDATED DURING THE RELEASE PROCESS ^^^^^

	DefaultJavascriptReplImage = "kurtosistech/javascript-interactive-repl:" + ownVersion

	DefaultApiContainerImage = "kurtosistech/kurtosis-core_api:" + kurtosis_core_version.KurtosisCoreVersion
)
