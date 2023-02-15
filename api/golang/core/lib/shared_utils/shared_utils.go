package shared_utils

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	grpcDataTransferLimit     = 3999000 //3.999 Mb. 1kb wiggle room. 1kb being about the size of a simple 2 paragraph readme.
	tempCompressionDirPattern = "upload-compression-cache-"
	compressionExtension      = ".tgz"
	defaultTmpDir             = ""
)

func CompressPath(pathToCompress string, accountForGRPCLimit bool) ([]byte, error) {
	pathToCompress = strings.TrimRight(pathToCompress, string(filepath.Separator))
	uploadFileInfo, err := os.Stat(pathToCompress)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was a path error for '%s' during file compression.", pathToCompress)
	}

	// This allows us to archive contents of dirs in root instead of nesting
	var filepathsToUpload []string
	if uploadFileInfo.IsDir() {
		filesInDirectory, err := ioutil.ReadDir(pathToCompress)
		if err != nil {
			return nil, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
		}
		if len(filesInDirectory) == 0 {
			return nil, stacktrace.NewError("The directory '%s' you are trying to compress is empty", pathToCompress)
		}

		for _, fileInDirectory := range filesInDirectory {
			filepathToUpload := filepath.Join(pathToCompress, fileInDirectory.Name())
			filepathsToUpload = append(filepathsToUpload, filepathToUpload)
		}
	} else {
		filepathsToUpload = append(filepathsToUpload, pathToCompress)
	}

	tempDir, err := ioutil.TempDir(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)
	if err = archiver.Archive(filepathsToUpload, compressedFilePath); err != nil {
		return nil, stacktrace.Propagate(err, "Failed to compress '%s'.", pathToCompress)
	}

	compressedFileInfo, err := os.Stat(compressedFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"Failed to create a temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	if accountForGRPCLimit && compressedFileInfo.Size() >= grpcDataTransferLimit {
		return nil, stacktrace.NewError(
			"The files you are trying to upload, which are now compressed, exceed or reach 4mb, a limit imposed by gRPC. " +
				"Please reduce the total file size and ensure it can compress to a size below 4mb.")
	}
	content, err := ioutil.ReadFile(compressedFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"There was an error reading from the temporary tar file '%s' recently compressed for upload.",
			compressedFileInfo.Name())
	}

	return content, nil
}
