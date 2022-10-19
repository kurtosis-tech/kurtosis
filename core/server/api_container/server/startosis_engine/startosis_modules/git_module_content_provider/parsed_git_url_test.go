package git_module_content_provider

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testModuleAuthor = "kurtosis-tech"
	testModuleName   = "sample-startosis-load"
	testFileName     = "sample.star"
	githubSampleURL  = "github.com/" + testModuleAuthor + "/" + testModuleName + "/" + testFileName
)

func TestParsedGitURL_SimpleParse(t *testing.T) {
	parsedURL, err := parseGitURL(githubSampleURL)
	require.Nil(t, err)

	expectedParsedURL := newParsedGitURL(
		testModuleAuthor,
		testModuleName,
		fmt.Sprintf("https://github.com/%v/%v.git", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v/%v", testModuleAuthor, testModuleName, testFileName),
	)

	require.Equal(t, expectedParsedURL, parsedURL)
}

func TestParsedGitURL_FailsOnNonGithubURL(t *testing.T) {
	nonGithubURL := "kurtosis-git.com/" + testModuleAuthor + "/" + testModuleName + "/" + testFileName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := "We only support modules on Github for now"

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsOnNonNonEmptySchema(t *testing.T) {
	ftpSchema := "ftp"
	nonGithubURL := ftpSchema + "://github.com/" + testModuleAuthor + "/" + testModuleName + "/" + testFileName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := fmt.Sprintf("Expected schema to be empty got '%v'", ftpSchema)

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsWithoutPathToFile(t *testing.T) {
	nonGithubURL := "github.com/" + testModuleAuthor + "/" + testModuleName
	_, err := parseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := fmt.Sprintf("URL '%v' path should contain at least 3 subpaths got '[%v %v]'", nonGithubURL, testModuleAuthor, testModuleName)

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_ParsingGetsRidOfAnyPathEscapes(t *testing.T) {
	escapedURLWithoutStartosisFile := "github.com/../etc/passwd"
	_, err := parseGitURL(escapedURLWithoutStartosisFile)
	require.NotNil(t, err)
	expectedErrorMsg := fmt.Sprintf("URL '%v' path should contain at least 3 subpaths got '[etc passwd]'", escapedURLWithoutStartosisFile)
	require.Contains(t, err.Error(), expectedErrorMsg)

	escapedURLWithStartosisFile := "github.com/../../etc/passwd/startosis.star"
	parsedURL, err := parseGitURL(escapedURLWithStartosisFile)
	require.Nil(t, err)
	require.Equal(t, parsedURL.moduleAuthor, "etc")
	require.Equal(t, parsedURL.moduleName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeRepoPath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/passwd/startosis.star"
	parsedURL, err = parseGitURL(escapedURLWithStartosisFile)
	require.Nil(t, err)
	require.Equal(t, parsedURL.moduleAuthor, "etc")
	require.Equal(t, parsedURL.moduleName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeRepoPath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/../passwd/startosis.star"
	_, err = parseGitURL(escapedURLWithStartosisFile)
	require.NotNil(t, err)
	expectedErrorMsg = fmt.Sprintf("URL '%v' path should contain at least 3 subpaths got '[passwd startosis.star]'", escapedURLWithStartosisFile)
	require.Contains(t, err.Error(), expectedErrorMsg)
}
