/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"github.com/palantir/stacktrace"
	"os"
)

func ensureDirpathExists(absoluteDirpath string) error {
	if _, err := os.Stat(absoluteDirpath); os.IsNotExist(err) {
		if err := os.Mkdir(absoluteDirpath, 0777); err != nil {
			return stacktrace.Propagate(
				err,
				"Directory '%v' in the enclave data volume didn't exist, and an error occurred trying to create it",
				absoluteDirpath)
		}
	}
	return nil
}
