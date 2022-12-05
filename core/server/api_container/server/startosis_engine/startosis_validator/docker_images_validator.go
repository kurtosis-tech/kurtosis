package startosis_validator

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
	"sync"
)

var (
	maxNumberOfConcurrentDownloads = int64(4)
)

type DockerImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewDockerImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend) *DockerImagesValidator {
	return &DockerImagesValidator{
		kurtosisBackend,
	}
}

// Validate validates all container images by downloading them.
// It returns three channels:
// - One that receives an image name when this image validation starts
// - One that receives an image name when an image validation finishes
// - An error channel that receives all errors happening during validation
func (validator *DockerImagesValidator) Validate(ctx context.Context, environment *ValidatorEnvironment) (<-chan string, <-chan string, <-chan error) {
	pullErrors := make(chan error)
	imageDownloadStarted := make(chan string)
	imageDownloadFinished := make(chan string)

	// We use a buffered channel to control concurrency. We push a bool to this channel when a download starts, and
	// pop one when it finishes
	imageCurrentlyDownloading := make(chan bool, maxNumberOfConcurrentDownloads)
	var wg sync.WaitGroup

	go func() {
		// The condition that allows us to close the channels is the WaitGroup returning.
		// We need to make sure it always return, which is the case here since we decrease it immediately in
		// fetchImageFromBackend
		wg.Wait()
		close(pullErrors)
		close(imageCurrentlyDownloading)
		close(imageDownloadStarted)
		close(imageDownloadFinished)
	}()

	for image := range environment.requiredDockerImages {
		wg.Add(1)
		logrus.Debugf("Starting the download of image: '%s'", image)
		go fetchImageFromBackend(ctx, &wg, imageCurrentlyDownloading, validator.kurtosisBackend, image, pullErrors, imageDownloadStarted, imageDownloadFinished)
	}
	logrus.Debug("All image validation submitted, currently in progress.")
	return imageDownloadStarted, imageDownloadFinished, pullErrors
}

func fetchImageFromBackend(ctx context.Context, wg *sync.WaitGroup, imageCurrentlyDownloading chan bool, backend *backend_interface.KurtosisBackend, image string, pullErrors chan<- error, imageDownloadStarted chan<- string, imageDownloadFinished chan<- string) {
	defer wg.Done()
	imageCurrentlyDownloading <- true
	imageDownloadStarted <- image
	defer func() {
		<-imageCurrentlyDownloading
		imageDownloadFinished <- image
	}()

	err := (*backend).FetchImage(ctx, image)
	if err != nil {
		pullErrors <- startosis_errors.NewValidationError("Failed fetching the required image '%v', make sure that the image exists and is public", image)
	}
	logrus.Debugf("Container image '%s' successfully downloaded", image)
}
