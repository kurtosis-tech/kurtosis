/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package print_suite_metadata_mode

type PrintSuiteMetadataCodepath struct {
	args PrintSuiteMetadataArgs
}

func NewPrintSuiteMetadataCodepath(args PrintSuiteMetadataArgs) *PrintSuiteMetadataCodepath {
	return &PrintSuiteMetadataCodepath{args: args}
}

func (p PrintSuiteMetadataCodepath) Execute() (int, error) {
	panic("implement me")
}

