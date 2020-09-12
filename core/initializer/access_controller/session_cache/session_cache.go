/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package session_cache

import (
	"bytes"
	"encoding/json"
	"github.com/kurtosis-tech/kurtosis/initializer/access_controller/auth0"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

const (
	// By default, the kurtosis storage directory will be created in the user's home directory.
	kurtosisStorageDirectory     = ".kurtosis"
	kurtosisTokenStorageFileName = "access_token"
)

// Package-wide lock for reading/writing files.
var lock sync.Mutex

type SessionCache struct {
	StorageDirectoryFullPath string
	AccessTokenFileFullPath string
}

func NewSessionCache() (*SessionCache, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to find user home directory.")
	}
	storageDirectoryFullPath := userHomeDir + "/" + kurtosisStorageDirectory
	err = createDirectoryIfNotExist(storageDirectoryFullPath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to create-if-not-exists session cache directory %s", storageDirectoryFullPath)
	}
	accessTokenFileFullPath := storageDirectoryFullPath + "/" + kurtosisTokenStorageFileName
	return &SessionCache{storageDirectoryFullPath, accessTokenFileFullPath}, nil
}

/*
	Writes a tokenResponse to a local file in the user's home directory.
	On later runs of Kurtosis, the token will be preserved and re-auth will be unnecessary.
 */
func (cache *SessionCache) PersistToken(tokenResponse *auth0.TokenResponse) error {
	if err := saveObject(cache.AccessTokenFileFullPath, tokenResponse); err != nil {
		return stacktrace.Propagate(err, "Failed to cache users access token after authenticating.")
	}
	return nil
}

/*
	Loads a tokenResponse from a local file in the user's home directory.
	Returns a boolean alreadyAuthenticated to indicate if a tokenResponse had been written before.
	TODO TODO TODO Ensure that the token has been verified against the provider in the last 48 hours
*/
func (cache *SessionCache) LoadToken() (tokenResponse *auth0.TokenResponse, alreadyAuthenticated bool, err error){
	tokenResponse = new(auth0.TokenResponse)
	if _, err := os.Stat(cache.AccessTokenFileFullPath); err == nil {
		if err := loadObject(cache.AccessTokenFileFullPath, &tokenResponse); err != nil {
			return nil, false, stacktrace.Propagate(err, "Failed to load users access token.")
		}
		return tokenResponse, true, nil
	} else if os.IsNotExist(err) {
		return nil, false, nil
	} else {
		return nil, false, stacktrace.Propagate(err, "Received error checking for status of token file %s", cache.AccessTokenFileFullPath)
	}
}

// ================================= HELPER FUNCTIONS =========================================

/*
   marshals the object into an io.Reader. By default, it uses the JSON marshaller.
   Read: https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
*/
func marshalObject(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, stacktrace.Propagate(err, "Failed to marshal object")
	}
	return bytes.NewReader(b), nil
}

/*
	unmarshals data from an io Reader into the specified value. By default, it uses the JSON unmarshaller.
	Read: https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
*/
func unmarshalObject(r io.Reader, v interface{}) error {
	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to unmarshal object.")
	}
	return nil
}

// Save saves a representation of v to the file at path.
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
func saveObject(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to create %s", path)
	}
	defer f.Close()
	r, err := marshalObject(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// loads the file at path into v.
func loadObject(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return stacktrace.Propagate(err, "Failed to open %s", path)
	}
	defer f.Close()
	return unmarshalObject(f, v)
}

// checks if the directory specified in path exists. if not, creates it.
func createDirectoryIfNotExist(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Creating kurtosis storage directory at %s", path)
			os.Mkdir(path, 0777)
		} else {
			return stacktrace.Propagate(err, "Failed to check stat for %s", path)
		}
	}
	return nil
}
