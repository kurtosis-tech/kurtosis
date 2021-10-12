/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package defaults

const (
	// TODO Right now this needs to be manually kept in sync with the go.mod dependency. Ideally, we'd have some way
	//  to update both of these at the same time!!
	apiContainerVersion = "1.23.2"

	// Own version, so that we can start the proper Javascript REPL Docker image
	ownVersion = "0.3.0"

	// TODO These defaults aren't great - it will just start a Kurtosis interactive with the latest
	//  of both images, which may or may not be compatible - what we really need is a system that
	//  detects what version of the API container/REPL to start based off the Kurt Core API version
	// TODO It's also not great that these are hardcoded - they should be hooked into the build system,
	//  to guarantee that they're compatible with each other
	DefaultApiContainerImage = "kurtosistech/kurtosis-core_api:" + apiContainerVersion
	DefaultJavascriptReplImage = "kurtosistech/javascript-interactive-repl:" + ownVersion
)
