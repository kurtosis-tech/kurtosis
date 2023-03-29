package portal_manager

import (
	"github.com/kurtosis-tech/stacktrace"
	"os"
)

// CreateFileIfNecessary creates the file if it doesn't already exist. If it does, do nothing.
// Return true if the file was created, false if it existed already, and error if something went wrong
func CreateFileIfNecessary(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err != nil {
		if !os.IsNotExist(err) {
			return false, stacktrace.Propagate(err, "Error checking if file already exists or not")
		}
		createdFile, creationErr := os.Create(filePath)
		if creationErr != nil {
			return false, stacktrace.Propagate(creationErr, "Unable to create file at '%s'", filePath)
		}
		defer createdFile.Close()
		return true, nil
	}
	// file exists, all good
	return false, nil
}
