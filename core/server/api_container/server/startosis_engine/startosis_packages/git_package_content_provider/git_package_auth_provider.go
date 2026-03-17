package git_package_content_provider

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
)

const (
	// filename for storing a github auth token for a user
	githubUserTokenFilename = "token.txt"

	githubTokenFilePerms = 0644
)

type GitHubPackageAuthProvider struct {
	// Location inside APIC where GitHub auth info exists
	githubAuthStorageDirPath string
}

func NewGitHubPackageAuthProvider(githubAuthStorageDirPath string) *GitHubPackageAuthProvider {
	return &GitHubPackageAuthProvider{
		githubAuthStorageDirPath: githubAuthStorageDirPath,
	}
}

func (gitAuth *GitHubPackageAuthProvider) StoreGitHubTokenForPackage(packageId, token string) error {
	err := os.WriteFile(path.Join(gitAuth.githubAuthStorageDirPath, getGitHubTokenFileName(packageId)), []byte(token), githubTokenFilePerms)
	if err != nil {
		return err
	}
	logrus.Infof("Successfully stored GitHub auth token for package: '%v'", packageId)
	return nil
}

func (gitAuth *GitHubPackageAuthProvider) GetGitHubTokenForPackage(packageId string) string {
	tokenBytes, err := os.ReadFile(path.Join(gitAuth.githubAuthStorageDirPath, getGitHubTokenFileName(packageId)))
	if err != nil {
		return ""
	}
	logrus.Infof("Retrieved GitHub auth token for package: '%v'", packageId)
	return string(tokenBytes)
}

// store as <package id>.txt, which is usually github locator, so swap out the slashes
func getGitHubTokenFileName(packageId string) string {
	return fmt.Sprintf("%v.txt", strings.ReplaceAll(packageId, "/", "-"))
}

func (gitAuth *GitHubPackageAuthProvider) GetGitHubTokenForUser() string {
	tokenBytes, err := os.ReadFile(path.Join(gitAuth.githubAuthStorageDirPath, githubUserTokenFilename))
	if err != nil {
		return ""
	}
	logrus.Infof("Retrieved a GitHub auth token.")
	return string(tokenBytes)
}
