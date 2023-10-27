package startosis_validator

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_packages"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	maxNumberOfConcurrentDownloads = int64(4)
	defaultContainerImageFile      = "Dockerfile"
)

type ImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend

	// packageContentProvider enables retrieving the buildContext for building an image
	packageContentProvider startosis_packages.PackageContentProvider
}

func NewImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend, packageContentProvider startosis_packages.PackageContentProvider) *ImagesValidator {
	return &ImagesValidator{
		kurtosisBackend:        kurtosisBackend,
		packageContentProvider: packageContentProvider,
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
func (validator *ImagesValidator) Validate(
	ctx context.Context,
	environment *ValidatorEnvironment,
	imageValidationStarted chan<- string,
	imageValidationFinished chan<- *ValidatedImage,
	imageValidationErrors chan<- error) {
	// We use a buffered channel to control concurrency. We push a bool to this channel when a download starts, and
	// pop one when it finishes
	imageCurrentlyValidating := make(chan bool, maxNumberOfConcurrentDownloads)
	defer func() {
		close(imageValidationStarted)
		close(imageValidationFinished)
		close(imageValidationErrors)
		close(imageCurrentlyValidating)
	}()

	wg := &sync.WaitGroup{}
	for image := range environment.requiredDockerImages {
		wg.Add(1)
		go validator.fetchImageFromBackend(ctx, wg, imageCurrentlyValidating, validator.kurtosisBackend, image, imageValidationErrors, imageValidationStarted, imageValidationFinished)
	}
	for serviceName, imageBuildSpec := range environment.imagesToBuild {
		wg.Add(1)
		imageName := string(serviceName)
		go validator.buildImageUsingBackend(ctx, wg, imageCurrentlyValidating, validator.kurtosisBackend, imageName, imageBuildSpec, imageValidationErrors, imageValidationStarted, imageValidationFinished)
	}
	wg.Wait()
	logrus.Debug("All image validation submitted, currently in progress.")
}

func (validator *ImagesValidator) fetchImageFromBackend(
	ctx context.Context,
	wg *sync.WaitGroup,
	imageCurrentlyDownloading chan bool,
	backend *backend_interface.KurtosisBackend,
	imageName string,
	pullErrors chan<- error,
	imageDownloadStarted chan<- string,
	imageDownloadFinished chan<- *ValidatedImage) {
	logrus.Debugf("Requesting the download of image: '%s'", imageName)
	var imagePulledFromRemote bool
	defer wg.Done()
	imageCurrentlyDownloading <- true
	imageDownloadStarted <- imageName
	defer func() {
		<-imageCurrentlyDownloading
		imageDownloadFinished <- NewValidatedImage(imageName, imagePulledFromRemote, false)
	}()

	logrus.Debugf("Starting the download of image: '%s'", imageName)
	imagePulledFromRemote, err := (*backend).FetchImage(ctx, imageName)
	if err != nil {
		logrus.Warnf("Container image '%s' download failed. Error was: '%s'", imageName, err.Error())
		pullErrors <- startosis_errors.WrapWithValidationError(err, "Failed fetching the required image '%v'.", imageName)
		return
	}
	logrus.Debugf("Container image '%s' successfully downloaded", imageName)
}

func (validator *ImagesValidator) buildImageUsingBackend(
	ctx context.Context,
	wg *sync.WaitGroup,
	imageCurrentlyBuilding chan bool,
	backend *backend_interface.KurtosisBackend,
	imageName string,
	imageBuildSpec *image_build_spec.ImageBuildSpec,
	buildErrors chan<- error,
	imageBuildStarted chan<- string,
	imageBuildFinished chan<- *ValidatedImage) {
	logrus.Debugf("Requesting the build of image: '%s'", imageName)
	isImageBuiltLocally := true
	isPulledFromRemote := false
	defer wg.Done()
	imageCurrentlyBuilding <- true
	imageBuildStarted <- imageName
	defer func() {
		<-imageCurrentlyBuilding
		imageBuildFinished <- NewValidatedImage(imageName, isPulledFromRemote, isImageBuiltLocally)
	}()

	logrus.Debugf("Starting the build of image: '%s'", imageName)
	err := (*backend).BuildImage(ctx, imageName, imageBuildSpec)
	if err != nil {
		logrus.Warnf("Container image '%s' build failed. Error was: '%s'", imageName, err.Error())
		buildErrors <- startosis_errors.WrapWithValidationError(err, "Failed to build the required image '%v'.", imageName)
		return
	}
	logrus.Debugf("Container image '%s' successfully built", imageName)
}
