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

// ValidateAsync validates all container images by downloading them. It is an async function, and it takes as input a
// WaitGroup that will unblock once the function is complete (as opposed to when the function returns). It allows the
// consumer to run this function synchronously by calling it and then waiting for wait group to resolve.
// It returns three channels:
// - One that receives an image name when this image validation starts
// - One that receives an image name when an image validation finishes
// - An error channel that receives all errors happening during validation
// Note that since it is an async function, the channels aren't not closed by this function, consumers need to take
// care of closing them.
func (validator *DockerImagesValidator) ValidateAsync(ctx context.Context, environment *ValidatorEnvironment, wg *sync.WaitGroup) (chan string, chan string, chan error) {
	pullErrors := make(chan error)
	imageDownloadStarted := make(chan string)
	imageDownloadFinished := make(chan string)

	// We use a buffered channel to control concurrency. We push a bool to this channel when a download starts, and
	// pop one when it finishes
	imageCurrentlyDownloading := make(chan bool, maxNumberOfConcurrentDownloads)

	for image := range environment.requiredDockerImages {
		wg.Add(1)
		logrus.Debugf("Starting the download of image: '%s'", image)
		go fetchImageFromBackend(ctx, wg, imageCurrentlyDownloading, validator.kurtosisBackend, image, pullErrors, imageDownloadStarted, imageDownloadFinished)
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
