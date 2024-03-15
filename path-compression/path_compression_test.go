package path_compression

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path"
	"testing"
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
	expectedHashHex := "e69f390b8b262f2f4153a041157fcf2e"
	require.Equal(t, expectedHashHex, hex.EncodeToString(md5))

	// Check that the hash is idempotent
	compressedDataAgain, sizeAgain, md5Again, errAgain := CompressPath(dirPath, false)
	require.NoError(t, errAgain)
	compressedDataBytesAgain, err := io.ReadAll(compressedDataAgain)
	require.NoError(t, err)
	require.Equal(t, compressedDataBytes, compressedDataBytesAgain)
	require.Equal(t, sizeAgain, size)
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

func TestCompressPath_EmptyDirError(t *testing.T) {
	dirPath, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)

	_, _, _, err = CompressPath(dirPath, false)
	require.NoError(t, err)
}
