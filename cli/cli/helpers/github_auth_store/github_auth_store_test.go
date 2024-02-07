package github_auth_store

import (
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
	"os"
	"testing"
)

const (
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	//DO NOT CHANGE THIS VALUE
	//!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!! WARNING !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	expectedKeyringServiceName = "kurtosis-cli"

	tempFileDir                  = ""
	tempUsernameFileNamePattern  = "github-username"
	tempAuthTokenFileNamePattern = "github-token"
)

// The keyring service name in this package has to be always "kurtosis-cli"
// so we control that it does not change
func TestKeyringServiceNameDoesNotChange(t *testing.T) {
	require.Equal(t, expectedKeyringServiceName, kurtosisCliKeyringServiceName)
}

func TestGetUserReturnsEmptyStringIfNoUserExists(t *testing.T) {
	// setup mock GitHub store
	tempUsernameFile, err := os.CreateTemp(tempFileDir, tempUsernameFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempUsernameFile.Name())
	tempAuthTokenFile, err := os.CreateTemp(tempFileDir, tempAuthTokenFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempAuthTokenFile.Name())

	// run test
	store := newGitHubAuthStoreForTesting(tempUsernameFile.Name(), tempAuthTokenFile.Name())

	actualUsername, err := store.GetUser()
	require.NoError(t, err)
	require.Empty(t, actualUsername)
}

func TestGetUserReturnsUser(t *testing.T) {
	// setup mock GitHub store
	tempUsernameFile, err := os.CreateTemp(tempFileDir, tempUsernameFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempUsernameFile.Name())
	tempAuthTokenFile, err := os.CreateTemp(tempFileDir, tempAuthTokenFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempAuthTokenFile.Name())

	expectedUsername := "john123"
	_, err = tempUsernameFile.Write([]byte(expectedUsername))
	require.NoError(t, err)

	// run test
	store := newGitHubAuthStoreForTesting(tempUsernameFile.Name(), tempAuthTokenFile.Name())

	actualUsername, err := store.GetUser()
	require.NoError(t, err)
	require.Equal(t, expectedUsername, actualUsername)
}

func TestGetAuthTokenGetsTokenFromKeyring(t *testing.T) {

}

func TestGetAuthTokenReturnsEmptyStringIfNoUserExists(t *testing.T) {

}

func TestGetAuthTokenGetsTokenFromFile(t *testing.T) {

}

func TestGetAuthTokenReturnsNoTokenFoundIfUserExistsWithNoToken(t *testing.T) {

}

func TestSetUser(t *testing.T) {
	// setup mock GitHub store
	tempUsernameFile, err := os.CreateTemp(tempFileDir, tempUsernameFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempUsernameFile.Name())
	tempAuthTokenFile, err := os.CreateTemp(tempFileDir, tempAuthTokenFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempAuthTokenFile.Name())

	expectedUsername := "john123"
	expectedAuthToken := "password"

	// run test
	store := newGitHubAuthStoreForTesting(tempUsernameFile.Name(), tempAuthTokenFile.Name())

	currentUser, err := store.GetUser()
	require.NoError(t, err)
	require.Empty(t, currentUser)

	err = store.SetUser(expectedUsername, expectedAuthToken)
	require.NoError(t, err)

	actualUsername, err := store.GetUser()
	require.NoError(t, err)
	require.Equal(t, expectedUsername, actualUsername)

	actualAuthToken, err := store.GetAuthToken()
	require.NoError(t, err)
	require.Equal(t, expectedAuthToken, actualAuthToken)
}

func TestSetUserOverwritesExistingUser(t *testing.T) {
	// setup mock GitHub store
	tempUsernameFile, err := os.CreateTemp(tempFileDir, tempUsernameFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempUsernameFile.Name())
	tempAuthTokenFile, err := os.CreateTemp(tempFileDir, tempAuthTokenFileNamePattern)
	require.NoError(t, err)
	defer os.Remove(tempAuthTokenFile.Name())

	oldUser := "john123"
	oldToken := "password"
	_, err = tempUsernameFile.Write([]byte(oldUser))
	require.NoError(t, err)
	err = keyring.Set(kurtosisCliKeyringServiceName, oldUser, oldToken)
	require.NoError(t, err)

	// run test
	store := newGitHubAuthStoreForTesting(tempUsernameFile.Name(), tempAuthTokenFile.Name())

	currentUser, err := store.GetUser()
	require.NoError(t, err)
	require.Empty(t, currentUser)

	newUser := "tim"
	newToken := "wordpass"
	err = store.SetUser(newUser, newToken)
	require.NoError(t, err)

	actualNewUser, err := store.GetUser()
	require.NoError(t, err)
	require.Equal(t, newUser, actualNewUser)

	actualNewToken, err := store.GetAuthToken()
	require.NoError(t, err)
	require.Equal(t, newToken, actualNewToken)
}

func TestRemoveUserIsNoOpIfNoUserExists(t *testing.T) {

}

func TestRemoveUser(t *testing.T) {

}
