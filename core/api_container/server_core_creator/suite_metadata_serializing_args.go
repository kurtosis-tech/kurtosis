/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package server_core_creator

// Fields are public for JSON de/serialization
type SuiteMetadataSerializingArgs struct {
	// The filepath, RELATIVE to the suite execution volume root, where the serialized suite metadata
	//  should be written
	SuiteMetadataRelativeFilepath string	`json:"suiteMetadataRelativeFilepath"`
}
