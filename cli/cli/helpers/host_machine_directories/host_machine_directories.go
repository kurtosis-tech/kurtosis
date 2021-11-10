package host_machine_directories

import (
	"github.com/adrg/xdg"
	"github.com/palantir/stacktrace"
	"path"
)

const (
	applicationDirname = "kurtosis"

	sessionCacheFilename = "session-cache"

	latestCLIReleaseVersionCacheFilename = "latest-release-version-cache"

	// ------------ Names of dirs inside Kurtosis directory --------------
	engineDataDirname = "engine-data"
)

// Gets the engine data directory on the host machine, and ensures the path exists
func GetEngineDataDirpath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(engineDataDirname)
	engineDataDirpath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting engine data dirpath from relative path '%v'", xdgRelFilepath)
	}
	return engineDataDirpath, nil
}

// TODO Plug this into the 'test' auth framework in a different PR
func GetSessionCacheFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(sessionCacheFilename)
	sessionCacheFilepath, err := xdg.CacheFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the session cache filepath from relative path '%v'", xdgRelFilepath)
	}
	return sessionCacheFilepath, nil
}

func GetLatestCLIReleaseVersionCacheFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(latestCLIReleaseVersionCacheFilename)
	latestCLIReleaseVersionCacheFilepath, err := xdg.CacheFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the latest release version cache filepath from relative path '%v'", xdgRelFilepath)
	}
	return latestCLIReleaseVersionCacheFilepath, nil
}


// ====================================================================================================
//                                     Private Helper Functions
// ====================================================================================================
// Joins the "kurtosis" app directory in front of whichever filepath
func getRelativeFilepathForXDG(filepathRelativeToKurtosisDir string) string {
	return path.Join(applicationDirname, filepathRelativeToKurtosisDir)
}