package shared_utils

import (
	"crypto/md5"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	kurtosisDataTransferLimit = 100 * 1024 * 1024 // 100 MB
	tempCompressionDirPattern = "upload-compression-cache-"
	compressionExtension      = ".tgz"
	defaultTmpDir             = ""
	ownerAllPermissions       = 0700
	readonlyPermissions       = 0400
)

// CompressPath compressed the entire content of the file or directory at pathToCompress and returns an io.ReadCloser
// of the TGZ archive created, alongside the size (in bytes) of the archive
// The consumer should take care of closing the io.ReadClose returned
func CompressPath(pathToCompress string, enforceMaxFileSizeLimit bool) (io.ReadCloser, uint64, []byte, error) {
	// First we compute the hash of the content about to be compressed
	// Note that we're computing this "complex" hash here because the way tar.gz works, it writes files metadata to the
	// archive, like last date of modification for each file. In our case, we just want to hash the content, regardless
	// of those metadata.
	filesInPathRecursive, err := listFilesInPathDeterministic(pathToCompress, true)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
	}
	compressedFileContentMd5 := md5.New()
	for _, filePath := range filesInPathRecursive {
		pathRelativeToRoot := strings.Replace(filePath, pathToCompress, "", 1)
		// we write both the file path relative to the root AND the file content to the hash, such that if the structure
		// of the archive change but the files remains the same, the hash will change as well
		compressedFileContentMd5.Write([]byte(pathRelativeToRoot))
		if err = writeFileContent(filePath, compressedFileContentMd5); err != nil {
			return nil, 0, nil, stacktrace.Propagate(err, "An error occurred computing files artifact hash for '%s' in '%s'", filePath, pathToCompress)
		}
	}

	// Then we compress the content into an archive
	filepathsToUpload, err := listFilesInPathDeterministic(pathToCompress, false)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
	}

	tempDir, err := os.MkdirTemp(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)
	if err = archiver.Archive(filepathsToUpload, compressedFilePath); err != nil {
		return nil, 0, nil, stacktrace.Propagate(err, "Failed to compress '%s'.", pathToCompress)
	}

	compressedFile, err := os.OpenFile(compressedFilePath, os.O_RDONLY, ownerAllPermissions)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err,
			"Failed to create a temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	compressedFileInfo, err := compressedFile.Stat()
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err,
			"Failed to inspect temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	var compressedFileSize uint64
	if compressedFileInfo.Size() >= 0 {
		compressedFileSize = uint64(compressedFileInfo.Size())
	} else {
		return nil, 0, nil, stacktrace.Propagate(err,
			"Failed to compute archive size for temporary file '%s' obtained compressing path '%s'",
			tempDir, pathToCompress)
	}

	if enforceMaxFileSizeLimit && compressedFileSize >= kurtosisDataTransferLimit {
		return nil, 0, nil, stacktrace.NewError(
			"The files you are trying to upload, which are now compressed, exceed or reach 100mb. " +
				"Manipulation (i.e. uploads or downloads) of files larger than 100mb is currently disallowed in Kurtosis.")
	}

	return compressedFile, compressedFileSize, compressedFileContentMd5.Sum(nil), nil
}

// listFilesInPathDeterministic returns the list of file paths in the path passed as an argument.
// If the path points to a file, it only returns the given file path
// If recursiveMode is true, it recursively iterates over the directory inside the given path
func listFilesInPathDeterministic(path string, recursiveMode bool) ([]string, error) {
	var listOfFiles []string
	err := listFilesInPathInternal(&listOfFiles, path, recursiveMode, true)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error listing files in '%s'", path)
	}
	sort.Strings(listOfFiles)
	return listOfFiles, nil
}

func listFilesInPathInternal(filesInPath *[]string, path string, recursiveMode bool, topLevel bool) error {
	trimmedPath := strings.TrimRight(path, string(filepath.Separator))
	uploadFileInfo, err := os.Stat(trimmedPath)
	if err != nil {
		return stacktrace.Propagate(err, "There was a path error for '%s'.", trimmedPath)
	}

	// This allows us to archive contents of dirs in root instead of nesting
	if uploadFileInfo.IsDir() {
		filesInDirectory, err := os.ReadDir(trimmedPath)
		if err != nil {
			return stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", trimmedPath)
		}
		if topLevel && len(filesInDirectory) == 0 {
			return stacktrace.NewError("The directory '%s' you are trying to compress is empty", path)
		}
		for _, fileInDirectory := range filesInDirectory {
			fileInDirectoryPath := filepath.Join(trimmedPath, fileInDirectory.Name())
			*filesInPath = append(*filesInPath, fileInDirectoryPath)
			if recursiveMode && fileInDirectory.IsDir() {
				err := listFilesInPathInternal(filesInPath, fileInDirectoryPath, recursiveMode, false)
				if err != nil {
					return stacktrace.Propagate(err, "Error recursively listing files in '%s'", fileInDirectoryPath)
				}
			}
		}
	} else {
		*filesInPath = append(*filesInPath, trimmedPath)
	}
	return nil
}

// writeFileContent writes the content of the file at filePath to the provided writer.
// If the path points to a directory, it does nothing
func writeFileContent(filePath string, writer io.Writer) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return stacktrace.Propagate(err, "There was a path error for '%s'.", filePath)
	}
	if fileInfo.IsDir() {
		// no content for directories
		return nil
	}
	file, err := os.OpenFile(filePath, os.O_RDONLY, readonlyPermissions)
	defer func() {
		if err := file.Close(); err != nil {
			logrus.Errorf("An unexpected error occured closing file '%s'. It will remain open, this is not critical"+
				"but might indicate a malfunction in how files are handled. Error was:\n%v", filePath, err)
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "Fail to read file '%s' to hash its content.", filePath)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to hash file content for file '%s'.", filePath)
	}
	return nil
}
