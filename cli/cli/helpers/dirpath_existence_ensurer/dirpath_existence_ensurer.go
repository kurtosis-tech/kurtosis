package dirpath_existence_ensurer

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

// Ensures that the given directory exists
func EnsureDirpathExists(dirpath string) error {
	if _, err := os.Stat(dirpath); os.IsNotExist(err) {
		if err := os.Mkdir(dirpath, 0777); err != nil {
			return stacktrace.Propagate(
				err,
				"Kurtosis directory '%v' didn't exist, and an error occurred trying to create it",
				dirpath,
			)
		}
	}
	return nil
}
