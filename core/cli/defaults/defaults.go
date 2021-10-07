/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

const (
	// TODO These defaults aren't great - it will just start a Kurtosis interactive with the latest
	//  of both images, which may or may not be compatible - what we really need is a system that
	//  detects what version of the API container/REPL to start based off the Kurt Core API version
	// TODO It's also not great that these are hardcoded - they should be hooked into the build system,
	//  to guarantee that they're compatible with each other
	DefaultApiContainerImage = "kurtosistech/kurtosis-core_api"
	DefaultJavascriptReplImage = "kurtosistech/javascript-interactive-repl"
)
