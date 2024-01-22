package image_load

import (
	"encoding/json"

	"github.com/kurtosis-tech/stacktrace"
)

type ImageLoad struct {
	// we do this way in order to have exported fields which can be marshalled
	// and an unexported type for encapsulation
	privateImageLoad *privateImageLoad
}

// ImageLoad contains the information need for building a container image.
type privateImageLoad struct {
	// Location of the container image to load (eg. tar.gz) on the machine
	ContainerImageFilePath string
}

func NewImageLoad(containerImageFilePath string) *ImageLoad {
	internalImageLoad := &privateImageLoad{
		ContainerImageFilePath: containerImageFilePath,
	}
	return &ImageLoad{internalImageLoad}
}

func (imageLoad *ImageLoad) GetContainerImageFilePath() string {
	return imageLoad.privateImageLoad.ContainerImageFilePath
}

func (imageLoad *ImageLoad) MarshalJSON() ([]byte, error) {
	return json.Marshal(imageLoad.privateImageLoad)
}

func (imageLoad *ImageLoad) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateImageLoad{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	imageLoad.privateImageLoad = unmarshalledPrivateStructPtr
	return nil
}
