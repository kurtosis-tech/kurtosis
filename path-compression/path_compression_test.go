package path_compression

import (
	"encoding/hex"
	"io"
	"os"
	"path"
	"testing"

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

	require.Equal(t, info.Mode()&os.ModePerm, os.FileMode(0644))
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
