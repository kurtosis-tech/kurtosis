/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"github.com/palantir/stacktrace"
	"os"
)

func ensureDirpathExists(absoluteDirpath string) error {
	if _, err := os.Stat(absoluteDirpath); os.IsNotExist(err) {
		if err := os.Mkdir(absoluteDirpath, os.ModeDir); err != nil {
			return stacktrace.Propagate(
				err,
				"Directory '%v' in the suite execution volume didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	return nil
}
