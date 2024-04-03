package git_package_content_provider

import (
	"fmt"
	"os"
	"path"
	"strings"
)

const (
	githubUserTokenFilePath = "token.txt"
	githubTokenFilePerms    = 0644
)

type GitHubPackageAuthProvider struct {
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
	return nil
}

func (gitAuth *GitHubPackageAuthProvider) GetGitHubTokenForPackage(packageId string) string {
	tokenBytes, err := os.ReadFile(path.Join(gitAuth.githubAuthStorageDirPath, getGitHubTokenFileName(packageId)))
	if err != nil {
		return ""
	}
	return string(tokenBytes)
}

func getGitHubTokenFileName(packageId string) string {
	return fmt.Sprintf("%v.txt", strings.Replace(packageId, "/", "-", -1))
}

func (gitAuth *GitHubPackageAuthProvider) GetGitHubTokenForUser() string {
	tokenBytes, err := os.ReadFile(path.Join(gitAuth.githubAuthStorageDirPath, githubUserTokenFilePath))
	if err != nil {
		return ""
	}
	return string(tokenBytes)
}
