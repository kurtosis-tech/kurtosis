package files

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
	"os"
	"path"
)

const (
	filesArtifactExtension  = ".tgz"
	filesArtifactPermission = 0o744

	defaultTmpDir = ""
	tmpDirPattern = "tmp-dir-for-download-*"
)

func DownloadFilesArtifactToLocation(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, artifactIdentifier string, absoluteDestinationPath string) error {

	artifactBytes, err := enclaveCtx.DownloadFilesArtifact(ctx, artifactIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred downloading files with identifier '%v' from enclave '%v'", artifactIdentifier, enclaveCtx.GetEnclaveName())
	}

	fileNameToWriteTo := fmt.Sprintf("%v%v", artifactIdentifier, filesArtifactExtension)
	destinationPathToDownloadFileTo := path.Join(absoluteDestinationPath, fileNameToWriteTo)

	err = os.WriteFile(destinationPathToDownloadFileTo, artifactBytes, filesArtifactPermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", destinationPathToDownloadFileTo, filesArtifactPermission)
	}
	logrus.Infof("File package with identifier '%v' downloaded to '%v'", artifactIdentifier, destinationPathToDownloadFileTo)
	return nil

}

func DownloadAndExtractFilesArtifact(ctx context.Context, enclaveCtx *enclaves.EnclaveContext, artifactIdentifier string, absoluteDestinationPath string) error {
	artifactBytes, err := enclaveCtx.DownloadFilesArtifact(ctx, artifactIdentifier)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred downloading files with identifier '%v' from enclave '%v'", artifactIdentifier, enclaveCtx.GetEnclaveName())
	}
	fileNameToWriteTo := fmt.Sprintf("%v%v", artifactIdentifier, filesArtifactExtension)

	tmpDirPath, err := os.MkdirTemp(defaultTmpDir, tmpDirPattern)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while creating a temporary directory to download the files artifact with identifier '%v' to", artifactIdentifier)
	}
	shouldCleanupTmpDir := false
	defer func() {
		if shouldCleanupTmpDir {
			os.RemoveAll(tmpDirPath)
		}
	}()
	tmpFileToWriteTo := path.Join(tmpDirPath, fileNameToWriteTo)
	err = os.WriteFile(tmpFileToWriteTo, artifactBytes, filesArtifactPermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while writing bytes to file '%v' with permission '%v'", tmpDirPath, filesArtifactPermission)
	}
	err = archiver.Unarchive(tmpFileToWriteTo, absoluteDestinationPath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred while extracting '%v' to '%v'", tmpFileToWriteTo, absoluteDestinationPath)
	}
	logrus.Infof("File package with identifier '%v' extracted to '%v'", artifactIdentifier, absoluteDestinationPath)

	shouldCleanupTmpDir = true
	return nil
}
