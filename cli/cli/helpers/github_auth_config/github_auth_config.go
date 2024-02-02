package github_auth_config

import (
	"errors"
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
	"os"
	"sync"
)

const (
	// TODO: check what file permissions are needed
	githubUsernameFilePermission  os.FileMode = 0644
	githubAuthTokenFilePermission os.FileMode = 0644

	kurtosisCliKeyringServiceName = "kurtosis-cli"
)

// GitHubAuthConfig represents information regarding a GitHub user authorized with Kurtosis CLI, if one exists
// Only one user can be logged in at a time
type GitHubAuthConfig struct {
	mutex *sync.RWMutex

	// Empty string if no user is logged in
	username string
}

func GetGitHubAuthConfig() (*GitHubAuthConfig, error) {
	// Existence of GitHub user filepath determines whether a user is logged in or not
	var username string
	isUserLoggedIn, err := doesGitHubUsernameFileExist()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred verifying existence of a GitHub user.")
	}
	if isUserLoggedIn {
		username, err = getGitHubUsernameFromFile()
		if err != nil {
			return nil, stacktrace.Propagate(err, "GitHub user found but error occurred getting username.")
		}
		// TODO: verify an auth token exists
	}
	return &GitHubAuthConfig{
		mutex:    &sync.RWMutex{},
		username: username,
	}, nil
}

func (git *GitHubAuthConfig) Login() error {
	git.mutex.Lock()
	defer git.mutex.Unlock()

	// Don't log in if user already exists
	if git.isLoggedIn() {
		return nil
	}

	authToken, userLogin, err := AuthFlow("github.com", "", []string{}, true, *browser.New("", os.Stdout, os.Stderr))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred in the Github OAuth flow.")
	}

	err = saveGitHubUsernameFile(userLogin)
	shouldRemoveUsernameFile := true
	defer func() {
		if shouldRemoveUsernameFile {
			err := removeGitHubUsernameFile()
			if err != nil {
				logrus.Errorf("Failed to remove github username file after initial login failed!!! GitHub auth could be in a bad state.")
			}
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred saving github username file fo user: %v.", userLogin)
	}

	err = setAuthToken(userLogin, authToken)
	shouldRemoveAuthToken := true
	defer func() {
		if shouldRemoveAuthToken {
			err = removeAuthToken(git.username)
			if err != nil {
				logrus.Errorf("Failed to remove git auth token after setting the auth token failed!! GitHub auth could be in a bad state.")
			}
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting auth token for user: %v.", userLogin)
	}

	git.username = userLogin

	shouldRemoveAuthToken = false
	shouldRemoveUsernameFile = false
	return nil
}

// Logout "logs out" a GitHub user from Kurtosis by:
// -removing the auth token associated with [username] from keyring or plain text file
// -removing the GitHub username file
// -setting [username] to empty string
func (git *GitHubAuthConfig) Logout() error {
	// Don't log out if no user exists
	if git.isLoggedIn() {
		return nil
	}
	err := removeAuthToken(git.username)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing auth token for GitHub user: '%v'", git.username)
	}
	err = removeGitHubUsernameFile()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing username for GitHub user: '%v'", git.username)
	}
	git.username = ""
	return nil
}

func (git *GitHubAuthConfig) IsLoggedIn() bool {
	return git.isLoggedIn()
}

func (git *GitHubAuthConfig) GetCurrentUser() string {
	return git.username
}

// GetAuthToken retrieves git auth token of [username]
// Returns empty string if no user is logged in
// Returns err if err occurred getting auth token or no auth token found
func (git *GitHubAuthConfig) GetAuthToken() (string, error) {
	if git.isLoggedIn() {
		return "", nil
	}
	return getAuthToken(git.username)
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func (git *GitHubAuthConfig) isLoggedIn() bool {
	return git.username != ""
}

func doesGitHubUsernameFileExist() (bool, error) {
	filepath, err := host_machine_directories.GetGitHubUsernameFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the GitHub username filepath")
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred verifying if filepath '%v' exists", filepath)
	}
	return true, nil
}

func getGitHubUsernameFromFile() (string, error) {
	filepath, err := host_machine_directories.GetGitHubUsernameFilePath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the GitHub username filepath")
	}
	logrus.Debugf("Github username filepath: '%v'", filepath)

	fileContentBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading GitHub username file")
	}

	fileContentStr := string(fileContentBytes)

	return fileContentStr, nil
}

func saveGitHubUsernameFile(username string) error {
	fileContent := []byte(username)

	logrus.Debugf("Saving git username in file...")

	filepath, err := host_machine_directories.GetGitHubUsernameFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the GitHub username filepath")
	}

	err = os.WriteFile(filepath, fileContent, githubUsernameFilePermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing GitHub username to file '%v'", filepath)
	}
	logrus.Debugf("Saved GitHub username file")
	return nil
}

func removeGitHubUsernameFile() error {
	logrus.Debugf("Removing git username in file...")

	filepath, err := host_machine_directories.GetGitHubUsernameFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the GitHub username filepath")
	}

	err = os.Remove(filepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub username file '%v'", filepath)
	}
	logrus.Debugf("Removed Github username file")
	return nil
}

// getAuthToken attempts to retrieve auth token exists from keyring
// If not found or err occurs, attempts to retrieve auth token from plain text file
func getAuthToken(username string) (string, error) {
	var authToken string
	authToken, err := getAuthTokenFromKeyring(username)
	if err == nil {
		return authToken, nil
	}
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return "", stacktrace.Propagate(err, "An error getting auth token from keyring for GitHub user: %v.", username)
	}
	logrus.Debugf("No auth token found in keyring for user '%v'\nFalling back to retrieving auth token from plain text file.", username)
	githubAuthTokenFileExists, err := doesGitHubAuthTokenFileExist()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred verifying if GitHub auth token file exists for GitHub user: %v.", username)
	}
	if !githubAuthTokenFileExists {
		return "", stacktrace.NewError("No GitHub auth token found in keyring OR in plain text file for user '%v'.", username)
	}
	authToken, err = getGitHubAuthTokenFromFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting auth token from file for GitHub user: %v", username)
	}
	return authToken, nil
}

// setAuthToken attempts to set the git auth token for username
// Will attempt to store in secure system credential storage, but if no secure storage is found will resort to storing in a plain text file
func setAuthToken(username, authToken string) error {
	err := setAuthTokenInKeyring(username, authToken)
	if err == nil {
		return nil
	}
	logrus.Debugf("An error occurred setting GitHub auth token in keyring: %v\nFalling back to setting token in plain text file.", err)
	err = saveGitHubAuthTokenFile(authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred attempting to store GitHub auth token in plain text file after failing to store in keying.")
	}
	return nil
}

func removeAuthToken(username string) error {
	err := removeAuthTokenFromKeyring(username)
	if err == nil {
		return nil
	}
	logrus.Debugf("An error occurred removing GitHub auth token in keyring: %v\nAssuming token is in plain text file and removing from there.", err)
	err = removeGitHubAuthTokenFile()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub auth token from plain text file after failing to remove from keyring.")
	}
	return nil
}

func getAuthTokenFromKeyring(username string) (string, error) {
	authToken, err := keyring.Get(kurtosisCliKeyringServiceName, username)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting auth token for GitHub user '%v' from keyring", username)
	}
	return authToken, nil
}

func setAuthTokenInKeyring(username, authToken string) error {
	err := keyring.Set(kurtosisCliKeyringServiceName, username, authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting GitHub auth token for user '%v' in keyring.", username)
	}
	return nil
}

func removeAuthTokenFromKeyring(username string) error {
	err := keyring.Delete(kurtosisCliKeyringServiceName, username)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub auth token for user '%v' from keyring", username)
	}
	return nil
}

func doesGitHubAuthTokenFileExist() (bool, error) {
	filepath, err := host_machine_directories.GetGitHubAuthTokenFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the GitHub token filepath")
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred verifying if filepath '%v' exists", filepath)
	}
	return true, nil
}

func getGitHubAuthTokenFromFile() (string, error) {
	filepath, err := host_machine_directories.GetGitHubAuthTokenFilePath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the GitHub auth token filepath")
	}
	logrus.Debugf("Github username filepath: '%v'", filepath)

	fileContentBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading GitHub auth token file")
	}

	fileContentStr := string(fileContentBytes)

	return fileContentStr, nil
}

func saveGitHubAuthTokenFile(authToken string) error {
	fileContent := []byte(authToken)

	logrus.Debugf("Saving GitHub auth token in file...")

	filepath, err := host_machine_directories.GetGitHubAuthTokenFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the GitHub auth token filepath")
	}

	err = os.WriteFile(filepath, fileContent, githubAuthTokenFilePermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing GitHub auth token to file '%v'", filepath)
	}
	logrus.Debugf("Saved GitHub auth token")
	return nil
}

func removeGitHubAuthTokenFile() error {
	logrus.Debugf("Removing GitHub auth token file...")

	filepath, err := host_machine_directories.GetGitHubAuthTokenFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the GitHub auth token filepath")
	}

	err = os.Remove(filepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub auth token file '%v'", filepath)
	}
	logrus.Debugf("Removed GitHub auth token file")
	return nil
}
