/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package print_suite_metadata_mode

// Fields are public for JSON de/serialization
type PrintSuiteMetadataArgs struct {
	// The filepath, RELATIVE to the suite execution volume root, where the suite metadata
	//  will be written
	SuiteMetadataRelativeFilepath string
}

// TODO Constructor
