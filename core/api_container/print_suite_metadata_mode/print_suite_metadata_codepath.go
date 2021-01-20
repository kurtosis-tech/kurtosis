/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package print_suite_metadata_mode

type PrintSuiteMetadataArgs struct {
	// The filepath, RELATIVE to the suite execution volume root, where the suite metadata
	//  will be written
	SuiteMetadataRelativeFilepath string
}

type PrintSuiteMetadataCodepath struct {
	args PrintSuiteMetadataArgs
}

func NewPrintSuiteMetadataCodepath(args PrintSuiteMetadataArgs) *PrintSuiteMetadataCodepath {
	return &PrintSuiteMetadataCodepath{args: args}
}

func (p PrintSuiteMetadataCodepath) Execute() (int, error) {
	panic("implement me")
}

