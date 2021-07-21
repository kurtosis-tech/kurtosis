/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package enclave_data_volume

import (
	"bufio"
	"crypto"
	"fmt"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"sync"

	// This is a special type of import that includes the correct hashing algorithm that we use
	// If we don't have the "_" in front, Goland will complain it's unused
	_ "golang.org/x/crypto/sha3"
)

const (
	// Hash function for generating artifact filenames from artifact URLs
	artifactUrlHashFunction = crypto.SHA3_256
)

/*
An interface for interacting with the artifact cache directory that exists inside the suite execution volume,
	and is a) populated by the initializer and b) accessed by the files artifact expander
 */
type ArtifactCache struct {
	absoluteDirpath string
	dirpathRelativeToVolRoot string

	// Mutex to ensure we don't get race conditions when adding/removing from the cache
	mutex *sync.Mutex
}

func newArtifactCache(absoluteDirpath string, dirpathRelativeToVolRoot string) *ArtifactCache {
	return &ArtifactCache{
		absoluteDirpath: absoluteDirpath,
		dirpathRelativeToVolRoot: dirpathRelativeToVolRoot,
		mutex: &sync.Mutex{},
	}
}

func (cache ArtifactCache) AddArtifact(artifactId string, url string) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	newArtifactFile := cache.getArtifactFileObjFromArtifactId(artifactId)
	destAbsFilepath := newArtifactFile.absoluteFilepath
	if _, err := os.Stat(destAbsFilepath); err == nil {
		return stacktrace.NewError("Cannot add artifact '%v'; an artifact file with that key already exists", artifactId)
	}

	logrus.Debugf(
		"Downloading artifact with ID '%v' from URL '%v' to filepath '%v'",
		artifactId,
		url,
		destAbsFilepath,
	)
	if err := downloadArtifactToFilepath(url, destAbsFilepath); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred downloading artifact '%v' from '%v' to file '%v'",
			artifactId,
			url,
			destAbsFilepath,
		)
	}
	return nil
}

// Gets the artifact with the given URL, or throws an error if it doesn't exist
func (cache ArtifactCache) GetArtifact(artifactId string) (*EnclaveDataVolFile, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	artifactFile := cache.getArtifactFileObjFromArtifactId(artifactId)
	if _, err := os.Stat(artifactFile.absoluteFilepath); os.IsNotExist(err) {
		return nil, stacktrace.Propagate(err, "No artifact with ID '%v' exists in the cache", artifactId)
	}

	return artifactFile, nil
}

func (cache *ArtifactCache) getArtifactFileObjFromArtifactId(artifactId string) *EnclaveDataVolFile {
	absoluteFilepath := path.Join(cache.absoluteDirpath, artifactId)
	relativeFilepath := path.Join(cache.dirpathRelativeToVolRoot, artifactId)
	return newEnclaveDataVolFile(absoluteFilepath, relativeFilepath)
}

func downloadArtifactToFilepath(url string, destAbsFilepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to '%v' to get the artifact", url)
	}
	body := resp.Body
	defer body.Close()

	functionCompletedSuccessfully := false
	fp, err := os.Create(destAbsFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening the file to write the artifact to")
	}
	defer func() {
		// We need to delete the file we created if the function doesn't complete successfully
		if !functionCompletedSuccessfully {
			if err := os.Remove(destAbsFilepath); err != nil {
				logrus.Error(
					"Downloading artifact from URL '%v' to filepath '%v' encountered an error so we tried to remove " +
						"the file we created, but encountered an error removing it:",
					url,
					destAbsFilepath,
				)
				fmt.Fprintln(logrus.StandardLogger().Out, err)
				logrus.Error("WARNING: This will mean that you have a corrupted ghost artifact!!")
			}
		}
	}()
	defer fp.Close()

	bufferedWriter := bufio.NewWriter(fp)
	if _, err := io.Copy(bufferedWriter, body); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying the file bytes from the response to the filesystem")
	}

	functionCompletedSuccessfully = true
	return nil
}

