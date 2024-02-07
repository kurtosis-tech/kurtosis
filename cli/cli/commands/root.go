/*
 * Copyright (c) 2021 - present Kurtosis Technologies Inc.
 * All Rights Reserved.
 */

package commands

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_str_consts"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/analytics"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/clean"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/cloud"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/cluster"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/config"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/discord"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/docs"
	kurtosisdump "github.com/kurtosis-tech/kurtosis/cli/cli/commands/dump"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/enclave"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/engine"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/feedback"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/files"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/gateway"
	_import "github.com/kurtosis-tech/kurtosis/cli/cli/commands/import"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/kurtosis_context"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/lint"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/lsp"
	_package "github.com/kurtosis-tech/kurtosis/cli/cli/commands/package"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/port"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/portal"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/service"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/twitter"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/version"
	"github.com/kurtosis-tech/kurtosis/cli/cli/commands/web"
	"github.com/kurtosis-tech/kurtosis/cli/cli/defaults"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/logrus_log_levels"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/user_send_metrics_election"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/user_support_constants"
	"github.com/kurtosis-tech/kurtosis/kurtosis_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// !!! WARNING !!!! If you change the name of this flag, make sure to update it in the "Debugging User Issues" section of the README!!!
	cliLogLevelStrFlag = "cli-log-level"

	latestReleaseOnGitHubURL   = "https://api.github.com/repos/kurtosis-tech/kurtosis-cli-release-artifacts/releases/latest"
	acceptHttpHeaderKey        = "Accept"
	acceptHttpHeaderValue      = "application/json"
	contentTypeHttpHeaderKey   = "Content-Type"
	contentTypeHttpHeaderValue = "application/json"
	userAgentHttpHeaderKey     = "User-Agent"
	userAgentHttpHeaderValue   = "kurtosis-tech"

	latestCLIReleaseCacheFileContentSeparator       = ";"
	latestCLIReleaseCacheFileExpirationHours        = 24
	latestCLIReleaseCacheFileContentColumnsAmount   = 2
	latestCLIReleaseCacheFileContentDateIndex       = 0
	latestCLIReleaseCacheFileContentVersionIndex    = 1
	latestCLIReleaseCacheFileCreationDateTimeFormat = time.RFC3339

	frequencyToPesterUsersAboutVersions = 1 * time.Hour

	getLatestCLIReleaseCacheFilePermissions os.FileMode = 0644

	optionalSemverPrefix = "v"
)

type GitHubReleaseReponse struct {
	TagName string `json:"tag_name"`
}

var logLevelStr string
var defaultLogLevelStr = logrus.InfoLevel.String()

// RootCmd Suppressing exhaustruct requirement because this struct has ~40 properties
// nolint: exhaustruct
var RootCmd = &cobra.Command{
	Use:   command_str_consts.KurtosisCmdStr,
	Short: "A CLI for interacting with the Kurtosis engine",
	// Cobra will print usage whenever _any_ error occurs, including ones we throw in Kurtosis
	// This doesn't make sense in 99% of the cases, so just turn them off entirely
	SilenceUsage: true,
	// Cobra prints the errors itself, however, with this flag disabled it give Kurtosis control
	// and allows us to post process the error in the main.go file.
	SilenceErrors:     true,
	PersistentPreRunE: globalSetup,
}

func init() {
	RootCmd.PersistentFlags().StringVar(
		&logLevelStr,
		cliLogLevelStrFlag,
		defaultLogLevelStr,
		"Sets the level that the CLI will log at ("+strings.Join(logrus_log_levels.GetAcceptableLogLevelStrs(), "|")+")",
	)
	RootCmd.PersistentFlags().Bool(
		defaults.DebugModeFlagKey,
		defaults.DefaultEnableDebugMode,
		"Whether should enable Kurtosis in debug mode. The debug mode will use the Kurtosis container debug images version (only enabled for the engine server so far)",
	)

	RootCmd.AddCommand(analytics.AnalyticsCmd.MustGetCobraCommand())
	RootCmd.AddCommand(clean.CleanCmd.MustGetCobraCommand())
	RootCmd.AddCommand(cluster.ClusterCmd)
	RootCmd.AddCommand(kurtosis_context.ContextCmd)
	RootCmd.AddCommand(cloud.CloudCmd)
	RootCmd.AddCommand(kurtosisdump.KurtosisDump.MustGetCobraCommand())
	RootCmd.AddCommand(config.ConfigCmd)
	RootCmd.AddCommand(discord.DiscordCmd.MustGetCobraCommand())
	RootCmd.AddCommand(docs.DocsCmd.MustGetCobraCommand())
	RootCmd.AddCommand(enclave.EnclaveCmd)
	RootCmd.AddCommand(engine.EngineCmd)
	RootCmd.AddCommand(feedback.FeedbackCmd.MustGetCobraCommand())
	RootCmd.AddCommand(files.FilesCmd)
	RootCmd.AddCommand(gateway.GatewayCmd)
	RootCmd.AddCommand(lsp.NewLspCommand())
	RootCmd.AddCommand(lint.LintCmd.MustGetCobraCommand())
	RootCmd.AddCommand(port.PortCmd)
	RootCmd.AddCommand(portal.PortalCmd)
	RootCmd.AddCommand(run.StarlarkRunCmd.MustGetCobraCommand())
	RootCmd.AddCommand(service.ServiceCmd)
	RootCmd.AddCommand(_import.ImportCmd.MustGetCobraCommand())
	RootCmd.AddCommand(twitter.TwitterCmd.MustGetCobraCommand())
	RootCmd.AddCommand(version.VersionCmd)
	RootCmd.AddCommand(web.WebCmd.MustGetCobraCommand())
	RootCmd.AddCommand(_package.PackageCmd)
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func globalSetup(cmd *cobra.Command, args []string) error {
	if err := setupCLILogs(cmd); err != nil {
		return stacktrace.Propagate(err, "An error occurred setting up CLI logs")
	}
	checkCLIVersion(cmd)
	//It is necessary to try track this metric on every execution to have at least one successful deliver
	if err := user_send_metrics_election.SendAnyBackloggedUserMetricsElectionEvent(); err != nil {
		//We don't want to interrupt users flow if something fails when tracking metrics
		logrus.Debugf("An error occurred tracking user consent to send metrics election\n%v", err)
	}

	printKurtosisCommandToFile(cmd, args)
	return nil
}

func printKurtosisCommandToFile(cmd *cobra.Command, args []string) {
	fileLogger := out.GetFileLogger()
	flagsSetByUsers := getFlagsSetByUsers(cmd.Flags())
	fileLogger.Infof("===== Executing Command: %v %v %v %v =====",
		command_str_consts.KurtosisCmdStr,
		cmd.Name(),
		strings.Join(flagsSetByUsers, " "),
		strings.Join(args, " "))
}

func getFlagsSetByUsers(flagSet *pflag.FlagSet) []string {
	var flagsSetByUser []string
	visitOnlySetFlags := func(flag *pflag.Flag) {
		flagsSetByUser = append(flagsSetByUser, fmt.Sprintf("--%v", flag.Name), flag.Value.String())
	}
	flagSet.Visit(visitOnlySetFlags)
	return flagsSetByUser
}

func setupCLILogs(cmd *cobra.Command) error {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	if err != nil {
		return stacktrace.Propagate(err, "Could not parse log level string '%v'", logLevelStr)
	}
	logrus.SetOutput(cmd.OutOrStdout())
	logrus.SetLevel(logLevel)
	return nil
}

func checkCLIVersion(cmd *cobra.Command) {
	// We temporarily set the logrus output to STDERR so that only these version warning messages get sent there
	// This is so that if you're running a command that actually prints output (e.g. 'completion', to generate completions)
	//  then this version check message doesn't show up in the output and potentially mess things up
	currentOut := logrus.StandardLogger().Out
	logrus.SetOutput(cmd.ErrOrStderr())
	defer logrus.SetOutput(currentOut)

	if !shouldPesterUsersAboutVersions() {
		return
	}

	isLatestVersion, latestVersion, err := isLatestCLIVersion()
	if err != nil {
		logrus.Warning("An error occurred trying to check if you are running the latest Kurtosis CLI version.")
		logrus.Debugf("Checking latest version error: %v", err)
		logrus.Warningf("Your current version is '%v'. You can manually upgrade the CLI tool following these instructions: %v", kurtosis_version.KurtosisVersion, user_support_constants.UpgradeCLIInstructionsPage)
		return
	}

	if !isLatestVersion {
		logrus.Warningf("You are running an old version of the Kurtosis CLI; we suggest you to update it to the latest version, '%v'. You can manually upgrade the CLI tool following these instructions: %v", latestVersion, user_support_constants.UpgradeCLIInstructionsPage)
	}
}

func isLatestCLIVersion() (bool, string, error) {
	ownVersionStr := kurtosis_version.KurtosisVersion
	latestVersionStr, err := getLatestCLIReleaseVersion()
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred getting the latest release version number from the GitHub public API")
	}

	ownSemver, err := parseVersionStrToSemVer(ownVersionStr)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred parsing own version string '%v' to sem version", ownVersionStr)
	}

	latestSemver, err := parseVersionStrToSemVer(latestVersionStr)
	if err != nil {
		return false, "", stacktrace.Propagate(err, "An error occurred parsing latest version string '%v' to sem version", latestVersionStr)
	}

	compareResult := ownSemver.Compare(latestSemver)

	//compareResult = 1  means that the own version is newer than the latest version, (e.g.: during a new release)
	if compareResult >= 0 {
		return true, latestVersionStr, nil
	}

	return false, latestVersionStr, nil
}

func parseVersionStrToSemVer(versionStr string) (*semver.Version, error) {
	semVer, err := semver.NewVersion(versionStr)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing version string '%v' to sem version", versionStr)
	}
	return semVer, nil
}

func getLatestCLIReleaseVersion() (string, error) {

	latestCLIReleaseVersionCacheFilepath, err := host_machine_directories.GetLatestCLIReleaseVersionCacheFilepath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the latest release version cache filepath")
	}
	logrus.Debugf("Cache filepath: '%v'", latestCLIReleaseVersionCacheFilepath)

	latestCLIVersion, err := getLatestCLIReleaseVersionFromCacheFile(latestCLIReleaseVersionCacheFilepath)
	if err != nil {
		logrus.Debugf("An error occurred getting latest released CLI version from cache file. Error: \n%v", err)
	}
	if latestCLIVersion != "" {
		logrus.Debugf("Got the latest released CLI version '%v' from the cache file", latestCLIVersion)
		return latestCLIVersion, nil
	}

	latestCLIVersion, err = getLatestCLIReleaseVersionFromGitHub()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting latest released CLI version from GitHub")
	}

	if err := saveLatestCLIReleaseVersionInCacheFile(latestCLIReleaseVersionCacheFilepath, latestCLIVersion); err != nil {
		logrus.Debugf("We tried to save the latest release version '%v' in the cache file, but doing so threw an error:\n%v", latestCLIVersion, err)
	}

	logrus.Debugf("Got the latest released CLI version '%v' from GitHub API", latestCLIVersion)
	return latestCLIVersion, nil
}

func getLatestCLIReleaseVersionFromGitHub() (string, error) {
	var (
		client = &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       0,
		}
		requestMethod  = "GET"
		requestBody    io.Reader
		responseObject GitHubReleaseReponse
	)

	request, err := http.NewRequest(requestMethod, latestReleaseOnGitHubURL, requestBody)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred creating new HTTP GET request to URL '%v' ", latestReleaseOnGitHubURL)
	}

	request.Header.Add(acceptHttpHeaderKey, acceptHttpHeaderValue)
	request.Header.Add(contentTypeHttpHeaderKey, contentTypeHttpHeaderValue)
	request.Header.Add(userAgentHttpHeaderKey, userAgentHttpHeaderValue)

	response, err := client.Do(request)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred executing HTTP GET request to URL '%v' ", latestReleaseOnGitHubURL)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			logrus.Warnf("We tried to close the response body, but doing so threw an error:\n%v", err)
		}
	}()

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading the HTTP response body")
	}

	if err := json.Unmarshal(bodyBytes, &responseObject); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred deserializing the latest release body response")
	}

	latestVersion := strings.TrimLeft(responseObject.TagName, optionalSemverPrefix)
	if latestVersion == "" {
		return "", stacktrace.NewError("The latest release version got from GitHub releases is empty")
	}

	return latestVersion, nil
}

func saveLatestCLIReleaseVersionInCacheFile(filepath, latestReleaseVersion string) error {

	now := time.Now()
	cacheCreationDateString := now.Format(latestCLIReleaseCacheFileCreationDateTimeFormat)
	content := strings.Join([]string{cacheCreationDateString, latestReleaseVersion}, latestCLIReleaseCacheFileContentSeparator)
	fileContent := []byte(content)

	logrus.Debugf("Saving content '%v' in cache file...", content)
	if err := os.WriteFile(filepath, fileContent, getLatestCLIReleaseCacheFilePermissions); err != nil {
		return stacktrace.Propagate(err, "An error occurred saving content '%v' in latest release version cache file", content)
	}
	logrus.Debugf("Content successfully saved in cache file")
	return nil
}

func getLatestCLIReleaseVersionFromCacheFile(filepath string) (string, error) {
	logrus.Debugf("Getting cache file content...")
	cacheFile, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("The latest release cache file has not be created yet.")
			return "", nil
		}
		return "", stacktrace.Propagate(err, "An error occurred opening the '%v' file", filepath)
	}
	defer func() {
		if err := cacheFile.Close(); err != nil {
			logrus.Warnf("We tried to close the latest release CLI version cache file, but doing so threw an error:\n%v", err)
		}
	}()

	fileContentBytes, err := io.ReadAll(cacheFile)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading cache file")
	}

	fileContent := string(fileContentBytes)

	if fileContent == "" {
		logrus.Debug("The cache file is empty, skipping getting the latest released version from it")
		return "", nil
	}

	//cacheFileContent should have this schema [{cacheCreationDate, latestReleaseVersion}]
	cacheFileContent := strings.Split(fileContent, latestCLIReleaseCacheFileContentSeparator)
	if len(cacheFileContent) != latestCLIReleaseCacheFileContentColumnsAmount {
		return "", stacktrace.NewError("The cache file content only had %v elems, but we expected %v", len(cacheFileContent), latestCLIReleaseCacheFileContentColumnsAmount)
	}
	dateString := cacheFileContent[latestCLIReleaseCacheFileContentDateIndex]
	latestReleaseVersion := cacheFileContent[latestCLIReleaseCacheFileContentVersionIndex]
	logrus.Debugf("Successfully got cache file content '%+v'", cacheFileContent)

	cacheCreationDate, err := time.Parse(latestCLIReleaseCacheFileCreationDateTimeFormat, dateString)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing date string '%v' from cache file", dateString)
	}

	cacheExpirationDate := cacheCreationDate.Add(latestCLIReleaseCacheFileExpirationHours * time.Hour)
	logrus.Debugf("Cache Date expiration date '%v'", cacheExpirationDate)

	now := time.Now()
	logrus.Debugf("Now '%v'", now)

	if now.After(cacheExpirationDate) {
		logrus.Debugf("The latest release version cache file content is out-of-date, it was generated on '%v' and it expired on '%v'", cacheCreationDate, cacheExpirationDate)
		return "", nil
	}

	return latestReleaseVersion, nil
}

// if the check ever fails we just return true and check for versions anyway
func shouldPesterUsersAboutVersions() bool {
	lastPesteredUsersAboutVersionsFilepath, err := host_machine_directories.GetLastPesteredUserAboutOldVersionsFilepath()
	if err != nil {
		logrus.Debugf("Tried getting the path to the %s file but failed with error \n'%s'", host_machine_directories.LastPesteredUserAboutOldVersionFilename, err)
		return true
	}

	fileStatus, err := os.Stat(lastPesteredUsersAboutVersionsFilepath)

	if os.IsNotExist(err) {
		createOrChangeModificationTimeOfLastPesteredUserFile(lastPesteredUsersAboutVersionsFilepath, host_machine_directories.LastPesteredUserAboutOldVersionFilename)
		return true
	}

	if err != nil {
		logrus.Debugf("Tried checking if %s at '%s' exists but failed with error '%s'\n", host_machine_directories.LastPesteredUserAboutOldVersionFilename, lastPesteredUsersAboutVersionsFilepath, err)
		return true
	}

	now := time.Now()

	if now.After(fileStatus.ModTime().Add(frequencyToPesterUsersAboutVersions)) {
		// we touch it again so that the timer restarts
		createOrChangeModificationTimeOfLastPesteredUserFile(lastPesteredUsersAboutVersionsFilepath, host_machine_directories.LastPesteredUserAboutOldVersionFilename)
		return true
	}

	// no error occurred and the last time we bothered the user was not frequencyToPesterUsersAboutVersions duration before
	// we don't have to bother the user
	return false
}

// this file is an empty file, if we run Create on it again we don't lose anything but we do change the modification time
// the modification time is what we care about
func createOrChangeModificationTimeOfLastPesteredUserFile(lastPesteredUsersAboutVersionsFilepath, fileNameForLogging string) {
	_, err := os.Create(lastPesteredUsersAboutVersionsFilepath)
	if err != nil {
		logrus.Debugf("Tried creating the %s file at '%s' but failed with error \n'%s'", fileNameForLogging, lastPesteredUsersAboutVersionsFilepath, err)
	}
}
