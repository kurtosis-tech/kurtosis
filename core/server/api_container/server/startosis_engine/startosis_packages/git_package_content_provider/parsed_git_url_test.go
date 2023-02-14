package git_package_content_provider

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const (
	testModuleAuthor                                       = "kurtosis-tech"
	testModuleName                                         = "sample-startosis-load"
	testFileName                                           = "sample.star"
	githubSampleURL                                        = "github.com/" + testModuleAuthor + "/" + testModuleName + "/" + testFileName
	githubSampleUrlWithTag                                 = githubSampleURL + "@5.33.2"
	githubSampleUrlWithBranchContainingVersioningDelimiter = githubSampleURL + "@my@favorite-branch"
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
		emptyTagBranchOrCommit,
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

func TestParsedGitURL_IfNoFileThenRelativeFilePathIsEmpty(t *testing.T) {
	pathWithoutFile := "github.com/" + testModuleAuthor + "/" + testModuleName
	parsedURL, err := parseGitURL(pathWithoutFile)
	require.Nil(t, err)
	require.Equal(t, "", parsedURL.relativeFilePath)
}

func TestParsedGitURL_ParsingGetsRidOfAnyPathEscapes(t *testing.T) {
	escapedURLWithoutStartosisFile := "github.com/../etc/passwd"
	parsedURL, err := parseGitURL(escapedURLWithoutStartosisFile)
	require.Nil(t, err)
	require.Equal(t, "", parsedURL.relativeFilePath)

	escapedURLWithStartosisFile := "github.com/../../etc/passwd/startosis.star"
	parsedURL, err = parseGitURL(escapedURLWithStartosisFile)
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

	escapedURLWithStartosisFile = "github.com/foo/../etc/../passwd"
	_, err = parseGitURL(escapedURLWithStartosisFile)
	require.NotNil(t, err)
	expectedErrorMsg := fmt.Sprintf("Error parsing the URL of module: '%s'. The path should contain at least 2 subpaths got '[passwd]'", escapedURLWithStartosisFile)
	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_WorksWithVersioningInformation(t *testing.T) {
	parsedURL, err := parseGitURL(githubSampleUrlWithTag)
	require.Nil(t, err)

	expectedParsedURL := newParsedGitURL(
		testModuleAuthor,
		testModuleName,
		fmt.Sprintf("https://github.com/%v/%v.git", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v/%v", testModuleAuthor, testModuleName, testFileName),
		"5.33.2",
	)

	require.Equal(t, expectedParsedURL, parsedURL)

	parsedURL, err = parseGitURL(githubSampleUrlWithBranchContainingVersioningDelimiter)
	require.Nil(t, err)

	expectedParsedURL = newParsedGitURL(
		testModuleAuthor,
		testModuleName,
		fmt.Sprintf("https://github.com/%v/%v.git", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v", testModuleAuthor, testModuleName),
		fmt.Sprintf("%v/%v/%v", testModuleAuthor, testModuleName, testFileName),
		"my@favorite-branch",
	)

	require.Equal(t, expectedParsedURL, parsedURL)
}
