package host_machine_directories

import (
	"github.com/adrg/xdg"
	"github.com/kurtosis-tech/stacktrace"
	"path"
)

const (
	applicationDirname = "kurtosis"

	kurtosisConfigYAMLFilename = "kurtosis-config.yml"

	kurtosisClusterSettingFilename = "cluster-setting"

	latestCLIReleaseVersionCacheFilename = "latest-cli-release-version-cache"

	metricsUserIDFilename = "metrics-user-id"

	userSendMetricsElection = "user-send-metrics-election"

	LastPesteredUserAboutOldVersionFilename = "last-pestered-user-about-old-version"

	portalBinaryFilename  = "kurtosis-portal"
	portalLogFilename     = "kurtosis-portal.log"
	portalVersionFilename = "kurtosis-portal.version"
	portalPidFilename     = "kurtosis-portal.pid"

	// ------------ Names of dirs inside Kurtosis directory --------------
	engineDataDirname      = "engine-data"
	portalSubDirname       = "portal"
	kurtosisCliLogsDirname = "cli"
)

// TODO after 2022-07-08, when we're confident nobody is using engines without engine data directories anymore,
//
//	add a step to 'engine stop' that will delete this directory on the user's machine if it exists
//
// Gets the engine data directory on the host machine, and ensures the path exists
func GetEngineDataDirpath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(engineDataDirname)
	engineDataDirpath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting engine data dirpath from relative path '%v'", xdgRelFilepath)
	}
	return engineDataDirpath, nil
}

// Get the yaml filepath where the Kurtosis configs are saved
func GetKurtosisConfigYAMLFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(kurtosisConfigYAMLFilename)
	kurtosisConfigYAMLFilepath, err := xdg.ConfigFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis config YAML filepath from relative path '%v'", xdgRelFilepath)
	}
	return kurtosisConfigYAMLFilepath, nil
}

// Get the cluster setting filepath where the users' cluster selection setting is saved
func GetKurtosisClusterSettingFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(kurtosisClusterSettingFilename)
	kurtosisClusterSettingFilepath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the Kurtosis cluster setting filepath from relative path '%v'", xdgRelFilepath)
	}
	return kurtosisClusterSettingFilepath, nil
}

func GetMetricsUserIdFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(metricsUserIDFilename)
	filepath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath from relative path '%v'", xdgRelFilepath)
	}
	return filepath, nil
}

func GetUserSendMetricsElectionFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(userSendMetricsElection)
	filepath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the user-send-metrics-election filepath from relative path '%v'", xdgRelFilepath)
	}
	return filepath, nil
}

func GetLatestCLIReleaseVersionCacheFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(latestCLIReleaseVersionCacheFilename)
	latestCLIReleaseVersionCacheFilepath, err := xdg.CacheFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the latest release version cache filepath from relative path '%v'", xdgRelFilepath)
	}
	return latestCLIReleaseVersionCacheFilepath, nil
}

func GetLastPesteredUserAboutOldVersionsFilepath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForXDG(LastPesteredUserAboutOldVersionFilename)
	lastPesteredUserForOldVersionsFilePath, err := xdg.CacheFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the last pestered user about old version file path using '%v'", xdgRelFilepath)
	}
	return lastPesteredUserForOldVersionsFilePath, nil
}

func GetKurtosisCliLogsFileDirPath(fileName string) (string, error) {
	xdgRelDirPath := getRelativeFilePathForKurtosisCliLogs()
	kurtosisCliLogFilePath, err := xdg.DataFile(path.Join(xdgRelDirPath, fileName))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the kurtosis cli logs file path using '%v'", kurtosisCliLogFilePath)
	}
	return kurtosisCliLogFilePath, nil
}

func GetPortalBinaryFilePath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForPortalForXDG(portalBinaryFilename)
	portalBinaryFilePath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting Kurtosis Portal binary file path using '%s'", xdgRelFilepath)
	}
	return portalBinaryFilePath, nil
}

func GetPortalLogFilePath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForPortalForXDG(portalLogFilename)
	portalBinaryFilePath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting Kurtosis Portal log file path using '%s'", xdgRelFilepath)
	}
	return portalBinaryFilePath, nil
}

func GetPortalVersionFilePath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForPortalForXDG(portalVersionFilename)
	portalVersionFilePath, err := xdg.DataFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting Kurtosis Portal version file path using '%s'", xdgRelFilepath)
	}
	return portalVersionFilePath, nil
}

func GetPortalPidFilePath() (string, error) {
	xdgRelFilepath := getRelativeFilepathForPortalForXDG(portalPidFilename)
	portalPidFilePath, err := xdg.StateFile(xdgRelFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting Kurtosis Portal PID file path using '%s'", xdgRelFilepath)
	}
	return portalPidFilePath, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
// Joins the "kurtosis" app directory in front of whichever filepath
func getRelativeFilepathForXDG(filepathRelativeToKurtosisDir string) string {
	return path.Join(applicationDirname, filepathRelativeToKurtosisDir)
}

func getRelativeFilepathForPortalForXDG(filepathRelativeToKurtosisPortalDir string) string {
	return path.Join(applicationDirname, portalSubDirname, filepathRelativeToKurtosisPortalDir)
}

func getRelativeFilePathForKurtosisCliLogs() string {
	return path.Join(applicationDirname, kurtosisCliLogsDirname)
}
