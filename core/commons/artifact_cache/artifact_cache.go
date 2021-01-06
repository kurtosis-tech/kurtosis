/*
 * Copyright (c) 2021 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package artifact_cache

import (
	"bufio"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
)

const (
	// The name of the directory INSIDE THE TEST EXECUTION VOLUME where artifacts are being
	//  a) stored using the initializer and b) retrieved using the files artifact expander
	artifactCacheDirname = "artifact-cache"
)

/*
An interface for interacting with the artifact cache directory that exists inside the suite execution volume,
	and is a) populated by the initializer and b) accessed by the files artifact expander
 */
type ArtifactCache struct {
	// Dirpath where artifacts should be written to/read from
	artifactCacheDirpath string
}

/*
Args:
	suiteExecutionVolumeMountDirpath: Dirpath where the suite execution volume is mounted, either on the
		initializer or the files artifact expander container (depending on who's instantiating the artifact cache)
 */
func NewArtifactCache(suiteExecutionVolumeMountDirpath string) *ArtifactCache {
	artifactCacheDirpath := path.Join(suiteExecutionVolumeMountDirpath, artifactCacheDirname)
	return &ArtifactCache{artifactCacheDirpath: artifactCacheDirpath}
}

func (downloader ArtifactCache) DownloadArtifacts(artifactUrlsById map[string]string) error {
	logrus.Debug("Downloading the following artifacts with the given IDs and URLs:")
	for artifactId, artifactUrl := range artifactUrlsById {
		logrus.Debugf("- %v:%v", artifactId, artifactUrl)
	}

	// TODO Download in parallel to increase instantiation speed
	for artifactId, artifactUrl := range artifactUrlsById {
		destFilepath := downloader.GetArtifactFilepath(artifactId)
		if err := downloadArtifactToFilepath(artifactUrl, destFilepath); err != nil {
			return stacktrace.Propagate(
				err,
				"An error occurred downloading artifact from '%v' to file '%v'",
				artifactUrl,
				destFilepath)
		}
	}
	return nil
}

func (downloader ArtifactCache) GetArtifactFilepath(artifactId string) string {
	return path.Join(downloader.artifactCacheDirpath, artifactId)
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

