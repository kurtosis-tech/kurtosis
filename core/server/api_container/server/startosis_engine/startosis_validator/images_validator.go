package startosis_validator

import (
	"context"
	"sync"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_build_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_registry_spec"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/nix_build_spec"

	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/sirupsen/logrus"
)

const maxNumberOfConcurrentDownloads = int64(4)

type ImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend) *ImagesValidator {
	return &ImagesValidator{
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
	for imageName, maybeImageRegistrySpec := range environment.imagesToPull {
		wg.Add(1)
		go fetchImageFromBackend(ctx, wg, imageCurrentlyValidating, validator.kurtosisBackend, imageName, maybeImageRegistrySpec, environment.imageDownloadMode, imageValidationErrors, imageValidationStarted, imageValidationFinished)
	}
	for imageName, imageBuildSpec := range environment.imagesToBuild {
		wg.Add(1)
		go validator.buildImageUsingBackend(ctx, wg, imageCurrentlyValidating, validator.kurtosisBackend, imageName, imageBuildSpec, imageValidationErrors, imageValidationStarted, imageValidationFinished)
	}
	for imageName, nixBuildSpec := range environment.nixToBuild {
		wg.Add(1)
		logrus.Warnf("%v - %v", imageName, nixBuildSpec)
		go validator.nixBuildUsingBackend(ctx, wg, imageCurrentlyValidating, validator.kurtosisBackend, imageName, nixBuildSpec, imageValidationErrors, imageValidationStarted, imageValidationFinished)
	}
	wg.Wait()
	logrus.Debug("All image validation submitted, currently in progress.")
}

func fetchImageFromBackend(ctx context.Context, wg *sync.WaitGroup, imageCurrentlyDownloading chan bool, backend *backend_interface.KurtosisBackend, imageName string, registrySpec *image_registry_spec.ImageRegistrySpec, imageDownloadMode image_download_mode.ImageDownloadMode, pullErrors chan<- error, imageDownloadStarted chan<- string, imageDownloadFinished chan<- *ValidatedImage) {
	logrus.Debugf("Requesting the download of image: '%s'", imageName)
	var imagePulledFromRemote bool
	var imageArch string
	imageBuiltLocally := false
	defer wg.Done()
	imageCurrentlyDownloading <- true
	imageDownloadStarted <- imageName
	defer func() {
		<-imageCurrentlyDownloading
		imageDownloadFinished <- NewValidatedImage(imageName, imagePulledFromRemote, imageBuiltLocally, imageArch)
	}()

	logrus.Debugf("Starting the download of image: '%s'", imageName)
	imagePulledFromRemote, imageArch, err := (*backend).FetchImage(ctx, imageName, registrySpec, imageDownloadMode)
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
	var imageArch string
	imageBuiltLocally := true
	imagePulledFromRemote := false
	defer wg.Done()
	imageCurrentlyBuilding <- true
	imageBuildStarted <- imageName
	defer func() {
		<-imageCurrentlyBuilding
		imageBuildFinished <- NewValidatedImage(imageName, imagePulledFromRemote, imageBuiltLocally, imageArch)
	}()

	logrus.Debugf("Starting the build of image: '%s'", imageName)
	imageArch, err := (*backend).BuildImage(ctx, imageName, imageBuildSpec)
	if err != nil {
		logrus.Warnf("Container image '%s' build failed. Error was: '%s'", imageName, err.Error())
		buildErrors <- startosis_errors.WrapWithValidationError(err, "Failed to build the required image '%v'.", imageName)
		return
	}
	logrus.Debugf("Container image '%s' successfully built", imageName)
}

func (validator *ImagesValidator) nixBuildUsingBackend(
	ctx context.Context,
	wg *sync.WaitGroup,
	imageCurrentlyBuilding chan bool,
	backend *backend_interface.KurtosisBackend,
	imageName string,
	nixBuildSpec *nix_build_spec.NixBuildSpec,
	buildErrors chan<- error,
	nixBuildStarted chan<- string,
	nixBuildFinished chan<- *ValidatedImage) {
	logrus.Debugf("Requesting the build of image: '%s'", imageName)
	var imageArch string
	imageBuiltLocally := true
	imagePulledFromRemote := false
	defer wg.Done()
	imageCurrentlyBuilding <- true
	nixBuildStarted <- imageName
	defer func() {
		<-imageCurrentlyBuilding
		nixBuildFinished <- NewValidatedImage(imageName, imagePulledFromRemote, imageBuiltLocally, imageArch)
	}()

	logrus.Debugf("Starting the build of image: '%s'", imageName)
	imageArch, err := (*backend).NixBuild(ctx, nixBuildSpec)
	if err != nil {
		logrus.Warnf("Container image '%s' build failed. Error was: '%s'", imageName, err.Error())
		buildErrors <- startosis_errors.WrapWithValidationError(err, "Failed to build the required image '%v'.", imageName)
		return
	}
	logrus.Debugf("Container image '%s' successfully built", imageName)
}
