/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package session_cache

import (
	"bytes"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0_device_authorizer"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	// By default, the kurtosis storage directory will be created in the user's home directory.
	storageDirName            = ".kurtosis"
	tokenFileName             = "access_token"
	userReadWriteExecutePerms = 0700
)


type SessionCache struct {
	StorageDirPath string
	TokenFilePath  string
	lock           sync.Mutex
}

func NewSessionCache() (*SessionCache, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to find user home directory.")
	}

	storageDirectoryFullPath := filepath.Join(userHomeDir, storageDirName)
	err = createDirectoryIfNotExist(storageDirectoryFullPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create-if-not-exists session cache directory %s", storageDirectoryFullPath)
	}

	accessTokenFileFullPath := filepath.Join(storageDirectoryFullPath, tokenFileName)

	var lock sync.Mutex
	return &SessionCache{storageDirectoryFullPath, accessTokenFileFullPath, lock}, nil
}

/*
	Writes a tokenResponse to a local file in the user's home directory.
	On later runs of Kurtosis, the token will be preserved and re-auth will be unnecessary.
 */
func (cache *SessionCache) PersistToken(tokenResponse *auth0_device_authorizer.TokenResponse) error {
	if err := cache.saveObject(cache.TokenFilePath, tokenResponse); err != nil {
		return stacktrace.Propagate(err, "Failed to cache users access token after authenticating.")
	}
	return nil
}

/*
	Loads a tokenResponse from a local file in the user's home directory.
	Returns a boolean alreadyAuthenticated to indicate if a tokenResponse had been written before.
	TODO TODO TODO Ensure that the token has been verified against the provider in the last 48 hours
*/
func (cache *SessionCache) LoadToken() (tokenResponse *auth0_device_authorizer.TokenResponse, alreadyAuthenticated bool, err error){
	tokenResponse = new(auth0_device_authorizer.TokenResponse)
	if _, err := os.Stat(cache.TokenFilePath); err == nil {
		if err := cache.loadObject(cache.TokenFilePath, &tokenResponse); err != nil {
			return nil, false, stacktrace.Propagate(err, "Failed to load users access token.")
		}
		return tokenResponse, true, nil
	} else if os.IsNotExist(err) {
		return nil, false, nil
	} else {
		return nil, false, stacktrace.Propagate(err, "Received error checking for status of token file %s", cache.TokenFilePath)
	}
}

// ================================= HELPER FUNCTIONS =========================================

// saves a representation of object to the file at path.
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
func (cache *SessionCache) saveObject(path string, object interface{}) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	filePointer, err := os.Create(path)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create %s", path)
	}
	defer filePointer.Close()

	jsonBytes, err := json.MarshalIndent(object, "", "\t")
	if err != nil {
		return stacktrace.Propagate(err, "Failed to marshal object")
	}

	jsonBytesReader := bytes.NewReader(jsonBytes)
	_, err = io.Copy(filePointer, jsonBytesReader)

	return err
}

// loads the file at path into v.
func (cache *SessionCache) loadObject(path string, object interface{}) error {
	cache.lock.Lock()
	defer cache.lock.Unlock()

	filePointer, err := os.Open(path)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to open %s", path)
	}
	defer filePointer.Close()

	err = json.NewDecoder(filePointer).Decode(object)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to unmarshal object.")
	}

	return nil
}

// checks if the directory specified in path exists. if not, creates it.
func createDirectoryIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Creating kurtosis storage directory at %s", path)
			err := os.Mkdir(path, userReadWriteExecutePerms)
			if err != nil {
				return stacktrace.Propagate(err, "Failed to create directory %s", path)
			}
		} else {
			return stacktrace.Propagate(err, "Failed to check stat for %s", path)
		}
	}
	return nil
}
