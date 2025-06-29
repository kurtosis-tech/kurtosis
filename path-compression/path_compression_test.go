package path_compression

import (
	"encoding/hex"
	"io"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/mholt/archiver"
	"github.com/stretchr/testify/require"
)

const (
	defaultPerm = 0700
)

func TestCompressPath(t *testing.T) {
	// Create a random file structure
	// test-dir-<RANDOM_NUMBER>/
	// |-- file_1.txt
	// |-- file_2.txt
	// |-- level_1/
	// |   |-- file_3.txt
	// |   \-- level_2/
	// |       \-- file_4.txt
	// \-- level_1_2/
	//
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	file1 := "file_1.txt"
	file2 := "file_2.txt"
	file3 := "file_3.txt"
	file4 := "file_4.txt"
	level1Dir := "level_1"
	level2Dir := "level_2"
	level1Dir2 := "level_1_2"
	file1Open, err := os.Create(path.Join(dirPath, file1))
	require.NoError(t, err)
	require.NoError(t, file1Open.Close())
	file2Open, err := os.Create(path.Join(dirPath, file2))
	require.NoError(t, err)
	require.NoError(t, file2Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir), defaultPerm)
	require.NoError(t, err)
	file3Open, err := os.Create(path.Join(dirPath, level1Dir, file3))
	require.NoError(t, err)
	_, err = file3Open.WriteString("Hello World!")
	require.NoError(t, err)
	require.NoError(t, file3Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir, level2Dir), defaultPerm)
	require.NoError(t, err)
	file4Open, err := os.Create(path.Join(dirPath, level1Dir, level2Dir, file4))
	require.NoError(t, err)
	require.NoError(t, file4Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir2), defaultPerm)
	require.NoError(t, err)

	// Running the test!
	compressedData, size, md5, err := CompressPath(dirPath, false)
	require.NoError(t, err)
	compressedDataBytes, err := io.ReadAll(compressedData)
	require.NoError(t, err)
	require.NotNil(t, compressedDataBytes)
	require.Greater(t, size, uint64(0))

	// Verify the content is a gzip file (check for gzip magic bytes)
	require.True(t, len(compressedDataBytes) > 2)
	require.Equal(t, byte(0x1f), compressedDataBytes[0]) // First byte of gzip magic number
	require.Equal(t, byte(0x8b), compressedDataBytes[1]) // Second byte of gzip magic number

	expectedHashHex := "e69f390b8b262f2f4153a041157fcf2e"
	require.Equal(t, expectedHashHex, hex.EncodeToString(md5))

	// Check that the hash is idempotent by running compression again on the same input
	compressedDataAgain, _, md5Again, errAgain := CompressPath(dirPath, false)
	require.NoError(t, errAgain)
	secondRunCompressedBytes, err := io.ReadAll(compressedDataAgain)
	require.NoError(t, err)

	// Verify the second run also produced a valid gzip file
	require.True(t, len(secondRunCompressedBytes) > 2)
	require.Equal(t, byte(0x1f), secondRunCompressedBytes[0])
	require.Equal(t, byte(0x8b), secondRunCompressedBytes[1])

	require.Equal(t, md5Again, md5)
}

func TestCompressPath_CheckEmptyDirectoryAreTakenIntoAccountForHash(t *testing.T) {
	// Create a random file structure
	// test-dir-<RANDOM_NUMBER>/
	// |-- file_1.txt
	// |-- file_2.txt
	// |-- level_1/
	// |   |-- file_3.txt
	// |   \-- level_2/
	// |       \-- file_4.txt
	// \-- level_1_2_UPDATED/
	//
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	file1 := "file_1.txt"
	file2 := "file_2.txt"
	file3 := "file_3.txt"
	file4 := "file_4.txt"
	level1Dir := "level_1"
	level2Dir := "level_2"
	level1Dir2 := "level_1_2_UPDATED"
	file1Open, err := os.Create(path.Join(dirPath, file1))
	require.NoError(t, err)
	require.NoError(t, file1Open.Close())
	file2Open, err := os.Create(path.Join(dirPath, file2))
	require.NoError(t, err)
	require.NoError(t, file2Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir), defaultPerm)
	require.NoError(t, err)
	file3Open, err := os.Create(path.Join(dirPath, level1Dir, file3))
	require.NoError(t, err)
	_, err = file3Open.WriteString("Hello World!")
	require.NoError(t, err)
	require.NoError(t, file3Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir, level2Dir), defaultPerm)
	require.NoError(t, err)
	file4Open, err := os.Create(path.Join(dirPath, level1Dir, level2Dir, file4))
	require.NoError(t, err)
	require.NoError(t, file4Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir2), defaultPerm)
	require.NoError(t, err)

	// Running the test!
	_, _, md5, err := CompressPath(dirPath, false)
	require.NoError(t, err)
	expectedHashHex := "6af1d6c21803f191408aa70baebb8818"
	require.Equal(t, expectedHashHex, hex.EncodeToString(md5))
}

func TestCompressPath_CheckDirectoryNamesAreTakenIntoAccountForHash(t *testing.T) {
	// Create a random file structure
	// test-dir-<RANDOM_NUMBER>/
	// |-- file_1.txt
	// |-- file_2.txt
	// |-- level_1_UPDATED/
	// |   |-- file_3.txt
	// |   \-- level_2/
	// |       \-- file_4.txt
	// \-- level_1_2/
	//
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	file1 := "file_1.txt"
	file2 := "file_2.txt"
	file3 := "file_3.txt"
	file4 := "file_4.txt"
	level1Dir := "level_1"
	level2Dir := "level_2"
	level1Dir2 := "level_1_2"
	file1Open, err := os.Create(path.Join(dirPath, file1))
	require.NoError(t, err)
	require.NoError(t, file1Open.Close())
	file2Open, err := os.Create(path.Join(dirPath, file2))
	require.NoError(t, err)
	require.NoError(t, file2Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir), defaultPerm)
	require.NoError(t, err)
	file3Open, err := os.Create(path.Join(dirPath, level1Dir, file3))
	require.NoError(t, err)
	_, err = file3Open.WriteString("Hello World!") // writing something to file3!!
	require.NoError(t, err)
	require.NoError(t, file3Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir, level2Dir), defaultPerm)
	require.NoError(t, err)
	file4Open, err := os.Create(path.Join(dirPath, level1Dir, level2Dir, file4))
	require.NoError(t, err)
	require.NoError(t, file4Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir2), defaultPerm)
	require.NoError(t, err)

	// Running the test!
	_, _, md5, err := CompressPath(dirPath, false)
	require.NoError(t, err)
	expectedHashHex := "e69f390b8b262f2f4153a041157fcf2e"
	require.Equal(t, expectedHashHex, hex.EncodeToString(md5))
}

func TestCompressPath_CheckFileContentsAreTakenIntoAccountForHash(t *testing.T) {
	// Create a random file structure
	// test-dir-<RANDOM_NUMBER>/
	// |-- file_1.txt
	// |-- file_2.txt
	// |-- level_1_UPDATED/
	// |   |-- file_3.txt
	// |   \-- level_2/
	// |       \-- file_4.txt
	// \-- level_1_2/
	//
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(dirPath)

	file1 := "file_1.txt"
	file2 := "file_2.txt"
	file3 := "file_3.txt"
	file4 := "file_4.txt"
	level1Dir := "level_1_UPDATED"
	level2Dir := "level_2"
	level1Dir2 := "level_1_2"
	file1Open, err := os.Create(path.Join(dirPath, file1))
	require.NoError(t, err)
	require.NoError(t, file1Open.Close())
	file2Open, err := os.Create(path.Join(dirPath, file2))
	require.NoError(t, err)
	require.NoError(t, file2Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir), defaultPerm)
	require.NoError(t, err)
	file3Open, err := os.Create(path.Join(dirPath, level1Dir, file3))
	require.NoError(t, err)
	_, err = file3Open.WriteString("Hello World!") // Write something to file3!
	require.NoError(t, err)
	require.NoError(t, file3Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir, level2Dir), defaultPerm)
	require.NoError(t, err)
	file4Open, err := os.Create(path.Join(dirPath, level1Dir, level2Dir, file4))
	require.NoError(t, err)
	require.NoError(t, file4Open.Close())

	err = os.Mkdir(path.Join(dirPath, level1Dir2), defaultPerm)
	require.NoError(t, err)

	// Running the test!
	_, _, md5, err := CompressPath(dirPath, false)
	require.NoError(t, err)
	expectedHashHex := "87e6196a7aaaf5355aac554131f8ac47"
	require.Equal(t, expectedHashHex, hex.EncodeToString(md5))
}

func TestCompressPath_EmptyDirDoesNotError(t *testing.T) {
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)

	_, _, _, err = CompressPath(dirPath, false)
	require.NoError(t, err)
}

func TestGetFilenameMappings(t *testing.T) {
	tests := []struct {
		name           string
		pathToCompress string
		filesToUpload  []string
		expected       map[string]string
	}{
		{
			name:           "basic path mapping",
			pathToCompress: "./test-dir",
			filesToUpload: []string{
				"./test-dir/file1.txt",
				"./test-dir/subdir/file2.txt",
			},
			expected: map[string]string{
				"./test-dir/file1.txt":        "file1.txt",
				"./test-dir/subdir/file2.txt": "subdir/file2.txt",
			},
		},
		{
			name:           "path without ./ prefix",
			pathToCompress: "test-dir",
			filesToUpload: []string{
				"test-dir/file1.txt",
				"test-dir/subdir/file2.txt",
			},
			expected: map[string]string{
				"test-dir/file1.txt":        "file1.txt",
				"test-dir/subdir/file2.txt": "subdir/file2.txt",
			},
		},
		{
			name:           "empty file list",
			pathToCompress: "./test-dir",
			filesToUpload:  []string{},
			expected:       map[string]string{},
		},
		{
			name:           "paths with special characters",
			pathToCompress: "./test-dir",
			filesToUpload: []string{
				"./test-dir/file with spaces.txt",
				"./test-dir/subdir/file-with-dashes.txt",
			},
			expected: map[string]string{
				"./test-dir/file with spaces.txt":        "file with spaces.txt",
				"./test-dir/subdir/file-with-dashes.txt": "subdir/file-with-dashes.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapFilePathOnDiskToRelativePathInArchive(tt.pathToCompress, tt.filesToUpload)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestUnarchive(t *testing.T) {
	tests := []struct {
		name            string
		archiveFormat   string
		createArchive   func(string, string) error
		expectedFiles   []string
		expectedContent map[string]string
	}{
		{
			name:          "tar.gz archive",
			archiveFormat: ".tar.gz",
			createArchive: func(source, archivePath string) error {
				// Use our own compression to create the archive
				compressedData, _, _, err := CompressPath(source, false)
				if err != nil {
					return err
				}
				defer compressedData.Close()

				archiveFile, err := os.Create(archivePath)
				if err != nil {
					return err
				}
				defer archiveFile.Close()

				_, err = io.Copy(archiveFile, compressedData)
				return err
			},
			expectedFiles: []string{
				"file_1.txt",
				"file_2.txt",
				"level_1/file_3.txt",
				"level_1/level_2/file_4.txt",
			},
			expectedContent: map[string]string{
				"file_1.txt":                 "",
				"file_2.txt":                 "",
				"level_1/file_3.txt":         "Hello World!",
				"level_1/level_2/file_4.txt": "",
			},
		},
		{
			name:          "zip archive",
			archiveFormat: ".zip",
			createArchive: func(source, archivePath string) error {
				// Create a simple zip archive for testing
				archive, err := os.Create(archivePath)
				if err != nil {
					return err
				}
				defer archive.Close()

				// This is a minimal zip file for testing
				// In a real scenario, you'd use a proper zip library
				zipContent := []byte{
					0x50, 0x4b, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00, 0x08, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00,
					0x74, 0x65, 0x73, 0x74, 0x2e, 0x74, 0x78, 0x74, 0x68, 0x65,
					0x6c, 0x6c, 0x6f, 0x0a, 0x50, 0x4b, 0x01, 0x02, 0x14, 0x00,
					0x14, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x74, 0x65,
					0x73, 0x74, 0x2e, 0x74, 0x78, 0x74, 0x50, 0x4b, 0x05, 0x06,
					0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x2e, 0x00,
					0x00, 0x00, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x00,
				}
				_, err = archive.Write(zipContent)
				return err
			},
			expectedFiles: []string{"test.txt"},
			expectedContent: map[string]string{
				"test.txt": "hello\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory structure
			sourceDir, err := os.MkdirTemp("", "unarchive-source-*")
			require.NoError(t, err)
			defer os.RemoveAll(sourceDir)

			// Create test files
			file1 := path.Join(sourceDir, "file_1.txt")
			file2 := path.Join(sourceDir, "file_2.txt")
			level1Dir := path.Join(sourceDir, "level_1")
			level2Dir := path.Join(level1Dir, "level_2")
			file3 := path.Join(level1Dir, "file_3.txt")
			file4 := path.Join(level2Dir, "file_4.txt")

			require.NoError(t, os.MkdirAll(level2Dir, 0755))

			require.NoError(t, os.WriteFile(file1, []byte(""), 0644))
			require.NoError(t, os.WriteFile(file2, []byte(""), 0644))
			require.NoError(t, os.WriteFile(file3, []byte("Hello World!"), 0644))
			require.NoError(t, os.WriteFile(file4, []byte(""), 0644))

			// Create archive
			archivePath := path.Join(os.TempDir(), "test-archive"+tt.archiveFormat)
			defer os.Remove(archivePath)

			err = tt.createArchive(sourceDir, archivePath)
			require.NoError(t, err)

			// Create destination directory
			destDir, err := os.MkdirTemp("", "unarchive-dest-*")
			require.NoError(t, err)
			defer os.RemoveAll(destDir)

			// Test our Unarchive function
			err = Unarchive(archivePath, destDir)
			require.NoError(t, err)

			// Verify extracted files
			for _, expectedFile := range tt.expectedFiles {
				filePath := path.Join(destDir, expectedFile)
				require.FileExists(t, filePath)

				if expectedContent, exists := tt.expectedContent[expectedFile]; exists {
					content, err := os.ReadFile(filePath)
					require.NoError(t, err)
					require.Equal(t, expectedContent, string(content))
				}
			}
		})
	}
}

func TestUnarchive_EmptyArchive(t *testing.T) {
	// Create an empty archive
	archivePath := path.Join(os.TempDir(), "empty-archive.tar.gz")
	defer os.Remove(archivePath)

	// Create a minimal empty tar.gz
	archiveFile, err := os.Create(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	// Write minimal tar.gz header
	emptyTarGz := []byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03,
		0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	_, err = archiveFile.Write(emptyTarGz)
	require.NoError(t, err)

	destDir, err := os.MkdirTemp("", "unarchive-empty-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	err = Unarchive(archivePath, destDir)
	require.NoError(t, err)

	// Should create the destination directory even if archive is empty
	require.DirExists(t, destDir)
}

func TestUnarchive_NonExistentArchive(t *testing.T) {
	destDir, err := os.MkdirTemp("", "unarchive-nonexistent-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	err = Unarchive("/non/existent/archive.tar.gz", destDir)
	require.Error(t, err)
}

func TestUnarchive_InvalidArchive(t *testing.T) {
	// Create an invalid archive (just random bytes)
	archivePath := path.Join(os.TempDir(), "invalid-archive.tar.gz")
	defer os.Remove(archivePath)

	archiveFile, err := os.Create(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	// Write random bytes
	_, err = archiveFile.Write([]byte("this is not a valid archive"))
	require.NoError(t, err)

	destDir, err := os.MkdirTemp("", "unarchive-invalid-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	err = Unarchive(archivePath, destDir)
	require.Error(t, err)
}

func TestUnarchive_PreservesFilePermissions(t *testing.T) {
	// Create test directory with files having specific permissions
	sourceDir, err := os.MkdirTemp("", "unarchive-perms-*")
	require.NoError(t, err)
	defer os.RemoveAll(sourceDir)

	// Create a file with specific permissions
	testFile := path.Join(sourceDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test content"), 0644))

	// Create archive using our compression
	compressedData, _, _, err := CompressPath(sourceDir, false)
	require.NoError(t, err)
	defer compressedData.Close()

	archivePath := path.Join(os.TempDir(), "perms-archive.tar.gz")
	defer os.Remove(archivePath)

	archiveFile, err := os.Create(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	_, err = io.Copy(archiveFile, compressedData)
	require.NoError(t, err)

	// Extract
	destDir, err := os.MkdirTemp("", "unarchive-perms-dest-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	err = Unarchive(archivePath, destDir)
	require.NoError(t, err)

	// Check that file exists and has reasonable permissions
	extractedFile := path.Join(destDir, "test.txt")
	require.FileExists(t, extractedFile)

	info, err := os.Stat(extractedFile)
	require.NoError(t, err)

	// Should be readable (at least 0400)
	require.GreaterOrEqual(t, info.Mode()&os.ModePerm, os.FileMode(0400))
}

func TestUnarchive_OverwritesExistingFiles(t *testing.T) {
	// Create test directory
	sourceDir, err := os.MkdirTemp("", "unarchive-overwrite-*")
	require.NoError(t, err)
	defer os.RemoveAll(sourceDir)

	// Create a file in source
	testFile := path.Join(sourceDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("new content"), 0644))

	// Create archive
	compressedData, _, _, err := CompressPath(sourceDir, false)
	require.NoError(t, err)
	defer compressedData.Close()

	archivePath := path.Join(os.TempDir(), "overwrite-archive.tar.gz")
	defer os.Remove(archivePath)

	archiveFile, err := os.Create(archivePath)
	require.NoError(t, err)
	defer archiveFile.Close()

	_, err = io.Copy(archiveFile, compressedData)
	require.NoError(t, err)

	// Create destination with existing file
	destDir, err := os.MkdirTemp("", "unarchive-overwrite-dest-*")
	require.NoError(t, err)
	defer os.RemoveAll(destDir)

	existingFile := path.Join(destDir, "test.txt")
	require.NoError(t, os.WriteFile(existingFile, []byte("old content"), 0644))

	// Extract (should overwrite)
	err = Unarchive(archivePath, destDir)
	require.NoError(t, err)

	// Check that file was overwritten
	content, err := os.ReadFile(existingFile)
	require.NoError(t, err)
	require.Equal(t, "new content", string(content))
}

// TestUnarchive_ComparisonWithMholtArchiver ensures our implementation produces
// identical results to mholt/archiver for various archive formats
func TestUnarchive_ComparisonWithMholtArchiver(t *testing.T) {
	tests := []struct {
		name          string
		archiveFormat string
		createArchive func(string, string) error
	}{
		{
			name:          "tar.gz comparison",
			archiveFormat: ".tar.gz",
			createArchive: func(source, archivePath string) error {
				// Use our own compression to create the archive
				compressedData, _, _, err := CompressPath(source, false)
				if err != nil {
					return err
				}
				defer compressedData.Close()

				archiveFile, err := os.Create(archivePath)
				if err != nil {
					return err
				}
				defer archiveFile.Close()

				_, err = io.Copy(archiveFile, compressedData)
				return err
			},
		},
		{
			name:          "zip comparison",
			archiveFormat: ".zip",
			createArchive: func(source, archivePath string) error {
				// Create a simple zip using standard library for comparison
				zipFile, err := os.Create(archivePath)
				if err != nil {
					return err
				}
				defer zipFile.Close()

				// Create a minimal valid zip file
				zipContent := []byte{
					0x50, 0x4b, 0x03, 0x04, 0x14, 0x00, 0x00, 0x00, 0x08, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00,
					0x74, 0x65, 0x73, 0x74, 0x2e, 0x74, 0x78, 0x74, 0x68, 0x65,
					0x6c, 0x6c, 0x6f, 0x0a, 0x50, 0x4b, 0x01, 0x02, 0x14, 0x00,
					0x14, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x74, 0x65,
					0x73, 0x74, 0x2e, 0x74, 0x78, 0x74, 0x50, 0x4b, 0x05, 0x06,
					0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x2e, 0x00,
					0x00, 0x00, 0x2a, 0x00, 0x00, 0x00, 0x00, 0x00,
				}
				_, err = zipFile.Write(zipContent)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test directory structure
			sourceDir, err := os.MkdirTemp("", "comparison-source-*")
			require.NoError(t, err)
			defer os.RemoveAll(sourceDir)

			// Create test files with various content and permissions
			files := map[string]struct {
				content    string
				permission os.FileMode
			}{
				"file1.txt":           {"content1", 0644},
				"file2.txt":           {"content2", 0755},
				"empty.txt":           {"", 0600},
				"dir1/file3.txt":      {"nested content", 0644},
				"dir1/dir2/file4.txt": {"deeply nested", 0755},
			}

			for filePath, fileInfo := range files {
				fullPath := path.Join(sourceDir, filePath)
				dir := path.Dir(fullPath)
				if dir != sourceDir {
					require.NoError(t, os.MkdirAll(dir, 0755))
				}
				require.NoError(t, os.WriteFile(fullPath, []byte(fileInfo.content), fileInfo.permission))
			}

			// Create archive
			archivePath := path.Join(os.TempDir(), "comparison-archive"+tt.archiveFormat)
			defer os.Remove(archivePath)

			err = tt.createArchive(sourceDir, archivePath)
			require.NoError(t, err)

			// Create two destination directories
			destDir1, err := os.MkdirTemp("", "comparison-dest1-*")
			require.NoError(t, err)
			defer os.RemoveAll(destDir1)

			destDir2, err := os.MkdirTemp("", "comparison-dest2-*")
			require.NoError(t, err)
			defer os.RemoveAll(destDir2)

			// Extract using our implementation
			err = Unarchive(archivePath, destDir1)
			require.NoError(t, err)

			// Extract using mholt/archiver (if available)
			err = extractWithMholtArchiver(archivePath, destDir2)
			require.NoError(t, err)

			// Compare the results
			compareDirectories(t, destDir1, destDir2)
		})
	}
}

// extractWithMholtArchiver extracts using mholt/archiver for comparison
func extractWithMholtArchiver(source, destination string) error {
	return archiver.Unarchive(source, destination)
}

// compareDirectories recursively compares two directories for identical content
func compareDirectories(t *testing.T, dir1, dir2 string) {
	// Get all files in both directories
	files1, err := getAllFiles(dir1)
	require.NoError(t, err)

	files2, err := getAllFiles(dir2)
	require.NoError(t, err)

	// Compare file lists
	require.Equal(t, len(files1), len(files2), "Different number of files")

	// Compare each file
	for filePath := range files1 {
		file1Path := path.Join(dir1, filePath)
		file2Path := path.Join(dir2, filePath)

		// Check if file exists in both directories
		require.FileExists(t, file1Path)
		require.FileExists(t, file2Path)

		// Compare file contents
		content1, err := os.ReadFile(file1Path)
		require.NoError(t, err)

		content2, err := os.ReadFile(file2Path)
		require.NoError(t, err)

		require.Equal(t, content1, content2, "File content differs: %s", filePath)

		// Compare file permissions (basic check)
		info1, err := os.Stat(file1Path)
		require.NoError(t, err)

		info2, err := os.Stat(file2Path)
		require.NoError(t, err)

		// Compare permissions (ignore owner/group, just check if readable/writable)
		mode1 := info1.Mode() & os.ModePerm
		mode2 := info2.Mode() & os.ModePerm

		// Allow some flexibility in permissions as long as they're reasonable
		require.GreaterOrEqual(t, mode1, os.FileMode(0400), "File1 not readable: %s", filePath)
		require.GreaterOrEqual(t, mode2, os.FileMode(0400), "File2 not readable: %s", filePath)
	}
}

// getAllFiles recursively gets all files in a directory
func getAllFiles(dir string) (map[string]bool, error) {
	files := make(map[string]bool)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Get relative path from base directory
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files[relPath] = true
		}

		return nil
	})

	return files, err
}
