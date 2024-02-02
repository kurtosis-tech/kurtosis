package github_auth_config

import (
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
	"os"
	"sync"
)

const (
	// Check what file permissions are needed
	githubUsernameFilePermission  os.FileMode = 0644
	githubAuthTokenFilePermission os.FileMode = 0644

	kurtosisCliKeyringServiceName = "kurtosis-cli"
)

type GithubAuthConfig struct {
	// GithubAuthConfig is protected by a mutex
	// do we need to protect ?
	mutex *sync.RWMutex

	// Empty string if no user is currently logged in
	currentUsername string
}

func GetGithubAuthConfig() (*GithubAuthConfig, error) {
	// Existence of GitHub user filepath determines whether a user is logged in or not
	var currentUsername string
	isUserLoggedIn, err := doesGithubUsernameFilepathExist()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred verifying existence of a GitHub user.")
	}
	if isUserLoggedIn {
		username, err := getGithubUsernameFromFile()
		if err != nil {
			return nil, stacktrace.Propagate(err, "GitHub user found but error occurred getting username.")
		}
		currentUsername = username
	}
	return &GithubAuthConfig{
		currentUsername: currentUsername,
	}, nil
}

func (git *GithubAuthConfig) Login() error {
	git.mutex.Lock()
	defer git.mutex.Unlock()

	// Don't log in if user already exists
	if git.currentUsername != "" {
		return nil
	}

	authToken, userLogin, err := AuthFlow("github.com", "", []string{}, true, *browser.New("", os.Stdout, os.Stderr))
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred in the Github OAuth flow.")
	}

	err = saveGithubUsernameFile(userLogin)
	shouldRemoveUsernameFile := true
	defer func() {
		if shouldRemoveUsernameFile {
			err := removeGithubUsernameFile()
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
			err = unsetAuthToken(git.currentUsername)
			if err != nil {
				logrus.Errorf("Failed to unset git auth token after setting the auth token failed!! GitHub auth could be in a bad state.")
			}
		}
	}()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting auth token for user: %v.", userLogin)
	}

	git.currentUsername = userLogin

	shouldRemoveAuthToken = false
	shouldRemoveUsernameFile = false
	return nil
}

// Logout "logs out" a GitHub user from Kurtosis by:
// -removing the auth token associated with [currentUsername] from keyring or plain text file
// -removing the GitHub username file
// -setting [currentUsername] to empty string
func (git *GithubAuthConfig) Logout() error {
	// TODO
	return nil
}

func (git *GithubAuthConfig) IsLoggedIn() bool {
	return git.currentUsername != ""
}

func (git *GithubAuthConfig) GetCurrentUser() string {
	return git.currentUsername
}

// GetAuthToken retrieves git auth token of [currentUsername]
// First, checks if auth token exists in keyring
// If not found in keyring, attempts to retrieve auth token from plain text file
// Returns empty string if no user is logged in
// Returns err if user is logged in but no auth token is associated with them
func (git *GithubAuthConfig) GetAuthToken() (string, error) {
	// TODO
	return "", nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func doesGithubUsernameFilepathExist() (bool, error) {
	filepath, err := host_machine_directories.GetGithubUserFilePath()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath")
	}

	_, err = os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred checking if filepath '%v' exists", filepath)
	}
	return true, nil
}

func getGithubUsernameFromFile() (string, error) {
	filepath, err := host_machine_directories.GetGithubUserFilePath()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting the github username filepath")
	}
	logrus.Debugf("Github username filepath: '%v'", filepath)

	fileContentBytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading github username file")
	}

	fileContentStr := string(fileContentBytes)

	return fileContentStr, nil
}

func saveGithubUsernameFile(username string) error {
	fileContent := []byte(username)

	logrus.Debugf("Saving git username in file...")

	filepath, err := host_machine_directories.GetGithubUserFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath")
	}

	err = os.WriteFile(filepath, fileContent, githubUsernameFilePermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing git username to file '%v'", filepath)
	}
	logrus.Debugf("Metrics user id file saved")
	return nil
}

func removeGithubUsernameFile() error {
	logrus.Debugf("Removing git username in file...")

	filepath, err := host_machine_directories.GetGithubUserFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the metrics user id filepath")
	}

	err = os.Remove(filepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing github username file '%v'", filepath)
	}
	logrus.Debugf("Github username file removed")
	return nil
}

// SetAuthToken attempts to set the git auth token for username
// Will attempt to store in secure system credential storage, but if it is not found will resort to storing in a plain text file
func setAuthToken(username, authToken string) error {
	// TODO
	return nil
}

// SetAuthToken attempts to set the git auth token for username
// Will attempt to store in secure system credential storage, but if it is not found will resort to storing in a plain text file
func unsetAuthToken(username string) error {
	// TODO
	return nil
}

func doesGitAuthTokenFilepathExist() (bool, error) {
	return false, nil
}

func saveGitAuthTokenFile(authToken string) error {
	fileContent := []byte(authToken)

	logrus.Debugf("Saving git auth token id in file...")

	filepath, err := host_machine_directories.GitGithubAuthTokenFilePath()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting the git auth token filepath")
	}

	err = os.WriteFile(filepath, fileContent, githubAuthTokenFilePermission)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing git auth token to file '%v'", filepath)
	}
	logrus.Debugf("Metrics user id file saved")
	return nil
}

// TODO: implement get auth token from file
// TODO: implement remove auth token file

func getAuthTokenFromKeyring(username string) (string, error) {
	authToken, err := keyring.Get(kurtosisCliKeyringServiceName, username)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting auth token for user '%v', from keyring", username)
	}
	return authToken, nil
}

func setAuthTokenForInKeyring(username, authToken string) error {
	err := keyring.Set(kurtosisCliKeyringServiceName, username, authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting auth token for user '%v' in keyring.", username)
	}
	return nil
}

func removeAuthTokenFromKeyring(username string) error {
	err := keyring.Delete(kurtosisCliKeyringServiceName, username)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing auth token for user '%v' from keyring", username)
	}
	return nil
}
