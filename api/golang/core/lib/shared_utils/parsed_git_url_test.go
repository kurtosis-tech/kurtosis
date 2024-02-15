package shared_utils

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
	githubSampleUrlWithVersionWithSlash                    = "github.com/kurtosis-tech/sample-startosis-load/sample.star@foo/bar"
	githubSampleUrlWithVersionWithSlashAndFile             = "github.com/kurtosis-tech/sample-startosis-load@foo/bar/main.star"
)

func TestParsedGitURL_SimpleParse(t *testing.T) {
	parsedURL, err := ParseGitURL(githubSampleURL)
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
	_, err := ParseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := "We only support modules on Github for now"

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_FailsOnNonNonEmptySchema(t *testing.T) {
	ftpSchema := "ftp"
	nonGithubURL := ftpSchema + "://github.com/" + testModuleAuthor + "/" + testModuleName + "/" + testFileName
	_, err := ParseGitURL(nonGithubURL)
	require.NotNil(t, err)

	expectedErrorMsg := fmt.Sprintf("Expected schema to be empty got '%v'", ftpSchema)

	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_IfNoFileThenRelativeFilePathIsEmpty(t *testing.T) {
	pathWithoutFile := "github.com/" + testModuleAuthor + "/" + testModuleName
	parsedURL, err := ParseGitURL(pathWithoutFile)
	require.Nil(t, err)
	require.Equal(t, "", parsedURL.relativeFilePath)
}

func TestParsedGitURL_ParsingGetsRidOfAnyPathEscapes(t *testing.T) {
	escapedURLWithoutStartosisFile := "github.com/../etc/passwd"
	parsedURL, err := ParseGitURL(escapedURLWithoutStartosisFile)
	require.Nil(t, err)
	require.Equal(t, "", parsedURL.relativeFilePath)

	escapedURLWithStartosisFile := "github.com/../../etc/passwd/startosis.star"
	parsedURL, err = ParseGitURL(escapedURLWithStartosisFile)
	require.Nil(t, err)
	require.Equal(t, parsedURL.repositoryAuthor, "etc")
	require.Equal(t, parsedURL.repositoryName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeRepoPath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/passwd/startosis.star"
	parsedURL, err = ParseGitURL(escapedURLWithStartosisFile)
	require.Nil(t, err)
	require.Equal(t, parsedURL.repositoryAuthor, "etc")
	require.Equal(t, parsedURL.repositoryName, "passwd")
	require.Equal(t, parsedURL.gitURL, "https://github.com/etc/passwd.git")
	require.Equal(t, parsedURL.relativeFilePath, "etc/passwd/startosis.star")
	require.Equal(t, parsedURL.relativeRepoPath, "etc/passwd")

	escapedURLWithStartosisFile = "github.com/foo/../etc/../passwd"
	_, err = ParseGitURL(escapedURLWithStartosisFile)
	require.NotNil(t, err)
	expectedErrorMsg := fmt.Sprintf("Error parsing the URL of module: '%s'. The path should contain at least 2 subpaths got '[passwd]'", escapedURLWithStartosisFile)
	require.Contains(t, err.Error(), expectedErrorMsg)
}

func TestParsedGitURL_WorksWithVersioningInformation(t *testing.T) {
	parsedURL, err := ParseGitURL(githubSampleUrlWithTag)
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

	parsedURL, err = ParseGitURL(githubSampleUrlWithBranchContainingVersioningDelimiter)
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

func TestParsedGitUrl_ResolvesRelativeUrl(t *testing.T) {
	parsedUrl, err := ParseGitURL(githubSampleURL)
	require.Nil(t, err)

	relativeUrl := "./lib.star"
	absoluteUrl := parsedUrl.GetAbsoluteLocatorRelativeToThisURL(relativeUrl)
	require.Nil(t, err)
	expected := "github.com/kurtosis-tech/sample-startosis-load/lib.star"
	require.Equal(t, expected, absoluteUrl)

	relativeUrl = "./src/lib.star"
	absoluteUrl = parsedUrl.GetAbsoluteLocatorRelativeToThisURL(relativeUrl)
	require.Nil(t, err)
	expected = "github.com/kurtosis-tech/sample-startosis-load/src/lib.star"
	require.Equal(t, expected, absoluteUrl)
}

func TestParsedGitUrl_ResolvesRelativeUrlForUrlWithTag(t *testing.T) {
	parsedUrl, err := ParseGitURL(githubSampleUrlWithTag)
	require.Nil(t, err)

	relativeUrl := "./lib.star"
	absoluteUrl := parsedUrl.GetAbsoluteLocatorRelativeToThisURL(relativeUrl)
	require.Nil(t, err)
	expected := "github.com/kurtosis-tech/sample-startosis-load/lib.star"
	require.Equal(t, expected, absoluteUrl)

	relativeUrl = "./src/lib.star"
	absoluteUrl = parsedUrl.GetAbsoluteLocatorRelativeToThisURL(relativeUrl)
	require.Nil(t, err)
	expected = "github.com/kurtosis-tech/sample-startosis-load/src/lib.star"
	require.Equal(t, expected, absoluteUrl)
}

func TestParsedGitUrl_ResolvesWithUrlWithVersionBranchWithSlash(t *testing.T) {
	parsedUrl, err := ParseGitURL(githubSampleUrlWithVersionWithSlash)
	require.Nil(t, err)

	require.Equal(t, "foo/bar", parsedUrl.tagBranchOrCommit)

	parsedUrl, err = ParseGitURL(githubSampleUrlWithVersionWithSlashAndFile)
	require.Nil(t, err)
	require.Equal(t, "foo/bar", parsedUrl.tagBranchOrCommit)
	require.Equal(t, "kurtosis-tech/sample-startosis-load/main.star", parsedUrl.relativeFilePath)
}
