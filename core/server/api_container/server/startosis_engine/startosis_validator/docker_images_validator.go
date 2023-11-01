package startosis_validator

import (
	"context"
	"sync"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
)

const maxNumberOfConcurrentDownloads = int64(4)

type DockerImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewDockerImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend) *DockerImagesValidator {
	return &DockerImagesValidator{
		kurtosisBackend,
	}
}

// Validate validates all container images by downloading them. It is an async function, and it takes as input a
// WaitGroup that will unblock once the function is complete (as opposed to when the function returns). It allows the
// consumer to run this function synchronously by calling it and then waiting for wait group to resolve.
// In addition to the total number of container images to validate, it returns three channels:
// - One that receives an image name when this image validation starts
// - One that receives an image name when an image validation finishes
// - An error channel that receives all errors happening during validation
// Note that since it is an async function, the channels are not closed by this function, consumers need to take
// care of closing them.
func (validator *DockerImagesValidator) Validate(ctx context.Context, environment *ValidatorEnvironment, imageDownloadStarted chan<- string, imageDownloadFinished chan<- *ValidatedImage, pullErrors chan<- error) {
	// We use a buffered channel to control concurrency. We push a bool to this channel when a download starts, and
	// pop one when it finishes
	imageCurrentlyDownloading := make(chan bool, maxNumberOfConcurrentDownloads)
	defer func() {
		close(imageDownloadStarted)
		close(imageDownloadFinished)
		close(pullErrors)
		close(imageCurrentlyDownloading)
	}()

	wg := &sync.WaitGroup{}
	for image := range environment.requiredDockerImages {
		wg.Add(1)
		go fetchImageFromBackend(ctx, wg, imageCurrentlyDownloading, validator.kurtosisBackend, image, environment.imageDownloadMode, pullErrors, imageDownloadStarted, imageDownloadFinished)
	}
	wg.Wait()
	logrus.Debug("All image validation submitted, currently in progress.")
}

func fetchImageFromBackend(ctx context.Context, wg *sync.WaitGroup, imageCurrentlyDownloading chan bool, backend *backend_interface.KurtosisBackend, imageName string, imageDownloadMode image_download_mode.ImageDownloadMode, pullErrors chan<- error, imageDownloadStarted chan<- string, imageDownloadFinished chan<- *ValidatedImage) {
	logrus.Debugf("Requesting the download of image: '%s'", imageName)
	var imagePulledFromRemote bool
	var imageArch string
	defer wg.Done()
	imageCurrentlyDownloading <- true
	imageDownloadStarted <- imageName
	defer func() {
		<-imageCurrentlyDownloading
		imageDownloadFinished <- NewValidatedImage(imageName, imagePulledFromRemote, imageArch)
	}()

	logrus.Debugf("Starting the download of image: '%s'", imageName)
	imagePulledFromRemote, imageArch, err := (*backend).FetchImage(ctx, imageName, imageDownloadMode)
	if err != nil {
		logrus.Warnf("Container image '%s' download failed. Error was: '%s'", imageName, err.Error())
		pullErrors <- startosis_errors.WrapWithValidationError(err, "Failed fetching the required image '%v'.", imageName)
		return
	}
	logrus.Debugf("Container image '%s' successfully downloaded", imageName)
}
