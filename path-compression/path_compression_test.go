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
