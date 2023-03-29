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
)

// DownloadLatestKurtosisPortalBinary downloads the latest version of Kurtosis Portal.
// It returns true if a new version was downloaded and installed, false if it no-oped because latest was already installed
// If no error occurs, it also returns the path to the binary file just installed
func DownloadLatestKurtosisPortalBinary(ctx context.Context) (bool, error) {
	binaryFilePath, err := host_machine_directories.GetPortalBinaryFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to get file path to Kurtosis Portal binary file")
	}
	currentVersionFilePath, err := host_machine_directories.GetPortalVersionFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to get file path to Kurtosis Portal version file")
	}

	ghClient := github.NewClient(http.DefaultClient)
	latestRelease, err := getLatestRelease(ctx, ghClient)
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to retrieve latest release of Kurtosis Portal from Github repository")
	}

	latestVersionStr, currentVersionStr, currentVersionMatchesLatest, err := compareLatestVersionWithCurrent(currentVersionFilePath, latestRelease)
	if err != nil {
		return false, stacktrace.Propagate(err, "Error checking if latest version matches current")
	}
	if currentVersionMatchesLatest {
		logrus.Infof("Kurtosis Portal version '%s' is the latest and already installed", latestVersionStr)
		return false, nil
	}

	if currentVersionStr == "" {
		logrus.Infof("Installing Kurtosis Portal version '%s'", latestVersionStr)
	} else {
		logrus.Infof("Upgrading currently Kurtosis Portal version '%s' to '%s'", currentVersionStr, latestVersionStr)
	}

	githubAssetContent, err := downloadGithubAsset(ctx, ghClient, latestRelease)
	if err != nil {
		return false, stacktrace.Propagate(err, "Unable to browse Kurtosis Portal released assets from Github to download the latest")
	}

	if err = extractBinaryToFile(githubAssetContent, latestVersionStr, binaryFilePath, currentVersionFilePath); err != nil {
		return false, stacktrace.Propagate(err, "Unable to extract Kurtosis Portal binary file content to file on disk")
	}
	return true, nil
}

func getLatestRelease(ctx context.Context, ghClient *github.Client) (*github.RepositoryRelease, error) {
	// First, browse the list of releases for the Kurtosis Portal repo
	latestRelease, _, err := ghClient.Repositories.GetLatestRelease(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to retrieve latest version of Kurtosis Portal form Github")
	}
	logrus.Debugf("Latest release is %s", latestRelease.GetName())
	return latestRelease, nil
}

func compareLatestVersionWithCurrent(currentVersionFilePath string, latestRelease *github.RepositoryRelease) (string, string, bool, error) {
	latestVersionStr := latestRelease.GetName()

	if _, err := CreateFileIfNecessary(currentVersionFilePath); err != nil {
		return "", "", false, stacktrace.Propagate(err, "Unable to create Kurtosis Portal version file on disk")
	}

	currentVersionBytes, err := os.ReadFile(currentVersionFilePath)
	if err != nil {
		return "", "", false, stacktrace.Propagate(err, "Unable to read content of Kurtosis Portal version file")
	}
	currentVersionStr := strings.TrimSpace(string(currentVersionBytes))
	if currentVersionStr == latestVersionStr {
		return latestVersionStr, currentVersionStr, true, nil
	}
	return latestVersionStr, currentVersionStr, false, nil
}

func downloadGithubAsset(ctx context.Context, ghClient *github.Client, latestRelease *github.RepositoryRelease) (io.ReadCloser, error) {
	// Get all assets associated with this release and identify the one matching the current machine architecture
	opts := &github.ListOptions{
		Page:    0,
		PerPage: 50, // we will probably never have more than 10 assets per release, hopefully...
	}
	allReleaseAssets, _, err := ghClient.Repositories.ListReleaseAssets(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName, latestRelease.GetID(), opts)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to get list of assets from latest version")
	}
	detectedArchitecture := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	assetFileExpectedSuffix := fmt.Sprintf("%s%s", detectedArchitecture, githubAssetExtension)
	var releaseAssetToUse *github.ReleaseAsset
	for _, releaseAsset := range allReleaseAssets {
		if strings.HasSuffix(releaseAsset.GetName(), assetFileExpectedSuffix) {
			releaseAssetToUse = releaseAsset
		}
	}
	if releaseAssetToUse == nil {
		return nil, stacktrace.NewError("Unable to find Kurtosis Portal binary matching the current architecture. Detected architecture was: '%s'", detectedArchitecture)
	}
	logrus.Debugf("Kurtosis Portal binary found: '%s' with ID '%d'", releaseAssetToUse.GetName(), releaseAssetToUse.GetID())

	// Download the content of the asset. It is expected to be a .tar.gz file containing the Kurtosis Portal binary
	artifactContent, _, err := ghClient.Repositories.DownloadReleaseAsset(ctx, kurtosisTechGithubOrg, kurtosisPortalGithubRepoName, releaseAssetToUse.GetID(), http.DefaultClient)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Unable to download Kurtosis portal binary. Name was: '%s', ID was: '%d'", releaseAssetToUse.GetName(), releaseAssetToUse.GetID())
	}
	return artifactContent, nil
}

func extractBinaryToFile(assetContent io.ReadCloser, assetVersion string, destFilePath string, destVersionFilePath string) error {
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
		return stacktrace.Propagate(err, "Archive seems to be a directory, but expecting a single file")
	}

	portalBinaryFile, err := os.Create(destFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to create a new executable file to store Kurtosis Portal binary")
	}
	if err = os.Chmod(portalBinaryFile.Name(), portalBinaryFileMode); err != nil {
		return stacktrace.Propagate(err, "Unable to create a new executable file to store Kurtosis Portal binary")
	}

	if _, err = io.Copy(portalBinaryFile, assetTarReader); err != nil {
		return stacktrace.Propagate(err, "Unable to copy content of Kurtosis Portal binary to executable file")
	}

	if err = os.WriteFile(destVersionFilePath, []byte(assetVersion), portalVersionFileMode); err != nil {
		return stacktrace.Propagate(err, "Portal binary file successfully stored but an error occurred persisting its corresponding version.")
	}
	logrus.Debugf("Kurtosis Portal binary downloaded to '%s'", destFilePath)
	return nil
}
