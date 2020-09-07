/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package access_controller

import (
	"bytes"
	"encoding/json"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"sync"
)

const (
	// By default, the kurtosis storage directory will be created in the user's home directory.
	kurtosisStorageDirectory = ".kurtosis"
	kurtosisTokenStorageFileName = "access_token"
)

// Package-wide lock for reading/writing files.
var lock sync.Mutex

/*
	Writes a tokenResponse to a local file in the user's home directory.
	On later runs of Kurtosis, the token will be preserved and re-auth will be unnecessary.
 */
func persistToken(tokenResponse *TokenResponse) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return stacktrace.Propagate(err, "Failed to find user home directory.")
	}
	logrus.Debugf("User home dir: %+v", userHomeDir)
	kurtosisStorageDirectoryFullPath := userHomeDir + "/" + kurtosisStorageDirectory
	_, err = os.Stat(kurtosisStorageDirectoryFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Creating kurtosis storage directory at %s", kurtosisStorageDirectoryFullPath)
			os.Mkdir(kurtosisStorageDirectoryFullPath, 0777)
		} else {
			return stacktrace.Propagate(err, "")
		}
	}
	if err := saveObject(kurtosisStorageDirectoryFullPath + "/" + kurtosisTokenStorageFileName, tokenResponse); err != nil {
		return stacktrace.Propagate(err, "Failed to cache users access token after authenticating.")
	}
	return nil
}

/*
	Loads a tokenResponse from a local file in the user's home directory.
	Returns a boolean alreadyAuthenticated to indicate if a tokenResponse had been written before.
*/
func loadToken() (tokenResponse *TokenResponse, alreadyAuthenticated bool, err error){
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, false, stacktrace.Propagate(err, "Failed to find user home directory.")
	}
	logrus.Debugf("User home dir: %+v", userHomeDir)
	kurtosisStorageDirectoryFullPath := userHomeDir + "/" + kurtosisStorageDirectory
	_, err = os.Stat(kurtosisStorageDirectoryFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("Creating kurtosis storage directory at %s", kurtosisStorageDirectoryFullPath)
			os.Mkdir(kurtosisStorageDirectoryFullPath, 0777)
			return nil, false, nil
		} else {
			return nil, false, stacktrace.Propagate(err, "")
		}
	}
	tokenResponse = new(TokenResponse)
	if err := loadObject(kurtosisStorageDirectoryFullPath + "/" + kurtosisTokenStorageFileName, &tokenResponse); err != nil {
		return nil, false, stacktrace.Propagate(err, "Failed to load users access token.")
	}
	return tokenResponse, true, nil
}

// ================================= HELPER FUNCTIONS =========================================

/*
   marshals the object into an io.Reader. By default, it uses the JSON marshaller.
   Read: https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
*/
func marshalObject(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

/*
	unmarshals data from an io Reader into the specified value. By default, it uses the JSON unmarshaller.
	Read: https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
*/
func unmarshalObject(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// Save saves a representation of v to the file at path.
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
func saveObject(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := marshalObject(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// Load loads the file at path into v.
func loadObject(path string, v interface{}) error {
	lock.Lock()
	defer lock.Unlock()
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return unmarshalObject(f, v)
}
