package github_auth_store

import (
	"errors"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/host_machine_directories"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"github.com/zalando/go-keyring"
	"os"
	"sync"
)

const (
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// DO NOT CHANGE THIS VALUE
	// Changing this value could leak tokens in a users keyring/make Kurtosis unable to retrieve/remove them.
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	kurtosisCliKeyringServiceName = "kurtosis-cli"

	githubAuthFilesPerms = 0644
)

var (
	// NOTE: This will be initialized exactly once (singleton pattern)
	githubAuthStore GitHubAuthStore
	once            sync.Once

	ErrNoTokenFound = errors.New("no token found for currently logged in user")
)

// GitHubAuthStore stores information about a GitHub user that has authorized Kurtosis CLI to perform git operations on their behalf
// [username] is their GitHub username
// [authToken] is a scoped token that authorizes Kurtosis CLI on behalf of [username
type GitHubAuthStore interface {
	// GetUser returns [username] of current user
	// If no user exists, returns empty string
	GetUser() (string, error)

	// GetAuthToken returns authToken for the user if they exist
	// If [authToken] doesn't exist in system credential storage, attempts to retrieve token from plain text file
	// Returns empty string if no user exists
	// Returns ErrNoTokenFound err if user exists but no [authToken] was found
	GetAuthToken() (string, error)

	// SetUser sets current user to [username] and stores their [authToken] in system credential storage if it exists
	// otherwise, stores [authToken] in plain text file
	SetUser(username, authToken string) error

	// RemoveUser removes user and user's [authToken] from store, if a user exists
	RemoveUser() error
}

func GetGitHubAuthStore() (GitHubAuthStore, error) {
	store, err := NewGitHubAuthStore()
	if err != nil {
		return nil, err
	}
	once.Do(func() {
		// NOTE: We use a 'once' to initialize the GitHubAuthStore because it contains a mutex to guard
		// the files, and we don't ever want multiple GitHubAuthStore instances in existence
		githubAuthStore = store
	})
	return githubAuthStore, nil
}

type githubConfigStoreImpl struct {
	*sync.RWMutex

	usernameFilePath, authTokenFilePath string
}

func NewGitHubAuthStore() (GitHubAuthStore, error) {
	usernameFilePath, err := host_machine_directories.GetGitHubUsernameFilePath()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the GitHub username filepath")
	}
	authTokenFilePath, err := host_machine_directories.GetGitHubAuthTokenFilePath()
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred getting the Github auth token filepath")
	}
	return &githubConfigStoreImpl{
		RWMutex:           &sync.RWMutex{},
		usernameFilePath:  usernameFilePath,
		authTokenFilePath: authTokenFilePath,
	}, nil
}

func newGitHubAuthStoreForTesting(testUsernameFilePath, testAuthTokenFilePath string) GitHubAuthStore {
	return &githubConfigStoreImpl{
		RWMutex:           &sync.RWMutex{},
		usernameFilePath:  testUsernameFilePath,
		authTokenFilePath: testAuthTokenFilePath,
	}
}

func (store *githubConfigStoreImpl) GetUser() (string, error) {
	store.RLock()
	defer store.RUnlock()

	username, userExists, err := store.getUserAndIfTheyExist()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred verifying if a user exists.")
	}
	if !userExists {
		return "", nil
	}
	return username, nil
}

func (store *githubConfigStoreImpl) GetAuthToken() (string, error) {
	store.RLock()
	defer store.RUnlock()

	username, userExists, err := store.getUserAndIfTheyExist()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred verifying if a user exists.")
	}
	if !userExists {
		return "", nil
	}

	authToken, err := getAuthTokenFromKeyring(username)
	if err == nil {
		return authToken, nil
	}
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return "", stacktrace.Propagate(err, "An error getting auth token from keyring for GitHub user: %v.", username)
	}
	logrus.Debugf("No auth token found in keyring for user '%v'\nFalling back to retrieving auth token from plain text file.", username)
	githubAuthTokenFileExists, err := store.doesGitHubAuthTokenFileExist()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred verifying if GitHub auth token file exists for GitHub user: %v.", username)
	}
	if !githubAuthTokenFileExists {
		return "", ErrNoTokenFound
	}
	authToken, err = store.getGitHubAuthTokenFromFile()
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred getting auth token from file for GitHub user: %v", username)
	}
	if authToken == "" {
		return "", ErrNoTokenFound
	}
	return authToken, nil
}

func (store *githubConfigStoreImpl) SetUser(username, authToken string) error {
	store.Lock()
	defer store.Unlock()

	err := store.saveGitHubUsernameFile(username)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred saving '%v' to store.", username)
	}
	shouldUnsetGitHubUsername := true
	defer func() {
		if shouldUnsetGitHubUsername {
			if err := store.removeGitHubUsernameFile(); err != nil {
				logrus.Errorf("Error occurred removing GitHub username after setting it failed!!! GitHub auth could be in a bad state.")
			}
		}
	}()

	err = store.setAuthToken(username, authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred setting auth token in store for user: %v", username)
	}
	shouldUnsetAuthToken := true
	defer func() {
		if shouldUnsetAuthToken {
			if err = store.removeAuthToken(username); err != nil {
				logrus.Errorf("Error occurred removing GitHub auth token after setting it failed!!! GitHub auth could be in a bad state.")
			}
		}
	}()

	shouldUnsetGitHubUsername = false
	shouldUnsetAuthToken = false
	return nil
}

func (store *githubConfigStoreImpl) RemoveUser() error {
	store.Lock()
	defer store.Unlock()

	username, userExists, err := store.getUserAndIfTheyExist()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred verifying if a user exists.")
	}
	if !userExists {
		return nil
	}

	err = store.removeAuthToken(username)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing user '%v' token from store.", username)
	}

	err = store.removeGitHubUsernameFile()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub user '%v' from store", username)
	}

	return nil
}

func (store *githubConfigStoreImpl) getUserAndIfTheyExist() (string, bool, error) {
	userExists, err := store.doesGitHubUsernameFileExist()
	if err != nil {
		return "", false, stacktrace.Propagate(err, "An error occurred discovering if user exists.")
	}
	if !userExists {
		return "", false, nil
	}
	username, err := store.getGitHubUsernameFromFile()
	if err != nil {
		return "", false, stacktrace.Propagate(err, "An error occurred getting user from store.")
	}
	if username == "" {
		return "", false, nil
	}
	return username, true, nil
}

// setAuthToken attempts to set the git auth token for username
// Will attempt to store in secure system credential storage, but if no secure storage is found will resort to storing in a plain text file
func (store *githubConfigStoreImpl) setAuthToken(username, authToken string) error {
	err := setAuthTokenInKeyring(username, authToken)
	if err == nil {
		return nil
	}
	logrus.Debugf("An error occurred setting GitHub auth token in keyring: %v\nFalling back to setting token in plain text file.", err)
	err = store.saveGitHubAuthTokenFile(authToken)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred attempting to store GitHub auth token in plain text file after failing to store in keyring.")
	}
	return nil
}

func (store *githubConfigStoreImpl) removeAuthToken(username string) error {
	err := removeAuthTokenFromKeyring(username)
	if err == nil {
		return nil
	}
	logrus.Debugf("An error occurred removing GitHub auth token in keyring: %v\nAssuming token is in plain text file and removing from there.", err)
	err = store.removeGitHubAuthTokenFile()
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub auth token from plain text file after failing to remove from keyring.")
	}
	return nil
}

func (store *githubConfigStoreImpl) doesGitHubUsernameFileExist() (bool, error) {
	_, err := os.Stat(store.usernameFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred verifying if filepath '%v' exists", store.usernameFilePath)
	}
	return true, nil
}

func (store *githubConfigStoreImpl) getGitHubUsernameFromFile() (string, error) {
	logrus.Debugf("Github username filepath: '%v'", store.usernameFilePath)
	fileContentBytes, err := os.ReadFile(store.usernameFilePath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading GitHub username file")
	}
	fileContentStr := string(fileContentBytes)
	return fileContentStr, nil
}

func (store *githubConfigStoreImpl) saveGitHubUsernameFile(username string) error {
	fileContent := []byte(username)
	logrus.Debugf("Saving git username in file...")
	err := os.WriteFile(store.usernameFilePath, fileContent, githubAuthFilesPerms)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing GitHub username to file '%v'", store.usernameFilePath)
	}
	logrus.Debugf("Saved GitHub username file")
	return nil
}

func (store *githubConfigStoreImpl) removeGitHubUsernameFile() error {
	logrus.Debugf("Removing git username in file...")
	err := os.Remove(store.usernameFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub username file '%v'", store.usernameFilePath)
	}
	logrus.Debugf("Removed Github username file")
	return nil
}

func (store *githubConfigStoreImpl) doesGitHubAuthTokenFileExist() (bool, error) {
	_, err := os.Stat(store.authTokenFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, stacktrace.Propagate(err, "An error occurred verifying if filepath '%v' exists", store.authTokenFilePath)
	}
	return true, nil
}

func (store *githubConfigStoreImpl) getGitHubAuthTokenFromFile() (string, error) {
	fileContentBytes, err := os.ReadFile(store.authTokenFilePath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred reading GitHub auth token file")
	}
	fileContentStr := string(fileContentBytes)
	return fileContentStr, nil
}

func (store *githubConfigStoreImpl) saveGitHubAuthTokenFile(authToken string) error {
	fileContent := []byte(authToken)
	logrus.Debugf("Saving GitHub auth token in file...")
	err := os.WriteFile(store.authTokenFilePath, fileContent, githubAuthFilesPerms)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred writing GitHub auth token to file '%v'", store.authTokenFilePath)
	}
	logrus.Debugf("Saved GitHub auth token")
	return nil
}

func (store *githubConfigStoreImpl) removeGitHubAuthTokenFile() error {
	logrus.Debugf("Removing GitHub auth token file...")
	err := os.Remove(store.authTokenFilePath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred removing GitHub auth token file '%v'", store.authTokenFilePath)
	}
	logrus.Debugf("Removed GitHub auth token file")
	return nil
}

func getAuthTokenFromKeyring(username string) (string, error) {
	authToken, err := keyring.Get(kurtosisCliKeyringServiceName, username)
	if err != nil && errors.Is(err, keyring.ErrNotFound) {
		return "", err // don't wrap so this specific err can be detected
	}
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred retrieving token for '%v' from keyring", username)
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
