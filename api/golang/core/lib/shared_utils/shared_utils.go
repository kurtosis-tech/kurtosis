package shared_utils

import (
	"crypto/md5"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archiver/v3"
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
)

func CompressPath(pathToCompress string, enforceMaxFileSizeLimit bool) ([]byte, []byte, error) {
	filesInPath, err := listFilesInPathDeterministic(pathToCompress, false)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Error reading path '%s'", pathToCompress)
	}

	tempDir, err := os.MkdirTemp(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	contentMd5 := md5.New()
	filesInPathRecursive, err := listFilesInPathDeterministic(pathToCompress, true)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err, "Error reading path '%s'", pathToCompress)
	}
	for _, filePath := range filesInPathRecursive {
		pathRelativeToRoot := strings.Replace(filePath, pathToCompress, "", 1)
		contentMd5.Write([]byte(pathRelativeToRoot))
		if err = writeFileContent(filePath, contentMd5); err != nil {
			return nil, nil, stacktrace.Propagate(err, "An error occurred computing files artifact hash for '%s'", pathToCompress)
		}
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)
	if err = archiver.Archive(filesInPath, compressedFilePath); err != nil {
		return nil, nil, stacktrace.Propagate(err, "Failed to compress '%s'.", pathToCompress)
	}

	compressedFileInfo, err := os.Stat(compressedFilePath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,
			"Failed to create a temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	if enforceMaxFileSizeLimit && compressedFileInfo.Size() >= kurtosisDataTransferLimit {
		return nil, nil, stacktrace.NewError(
			"The files you are trying to upload, which are now compressed, exceed or reach 100mb. " +
				"Manipulation (i.e. uploads or downloads) of files larger than 100mb is currently disallowed in Kurtosis.")
	}
	content, err := os.ReadFile(compressedFilePath)
	if err != nil {
		return nil, nil, stacktrace.Propagate(err,
			"There was an error reading from the temporary tar file '%s' recently compressed for upload.",
			compressedFileInfo.Name())
	}

	return content, contentMd5.Sum(nil), nil
}

func listFilesInPathDeterministic(path string, recursiveMode bool) ([]string, error) {
	var listOfFiles []string
	err := listFilesInPathRecursive(&listOfFiles, path, recursiveMode)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error listing files in '%s'", path)
	}
	sort.Strings(listOfFiles)
	return listOfFiles, nil
}

func listFilesInPathRecursive(filesInPath *[]string, path string, recursiveMode bool) error {
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
		for _, fileInDirectory := range filesInDirectory {
			fileInDirectoryPath := filepath.Join(trimmedPath, fileInDirectory.Name())
			if recursiveMode && fileInDirectory.IsDir() {
				err := listFilesInPathRecursive(filesInPath, fileInDirectoryPath, recursiveMode)
				if err != nil {
					return stacktrace.Propagate(err, "Error recursively listing files in '%s'", fileInDirectoryPath)
				}
			} else {
				*filesInPath = append(*filesInPath, fileInDirectoryPath)
			}
		}
	} else {
		*filesInPath = append(*filesInPath, trimmedPath)
	}
	return nil
}

func writeFileContent(filePath string, writer io.Writer) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0400)
	defer file.Close()
	if err != nil {
		return stacktrace.Propagate(err, "Fail to read file '%s' to hash its content.", filePath)
	}
	_, err = io.Copy(writer, file)
	if err != nil {
		return stacktrace.Propagate(err, "Unable to hash file content for file '%s'.", filePath)
	}
	return nil
}
