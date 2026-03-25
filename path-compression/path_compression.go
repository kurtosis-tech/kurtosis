package path_compression

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kurtosis-tech/stacktrace"
	"github.com/mholt/archives"

	"github.com/sirupsen/logrus"
)

const (
	extractDirPermissions     = 0755
	kurtosisDataTransferLimit = 2000 * 1024 * 1024 // ~2 GB
	tempCompressionDirPattern = "upload-compression-cache-"
	compressionExtension      = ".tgz"
	defaultTmpDir             = ""
	ownerAllPermissions       = 0700
	readonlyPermissions       = 0400
)

// ComputeContentHash computes an MD5 hash of the content at pathToHash.
// The hash includes both file paths (relative to root) and file contents, but NOT filesystem
// metadata like modification timestamps. This makes it stable across builds.
func ComputeContentHash(pathToHash string) ([]byte, error) {
	rules := loadIgnoreRules(pathToHash)
	filesInPathRecursive, err := listFilesInPathDeterministic(pathToHash, true, rules)
	if err != nil {
		return nil, stacktrace.Propagate(err, "There was an error listing files in '%s' for hashing", pathToHash)
	}
	hasher := md5.New()
	for _, filePath := range filesInPathRecursive {
		pathRelativeToRoot := strings.Replace(filePath, pathToHash, "", 1)
		hasher.Write([]byte(pathRelativeToRoot))
		if err = writeFileContent(filePath, hasher); err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred computing content hash for '%s' in '%s'", filePath, pathToHash)
		}
	}
	return hasher.Sum(nil), nil
}

// CompressPath is similar to CompressPathToFile but opens the archive and returns a ReadCloser to it.
// It's the consumer's responsibility to make sure the result gets closed appropriately
func CompressPath(pathToCompress string, enforceMaxFileSizeLimit bool) (io.ReadCloser, uint64, []byte, error) {
	compressedFilePath, compressedFileSize, compressedFileContentMd5, err := CompressPathToFile(pathToCompress, enforceMaxFileSizeLimit)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err,
			"An error occurred creating the archive from the files at '%s'", pathToCompress)
	}
	compressedFile, err := os.OpenFile(compressedFilePath, os.O_RDONLY, ownerAllPermissions)
	if err != nil {
		return nil, 0, nil, stacktrace.Propagate(err,
			"Failed to open the archive file at '%s' during files upload for '%s'.", compressedFilePath, pathToCompress)
	}
	return compressedFile, compressedFileSize, compressedFileContentMd5, nil
}

// CompressPathToFile compresses the entire content of the file or directory at pathToCompress and returns the path
// to the TGZ archive created, alongside the size (in bytes) of the archive and the md5 of its content
// Note: the MD5 is NOT the MD% of the archive file itself. See inline comments below for more info on this MD5 hash
func CompressPathToFile(pathToCompress string, enforceMaxFileSizeLimit bool) (string, uint64, []byte, error) {
	// First we compute the hash of the content about to be compressed
	compressedFileContentMd5, err := ComputeContentHash(pathToCompress)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred computing content hash for '%s'", pathToCompress)
	}

	// Then we compress the content into an archive
	rules := loadIgnoreRules(pathToCompress)
	filepathsToUpload, err := listFilesInPathDeterministic(pathToCompress, false, rules)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "There was an error in getting a list of files in the directory '%s' provided", pathToCompress)
	}

	tempDir, err := os.MkdirTemp(defaultTmpDir, tempCompressionDirPattern)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "Failed to create temporary directory '%s' for compression.", tempDir)
	}

	compressedFilePath := filepath.Join(tempDir, filepath.Base(pathToCompress)+compressionExtension)

	// Create output file
	outFile, err := os.Create(compressedFilePath)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred when creating the output file '%s'.", compressedFilePath)
	}
	defer outFile.Close()

	ctx := context.Background()

	files, err := archives.FilesFromDisk(
		ctx,
		nil, // use default settings
		mapFilePathOnDiskToRelativePathInArchive(pathToCompress, filepathsToUpload))
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred when creating files list for archive from '%s'.", pathToCompress)
	}

	format := archives.CompressedArchive{
		Compression: archives.Gz{
			CompressionLevel:   0,
			DisableMultistream: false,
			Multithreaded:      false,
		},
		Archival: archives.Tar{
			Format:          0,
			FormatGNU:       false,
			NumericUIDGID:   false,
			ContinueOnError: false,
			Uid:             0,
			Gid:             0,
			Uname:           "",
			Gname:           "",
		},
		Extraction: nil,
	}
	if err = format.Archive(ctx, outFile, files); err != nil {
		return "", 0, nil, stacktrace.Propagate(err, "An error occurred when compressing '%s'.", pathToCompress)
	}

	compressedFileInfo, err := os.Stat(compressedFilePath)
	if err != nil {
		return "", 0, nil, stacktrace.Propagate(err,
			"An error occurred when inspecting the temporary archive file at '%s' during files upload for '%s'.",
			tempDir, pathToCompress)
	}

	var compressedFileSize uint64
	if compressedFileInfo.Size() >= 0 {
		compressedFileSize = uint64(compressedFileInfo.Size())
	} else {
		return "", 0, nil, stacktrace.Propagate(err,
			"An error occurred when computing the archive size for the temporary file '%s' obtained compressing path '%s'",
			tempDir, pathToCompress)
	}

	if enforceMaxFileSizeLimit && compressedFileSize >= kurtosisDataTransferLimit {
		return "", 0, nil, stacktrace.NewError(
			"The files you are trying to upload, which are now compressed, exceed or reach 2 GB. " +
				"Manipulation (i.e. uploads or downloads) of files larger than 2 GB is currently disallowed in Kurtosis.")
	}

	return compressedFilePath, compressedFileSize, compressedFileContentMd5, nil
}

// listFilesInPathDeterministic returns the list of file paths in the path passed as an argument.
// If the path points to a file, it only returns the given file path.
// If recursiveMode is true, it recursively iterates over the directory inside the given path.
// If rules is non-nil, files/directories matching the ignore rules are excluded.
func listFilesInPathDeterministic(path string, recursiveMode bool, rules *ignoreRules) ([]string, error) {
	var listOfFiles []string
	err := listFilesInPathInternal(&listOfFiles, path, path, recursiveMode, true, rules)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Error listing files in '%s'", path)
	}
	sort.Strings(listOfFiles)
	return listOfFiles, nil
}

func listFilesInPathInternal(filesInPath *[]string, rootPath string, path string, recursiveMode bool, topLevel bool, rules *ignoreRules) error {
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

			// Check ignore rules
			if rules != nil {
				relativePath := strings.TrimPrefix(fileInDirectoryPath, strings.TrimRight(rootPath, string(filepath.Separator))+string(filepath.Separator))
				if rules.shouldIgnore(relativePath, fileInDirectory.IsDir()) {
					continue
				}
			}

			*filesInPath = append(*filesInPath, fileInDirectoryPath)
			if recursiveMode && fileInDirectory.IsDir() {
				err := listFilesInPathInternal(filesInPath, rootPath, fileInDirectoryPath, recursiveMode, false, rules)
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

// mapFilePathOnDiskToRelativePathInArchive returns a map of the form {filePathOnDisk: relativePath}
// where filePathOnDisk is the path of the file on disk and relativePath is the path of the file within the archive
func mapFilePathOnDiskToRelativePathInArchive(pathToCompress string, filesToUpload []string) map[string]string {
	filenameMappings := make(map[string]string)

	for _, filePath := range filesToUpload {
		relativePath := strings.TrimPrefix(filePath, pathToCompress)
		relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
		filenameMappings[filePath] = relativePath
	}

	return filenameMappings
}

// Unarchive extracts an archive to a destination directory.
// Unlike the previous arc.Unarchive implementation, directories are always
// created with 0755 permissions (matching the old mholt/archiver behavior).
// This prevents "permission denied" errors when archives contain directories
// with restrictive permissions (e.g. validator secrets with 0500).
func Unarchive(source, destination string) error {
	archiveFile, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open archive %s: %w", source, err)
	}
	defer archiveFile.Close()

	format, input, err := archives.Identify(context.Background(), source, archiveFile)
	if err != nil {
		return fmt.Errorf("identify format: %w", err)
	}

	extractor, ok := format.(archives.Extractor)
	if !ok {
		return fmt.Errorf("unsupported format for extraction")
	}

	if err := os.MkdirAll(destination, extractDirPermissions); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	handler := func(ctx context.Context, f archives.FileInfo) error {
		clean := filepath.Clean("/" + f.NameInArchive)
		relative := strings.TrimPrefix(clean, string(os.PathSeparator))
		dstPath := filepath.Join(destination, relative)

		if !strings.HasPrefix(filepath.Clean(dstPath)+string(os.PathSeparator), filepath.Clean(destination)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", dstPath)
		}

		if f.IsDir() {
			return os.MkdirAll(dstPath, extractDirPermissions)
		}

		if f.LinkTarget != "" {
			return nil
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), extractDirPermissions); err != nil {
			return fmt.Errorf("mkdir: %w", err)
		}

		reader, err := f.Open()
		if err != nil {
			return fmt.Errorf("open file: %w", err)
		}
		defer reader.Close()

		out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY, f.Mode())
		if err != nil {
			return fmt.Errorf("create file: %w", err)
		}
		defer out.Close()

		if _, err := io.Copy(out, reader); err != nil {
			return fmt.Errorf("copy: %w", err)
		}
		return nil
	}

	if err := extractor.Extract(context.Background(), input, handler); err != nil {
		return fmt.Errorf("extracting files: %w", err)
	}
	return nil
}
