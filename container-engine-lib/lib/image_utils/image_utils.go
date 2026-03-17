package image_utils

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"

	"github.com/kurtosis-tech/stacktrace"
)

const (
	manifestFileName = "manifest.json"
)

// ImageManifest represents the structure of the manifest.json file
type ImageManifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// GetRepoTags extracts the RepoTags from a Docker image file
func GetRepoTags(imageFilePath string) ([]string, error) {
	imageFile, err := os.Open(imageFilePath)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Fail to open image file %s", imageFilePath)
	}
	defer imageFile.Close()

	gzipReader, err := gzip.NewReader(imageFile)
	if err != nil {
		return nil, stacktrace.Propagate(err, "Fail to ungzip image file %s", stacktrace.Propagate(err, "Fail to read the image files"))
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	var found = false
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, stacktrace.Propagate(err, "Fail to read the image files")
		}
		if header.Name == manifestFileName {
			found = true
			break
		}
	}

	if !found {
		return nil, stacktrace.NewError("manifest.json not found in the image")
	}

	var imageManifest []ImageManifest
	jsonDecoder := json.NewDecoder(tarReader)
	if err := jsonDecoder.Decode(&imageManifest); err != nil {
		return nil, stacktrace.Propagate(err, "Could not parse the manifest.json")
	}

	if len(imageManifest) > 1 {
		return nil, stacktrace.NewError("Image has more than 1 label/tag, don't know which one to pick: %v", imageManifest)
	} else if len(imageManifest) < 1 {
		return nil, stacktrace.NewError("Image has no label/tag")
	}

	return imageManifest[0].RepoTags, nil
}
