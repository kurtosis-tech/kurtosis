/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package static_files_consts

import (
	"github.com/kurtosis-tech/kurtosis-client/golang/lib/services"
	"path"
)

const (
	// Directory where static files live inside the testsuite container
	staticFilesDirpathOnTestsuiteContainer = "/static-files"

	TestStaticFile1ID      services.StaticFileID = "test-static-file1"
	TestStaticFile2ID	   services.StaticFileID = "test-static-file2"

	testStaticFile1Filename = "test-static-file1.txt"
	testStaticFile2Filename = "test-static-file2.txt"
)

var StaticFileFilepaths = map[services.StaticFileID]string{
	TestStaticFile1ID: path.Join(staticFilesDirpathOnTestsuiteContainer, testStaticFile1Filename),
	TestStaticFile2ID: path.Join(staticFilesDirpathOnTestsuiteContainer, testStaticFile2Filename),
}
