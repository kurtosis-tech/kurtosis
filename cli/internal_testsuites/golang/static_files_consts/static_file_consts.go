/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package static_files_consts

const (
	// Directory where static files live inside the testsuite container
	StaticFilesDirpathOnTestsuiteContainer = "/static-files"

	testStaticFile1Filename = "test-static-file1.txt"
	testStaticFile2Filename = "test-static-file2.txt"
)

var StaticFilesNames = []string{testStaticFile1Filename, testStaticFile2Filename}
