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
	kurtosisStorageDirectory = ".kurtosis"
	kurtosisTokenStorageFileName = "access_token"
)

// Package-wide lock for reading/writing files.
var lock sync.Mutex

// marshal is a function that marshals the object into an
// io.Reader.
// By default, it uses the JSON marshaller.
// https://medium.com/@matryer/golang-advent-calendar-day-eleven-persisting-go-objects-to-disk-7caf1ee3d11d
func marshalObject(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

// Unmarshal is a function that unmarshals the data from the
// reader into the specified value.
// By default, it uses the JSON unmarshaller.
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
// Use os.IsNotExist() to see if the returned error is due
// to the file being missing.
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
