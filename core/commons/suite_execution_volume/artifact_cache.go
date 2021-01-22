/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package suite_execution_volume

import (
	"bufio"
	"crypto"
	"encoding/hex"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"

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
}

func newArtifactCache(absoluteDirpath string, dirpathRelativeToVolRoot string) *ArtifactCache {
	return &ArtifactCache{absoluteDirpath: absoluteDirpath, dirpathRelativeToVolRoot: dirpathRelativeToVolRoot}
}

func (cache ArtifactCache) AddArtifact(artifactUrl string) error {
	logrus.Debug("Downloading artifacts from URL: %v", artifactUrl)

	artifactUrlHash, err := hashArtifactUrl(artifactUrl)
	if err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred hashing artifact URL '%v' to get a filename for the artifact",
			artifactUrl)
	}

	absoluteFilepath := path.Join(cache.absoluteDirpath, artifactUrlHash)
	if _, err := os.Stat(absoluteFilepath); err == nil {
		// NOTE: we could just make this a no-op; we chose to throw an error here to make sure the user
		//  isn't double-adding things
		return stacktrace.NewError("Cannot download artifact with URL '%v'; artifact already exists")
	}

	if err := downloadArtifactToFilepath(artifactUrl, absoluteFilepath); err != nil {
		return stacktrace.Propagate(
			err,
			"An error occurred downloading artifact from '%v' to file '%v'",
			artifactUrl,
			absoluteFilepath)
	}
	return nil
}

// Gets the artifact with the given URL, or throws an error if it doesn't exist
func (cache ArtifactCache) GetArtifact(artifactUrl string) (*Artifact, error) {
	artifactUrlHash, err := hashArtifactUrl(artifactUrl)
	if err != nil {
		return nil, stacktrace.Propagate(
			err,
			"An error occurred hashing artifact URL '%v' to get a filename for the artifact",
			artifactUrl)
	}
	absoluteFilepath := path.Join(cache.absoluteDirpath, artifactUrlHash)
	relativeFilepath := path.Join(cache.dirpathRelativeToVolRoot, artifactUrlHash)
	file := newFile(absoluteFilepath, relativeFilepath)
	return newArtifact(artifactUrlHash, file), nil
}

func hashArtifactUrl(artifactUrl string) (string, error) {
	hasher := artifactUrlHashFunction.New()
	artifactUrlBytes := []byte(artifactUrl)
	if _, err := hasher.Write(artifactUrlBytes); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred writing the artifact URL to the hash function")
	}
	hexEncodedHash := hex.EncodeToString(hasher.Sum(nil))
	return hexEncodedHash, nil
}

func downloadArtifactToFilepath(url string, destFilepath string) error {
	fp, err := os.Create(destFilepath)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred opening the file to write the artifact to")
	}
	defer fp.Close()

	bufferedWriter := bufio.NewWriter(fp)

	resp, err := http.Get(url)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred making the request to '%v' to get the artifact", url)
	}
	body := resp.Body
	defer body.Close()

	if _, err := io.Copy(bufferedWriter, body); err != nil {
		return stacktrace.Propagate(err, "An error occurred copying the file bytes from the response to the filesystem")
	}

	return nil
}

