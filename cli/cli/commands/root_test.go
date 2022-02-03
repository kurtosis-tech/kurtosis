package commands

import (
	"bytes"
	"github.com/kurtosis-tech/kurtosis-cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

const (
	temporaryTestFileFilename = "temporary-test-file"
)

func TestVersion(t *testing.T) {
	filepath, err := host_machine_directories.GetLatestCLIReleaseVersionCacheFilepath()
	require.NoError(t, err, "An error occurred getting the latest CLI release version cache file filepath")

	fileInfo, err := os.Stat(filepath)
	if !os.IsNotExist(err) {
		require.NoError(t, err, "An error occurred getting the latest CLI release version cache file info")
	}

	if fileInfo != nil {
		err = os.Remove(filepath)
		require.NoError(t, err, "An error occurred removing latest CLI release version cache file")
	}

	buf := new(bytes.Buffer)
	root := RootCmd
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, kurtosis_cli_version.KurtosisCLIVersion + "\n", buf.String())
}

func TestGetLatestCLIReleaseVersionFromCacheFile_CacheFileDoesNotExist(t *testing.T) {
	filepath, removeTempFileFunc, err := createNewTempFileAndGetFilepath()
	defer removeTempFileFunc()
	require.NoError(t, err, "An error occurred getting the cache file filepath for test")

	version, err := getLatestCLIReleaseVersionFromCacheFile(filepath)
	require.NoError(t, err, "An error occurred getting the latest CLI release version from cache file")

	assert.Empty(t, version)
}

func TestGetLatestCLIReleaseVersionFromCacheFile_SaveVersionInCacheFileAndGetVersionFromIt(t *testing.T) {
	filepath, removeTempFileFunc, err := createNewTempFileAndGetFilepath()
	defer removeTempFileFunc()
	require.NoError(t, err, "An error occurred getting the cache file filepath for test")

	versionForTest := "1.1.99"

	err = saveLatestCLIReleaseVersionInCacheFile(filepath, versionForTest)
	require.NoError(t, err, "An error occurred saving latest CLI release version for test in cache file for test")

	version, err := getLatestCLIReleaseVersionFromCacheFile(filepath)
	require.NoError(t, err, "An error occurred getting the latest CLI release version from cache file")

	assert.Equal(t, versionForTest, version)

	err = os.Remove(filepath)
	require.NoError(t, err, "An error occurred removing the cache file for test")
}

func createNewTempFileAndGetFilepath() (string, func() error, error) {

	tempFile, err := ioutil.TempFile("", temporaryTestFileFilename)
	if err != nil {
		return "", nil, stacktrace.Propagate(err, "An error occurred creating temporary file for test purpose with name '%v'", temporaryTestFileFilename)
	}
	removeTempFileFunc := func() error {
		if err := os.Remove(tempFile.Name()); err != nil {
			return stacktrace.Propagate(err, "An error occurred removing temporary file with name '%v'", temporaryTestFileFilename)
		}
		return nil
	}

	return tempFile.Name(), removeTempFileFunc, nil
}

// TODO More tests here, but have to figure out how to spin up a test engine that won't conflict with the real engine
