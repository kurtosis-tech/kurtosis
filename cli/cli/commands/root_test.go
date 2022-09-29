package commands

import (
	"bytes"
	"github.com/kurtosis-tech/kurtosis-cli/cli/kurtosis_cli_version"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
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
	buf := new(bytes.Buffer)

	root := RootCmd
	root.SetOut(buf)
	root.SetErr(os.Stderr) // We do this because we don't want any "you're using an out-of-date version of the CLI" to fail this test
	root.SetArgs([]string{"version"})

	err := root.Execute()
	require.NoError(t, err)

	assert.Equal(t, kurtosis_cli_version.KurtosisCLIVersion + "\n", buf.String())
}

func TestGetLatestCLIReleaseVersionFromCacheFile_CacheFileDoesNotExist(t *testing.T) {
	filepath, removeTempFileFunc, err := createNewTempFileAndGetFilepath()
	defer func() {
		if err = removeTempFileFunc(); err != nil {
			logrus.Warnf("Error removing temporary file during test\n'%v'", err)
		}
	}()
	require.NoError(t, err, "An error occurred getting the cache file filepath for test")

	version, err := getLatestCLIReleaseVersionFromCacheFile(filepath)
	require.NoError(t, err, "An error occurred getting the latest CLI release version from cache file")

	assert.Empty(t, version)
}

func TestGetLatestCLIReleaseVersionFromCacheFile_SaveVersionInCacheFileAndGetVersionFromIt(t *testing.T) {
	filepath, removeTempFileFunc, err := createNewTempFileAndGetFilepath()
	defer func() {
		if err = removeTempFileFunc(); err != nil {
			logrus.Warnf("Error removing temporary file during test\n'%v'", err)
		}
	}()
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

func TestParseVersionStrToSemVer_CanParseValidVersions(t *testing.T){
	validStrictSemanticVersionStr := "0.23.1" //Semantic Versioning 2.0.0

	validPrefixedSemanticVersionStr := "v0.23.1" //Semantic Versioning 1.0.0 (tagging specification)

	invalidSemanticVersionStr := "v0.23.1v"

	_, err := parseVersionStrToSemVer(validStrictSemanticVersionStr)
	require.NoError(t, err, "The version string '%' can't be parsed to a valid semantic version", validStrictSemanticVersionStr)

	_, err = parseVersionStrToSemVer(validPrefixedSemanticVersionStr)
	require.NoError(t, err, "The version string '%' can't be parsed to a valid semantic version", validPrefixedSemanticVersionStr)

	_, err = parseVersionStrToSemVer(invalidSemanticVersionStr)
	require.Error(t, err, "The version string '%' was successfully parse and it is wrong because it is an invalid version string", invalidSemanticVersionStr)
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
