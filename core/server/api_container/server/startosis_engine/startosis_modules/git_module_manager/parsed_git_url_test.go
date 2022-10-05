package git_module_manager

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	moduleAuthor    = "kurtosis-tech"
	moduleName      = "sample-startosis-load"
	fileName        = "sample.star"
	githubSampleURL = "github.com/" + moduleAuthor + "/" + moduleName + "/" + fileName
)

func TestParsedGitURL_SimpleParse(t *testing.T) {
	parsedURL, err := parseGitURL(githubSampleURL)
	require.Nil(t, err)

	expectedParsedURL := newParsedGitURL(
		moduleAuthor,
		moduleName,
		fmt.Sprintf("https://github.com/%v/%v.git", moduleAuthor, moduleName),
		fmt.Sprintf("%v/%v", moduleAuthor, moduleName),
		fmt.Sprintf("%v/%v/%v", moduleAuthor, moduleName, fileName),
	)

	require.Equal(t, expectedParsedURL, parsedURL)
}

func TestParsedGitURL_FailsOnNonGithubURL(t *testing.T) {
	nonGithubURL := "kurtosis-git.com/" + moduleAuthor + "/" + moduleName + "/" + fileName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := "We only support modules on Github for now"

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsOnNonNonEmptySchema(t *testing.T) {
	ftpSchema := "ftp"
	nonGithubURL := ftpSchema + "://github.com/" + moduleAuthor + "/" + moduleName + "/" + fileName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := fmt.Sprintf("Expected schema to be empty got '%v'", ftpSchema)

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsWithoutPathToFile(t *testing.T) {
	nonGithubURL := "github.com/" + moduleAuthor + "/" + moduleName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := "URL path should contain at least 3 subpaths"

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsForNonStartosisFile(t *testing.T) {
	nonGithubURL := "github.com/" + moduleAuthor + "/" + moduleName + "/foo.srt"
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := fmt.Sprintf("Expected last subpath to be a '%v' file but it wasn't", startosisFileExtension)

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_ParsingGetsRidOfAnyPathEscapes(t *testing.T) {
	escapedURLWithoutStartosisFile := "github.com/../etc/passwd"
	_, err := parseGitURL(escapedURLWithoutStartosisFile)
	require.NotNil(t, err)
	expectedErrorMsg := "URL path should contain at least 3 subpath"
	require.Contains(t, err.Error(), expectedErrorMsg)

	escapedURLWithStartosisFile := "github.com/../../etc/passwd/startosis.star"
	parsedURL, err := parseGitURL(escapedURLWithStartosisFile)
	require.Nil(t,  err)
	require.Equal(t, parsedURL.moduleAuthor, "etc")
	require.Equal(t, parsedURL.moduleName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeModulePath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/passwd/startosis.star"
	parsedURL, err = parseGitURL(escapedURLWithStartosisFile)
	require.Nil(t,  err)
	require.Equal(t, parsedURL.moduleAuthor, "etc")
	require.Equal(t, parsedURL.moduleName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeModulePath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/../passwd/startosis.star"
	_, err = parseGitURL(escapedURLWithStartosisFile)
	require.NotNil(t,  err)
	expectedErrorMsg = "URL path should contain at least 3 subpath"
	require.Contains(t, err.Error(), expectedErrorMsg)
}
