package github_auth_config

import (
	"github.com/cli/go-gh/v2/pkg/browser"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
	"os"
)

const (
	githubUsernameFilePermission  os.FileMode = 0644
	githubAuthTokenFilePermission os.FileMode = 0644

	kurtosisCliKeyringServiceName = "kurtosis-cli"
)

type GithubAuthConfig struct {
	// Empty string if no user is logged in
	currentUsername string
}

func GetGithubAuthConfig() (*GithubAuthConfig, error) {
	// Existence of github user filepath determines whether a user is logged in or not
	var curentUsername string
	isUserLoggedIn, err := doesGithubUserFilepathExist()
	if err != nil {
		return nil, err
	}
	if isUserLoggedIn {

	}
	return &GithubAuthConfig{
		currentUsername: curentUsername,
	}, nil
}

// what happens if you login two accounts?
func (git *GithubAuthConfig) Login() (string, error) {

	secret, userLogin, err := AuthFlow("github.com", "", []string{}, true, *browser.New("", os.Stdout, os.Stderr))
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred in the user login flow.")
	}
	logrus.Infof("Successfully authorized git user: %v", userLogin)

	// set password
	err = keyring.Set("kurtosis-git", "tedim52", secret)
	if err != nil {
		logrus.Errorf("Unable to set token for keyring")
	}
	err = os.Setenv("GIT_USER", userLogin)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred setting git user env var.")
	}
	logrus.Debugf("Successfully set git token in keyring: %v", secret)
	logrus.Infof("Successfully set git auth info for user: %v", "tedim52")
	return "", nil
}

// Logout "logs out" a GitHub user from Kurtosis by:
// - removing the auth token associated with [currentUsername] from keyring or plain text file
// -removing the GitHub username file
func (git *GithubAuthConfig) Logout() error {

	return nil
}

func (git *GithubAuthConfig) GetCurrentUser() string {
	return ""
}

// GetAuthToken retrieves git auth token of currentUser
// First, checks if auth token exists in keyring
// If not found in keyring, attempts to retrieve auth token from plain text file
// Returns empty string if no auth token found
func (git *GithubAuthConfig) GetAuthToken() string {
	return ""
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func doesGithubUserFilepathExist() (bool, error) {
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
