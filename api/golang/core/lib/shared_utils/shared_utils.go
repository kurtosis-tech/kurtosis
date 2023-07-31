package shared_utils

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver/v3"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	kurtosisDataTransferLimit = 100 * 1024 * 1024 // 100 MB
	tempCompressionDirPattern = "upload-compression-cache-"
	compressionExtension      = ".tgz"
	defaultTmpDir             = ""
	ownerAllPermissions       = 0700
)

// CompressPath compressed the entire content of the file or directory at pathToCompress and returns an io.ReadCloser
// of the TGZ archive created, alongside the size (in bytes) of the archive
// The consumer should take care of closing the io.ReadClose returned
func CompressPath(pathToCompress string, enforceMaxFileSizeLimit bool) (io.ReadCloser, uint64, error) {
	pathToCompress = strings.TrimRight(pathToCompress, string(filepath.Separator))
	uploadFileInfo, err := os.Stat(pathToCompress)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "There was a path error for '%s' during file compression.", pathToCompress)
	}

	// This allows us to archive contents of dirs in root instead of nesting
	var filepathsToUpload []string
	if uploadFileInfo.IsDir() {
		filesInDirectory, err := os.ReadDir(pathToCompress)
		if err != nil {
			return nil, 0, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
		}
		if len(filesInDirectory) == 0 {
			return nil, 0, stacktrace.NewError("The directory '%s' you are trying to compress is empty", pathToCompress)
		}

		for _, fileInDirectory := range filesInDirectory {
			filepathToUpload := filepath.Join(pathToCompress, fileInDirectory.Name())
			filepathsToUpload = append(filepathsToUpload, filepathToUpload)
		}
	} else {
		filepathsToUpload = append(filepathsToUpload, pathToCompress)
	}

	tempDir, err := os.MkdirTemp(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)
	if err = archiver.Archive(filepathsToUpload, compressedFilePath); err != nil {
		return nil, 0, stacktrace.Propagate(err, "Failed to compress '%s'.", pathToCompress)
	}

	compressedFile, err := os.OpenFile(compressedFilePath, os.O_RDONLY, ownerAllPermissions)
	if err != nil {
		return nil, 0, stacktrace.Propagate(err,
			"Failed to create a temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	compressedFileInfo, err := compressedFile.Stat()
	if err != nil {
		return nil, 0, stacktrace.Propagate(err,
			"Failed to inspect temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	var compressedFileSize uint64
	if compressedFileInfo.Size() >= 0 {
		compressedFileSize = uint64(compressedFileInfo.Size())
	} else {
		return nil, 0, stacktrace.Propagate(err,
			"Failed to compute archive size for temporary file '%s' obtained compressing path '%s'",
			tempDir, pathToCompress)
	}

	if enforceMaxFileSizeLimit && compressedFileSize >= kurtosisDataTransferLimit {
		return nil, 0, stacktrace.NewError(
			"The files you are trying to upload, which are now compressed, exceed or reach 100mb. " +
				"Manipulation (i.e. uploads or downloads) of files larger than 100mb is currently disallowed in Kurtosis.")
	}

	return compressedFile, compressedFileSize, nil
}
