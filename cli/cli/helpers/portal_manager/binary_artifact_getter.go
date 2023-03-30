package portal_manager

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/google/go-github/v50/github"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	kurtosisTechGithubOrg        = "kurtosis-tech"
	kurtosisPortalGithubRepoName = "kurtosis-portal"

	githubAssetExtension = ".tar.gz"

	portalBinaryFileMode  = 0700
	portalVersionFileMode = 0600

	osArchitectureSeparator = "_"

	// not expected to ever have more than 50 assets per release, won't do pagination for now
	assetsFirstPageNumber = 0
	numberOfAssetsPerPage = 50
)

// DownloadLatestKurtosisPortalBinary downloads the latest version of Kurtosis Portal.
// It returns true if a new version was downloaded and installed, false if it no-oped because latest was already
// installed or because latest version information could not be retrieved.
// Note that it returns an error only if the end state is such that the portal cannot be run properly. I.e. if a
// Portal is already installed, and it failed to retrieve the latest version for example, it returns gracefully
// as this is not critical (current version will be used and will continue to run fine)
func DownloadLatestKurtosisPortalBinary(ctx context.Context) (bool, error) {
	binaryFilePath, err := host_machine_directories.GetPortalBinaryFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to get file path to Kurtosis Portal binary file")
	}
	currentVersionFilePath, err := host_machine_directories.GetPortalVersionFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to get file path to Kurtosis Portal version file")
	}

	currentVersionStrMaybe, err := getVersionCurrentlyInstalled(currentVersionFilePath)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error retrieving version of currently installed Portal")
	}

	ghClient := github.NewClient(http.DefaultClient)
	latestRelease, err := getLatestRelease(ctx, ghClient)
	if err != nil {
		return false, defaultToCurrentVersionOrError(currentVersionStrMaybe, err)
	}

	latestVersionStr, currentVersionMatchesLatest, err := compareLatestVersionWithCurrent(currentVersionStrMaybe, latestRelease)
	if err != nil {
		return false, defaultToCurrentVersionOrError(currentVersionStrMaybe, err)
	}
	if currentVersionMatchesLatest {
		logrus.Infof("Kurtosis Portal version '%s' is the latest and already installed", latestVersionStr)
		return false, nil
	}

	if currentVersionStrMaybe == "" {
		// no Portal is currently installed. Installing a brand new one
		logrus.Infof("Installing Kurtosis Portal version '%s'", latestVersionStr)
	} else {
		logrus.Infof("Upgrading currently Kurtosis Portal version '%s' to '%s'", currentVersionStrMaybe, latestVersionStr)
	}

	githubAssetContent, err := downloadGithubAsset(ctx, ghClient, latestRelease)
	if err != nil {
		return false, defaultToCurrentVersionOrError(currentVersionStrMaybe, err)
	}

	if err = extractAssetTgzToBinaryFileOnDisk(githubAssetContent, latestVersionStr, binaryFilePath, currentVersionFilePath); err != nil {
		return false, defaultToCurrentVersionOrError(currentVersionStrMaybe, err)
	}
	return true, nil
}

func getLatestRelease(ctx context.Context, ghClient *github.Client) (*github.RepositoryRelease, error) {
	// First, browse the list of releases for the Kurtosis Portal repo
	latestRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to retrieve latest version of Kurtosis Portal form GitHub")
	}
	logrus.Debugf("Latest release is %s", latestRelease.GetName())
	return latestRelease, nil
}

func getVersionCurrentlyInstalled(currentVersionFilePath string) (string, error) {
	if _, err := createFileIfNecessary(currentVersionFilePath); err != nil {
		return "", stacktrace.Propagate(err, "Unable to create Kurtosis Portal version file on disk")
	}
	currentVersionBytes, err := os.ReadFile(currentVersionFilePath)
	if err != nil {
		return "", stacktrace.Propagate(err, "Unable to read content of Kurtosis Portal version file")
	}
	return strings.TrimSpace(string(currentVersionBytes)), nil
}

func compareLatestVersionWithCurrent(currentVersionStrIfAny string, latestRelease *github.RepositoryRelease) (string, bool, error) {
	latestVersionStr := latestRelease.GetName()
	if currentVersionStrIfAny == latestVersionStr {
		return latestVersionStr, true, nil
	}
	return latestVersionStr, false, nil
}

func downloadGithubAsset(ctx context.Context, ghClient *github.Client, latestRelease *github.RepositoryRelease) (io.ReadCloser, error) {
	// Get all assets associated with this release and identify the one matching the current machine architecture
	opts := &github.ListOptions{
		Page:    assetsFirstPageNumber,
		PerPage: numberOfAssetsPerPage,
	}
	allReleaseAssets, _, err := ghClient.Repositories.ListReleaseAssets(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName, latestRelease.GetID(), opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get list of assets from latest version")
	}
	detectedOsArch := fmt.Sprintf("%s%s%s", runtime.GOOS, osArchitectureSeparator, runtime.GOARCH)
	assetFileExpectedSuffix := fmt.Sprintf("%s%s", detectedOsArch, githubAssetExtension)
	var releaseAssetToUse *github.ReleaseAsset
	for _, releaseAsset := range allReleaseAssets {
		if strings.HasSuffix(releaseAsset.GetName(), assetFileExpectedSuffix) {
			releaseAssetToUse = releaseAsset
		}
	}
	if releaseAssetToUse == nil {
		return nil, stacktrace.NewError("Unable to find Kurtosis Portal binary matching the current OS and architecture. Detected OS and architecture was: '%s'", detectedOsArch)
	}
	logrus.Debugf("Kurtosis Portal binary found: '%s' with ID '%d'", releaseAssetToUse.GetName(), releaseAssetToUse.GetID())

	// Download the content of the asset. It is expected to be a .tar.gz file containing the Kurtosis Portal binary
	artifactContent, _, err := ghClient.Repositories.DownloadReleaseAsset(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName, releaseAssetToUse.GetID(), http.DefaultClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to download Kurtosis Portal binary. Name was: '%s', ID was: '%d'", releaseAssetToUse.GetName(), releaseAssetToUse.GetID())
	}
	return artifactContent, nil
}

func extractAssetTgzToBinaryFileOnDisk(assetContent io.ReadCloser, assetVersion string, destFilePath string, destVersionFilePath string) error {
	gzipReader, err := gzip.NewReader(assetContent)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to open a GZIP reader on the asset content")
	}
	defer gzipReader.Close()

	assetTarReader := tar.NewReader(gzipReader)
	onlyFileHeader, err := assetTarReader.Next()
	if err != nil || onlyFileHeader == nil {
		return stacktrace.Propagate(err, "Asset archive seems to be empty, this is unexpected")
	}
	if onlyFileHeader.Typeflag != tar.TypeReg {
		return stacktrace.Propagate(err, "Archive seems to be a directory, but expecting a single binary file")
	}

	portalBinaryFile, err := os.Create(destFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create a new empty file to store Kurtosis Portal binary")
	}
	if err = os.Chmod(portalBinaryFile.Name(), portalBinaryFileMode); err != nil {
		return stacktrace.Propagate(err, "Unable to switch file mode to executable for Kurtosis Portal binary file")
	}

	if _, err = io.Copy(portalBinaryFile, assetTarReader); err != nil {
		return stacktrace.Propagate(err, "Unable to copy content of Kurtosis Portal binary to executable file")
	}

	if err = os.WriteFile(destVersionFilePath, []byte(assetVersion), portalVersionFileMode); err != nil {
		logrus.Warnf("Portal binary file successfully stored but an error occurred persisting its corresponding " +
			"version. This is not critical, it will be retried Kurtosis Portal version is checked.")
		logrus.Debugf("Error was: %v", err.Error())
	}
	logrus.Debugf("Kurtosis Portal binary downloaded to '%s'", destFilePath)
	return nil
}

func defaultToCurrentVersionOrError(currentVersionStr string, nonNilError error) error {
	if currentVersionStr != "" {
		logrus.Warnf("Checking for latest version of Kurtosis Portal failed. Currently installed version '%s' will be used", currentVersionStr)
		logrus.Debugf("Error was: %v", nonNilError.Error())
		return nil
	}
	return stacktrace.Propagate(nonNilError, "An error occurred installing Kurtosis Portal latest version")
}
