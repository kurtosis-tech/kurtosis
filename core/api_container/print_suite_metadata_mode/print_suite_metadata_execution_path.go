/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package print_suite_metadata_mode

type PrintSuiteMetadataExecutionPath struct {
	args PrintSuiteMetadataExecutionArgs
}

func NewPrintSuiteMetadataExecutionPath(args PrintSuiteMetadataExecutionArgs) *PrintSuiteMetadataExecutionPath {
	return &PrintSuiteMetadataExecutionPath{args: args}
}

func (p PrintSuiteMetadataExecutionPath) Execute() error {
	panic("implement me")
}

