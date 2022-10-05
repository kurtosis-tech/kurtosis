package startosis_validator

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

type DockerImagesValidator struct {
	kurtosisBackend *backend_interface.KurtosisBackend
}

func NewDockerImagesValidator(kurtosisBackend *backend_interface.KurtosisBackend) *DockerImagesValidator {
	return &DockerImagesValidator{
		kurtosisBackend,
	}
}

func (validator *DockerImagesValidator) ValidateIntermediateEnvironment(ctx context.Context, environment *ValidatorEnvironment) error {
	// We wait for all images to be fetched before validating
	return nil
}

func (validator *DockerImagesValidator) ValidateFinalEnvironment(ctx context.Context, environment *ValidatorEnvironment) error {
	pullErrors := make(chan error, len(environment.requiredDockerImages))
	var wg sync.WaitGroup
	for image := range environment.requiredDockerImages {
		wg.Add(1)
		go pullImageFromBackend(ctx, &wg, validator.kurtosisBackend, image, pullErrors)
	}
	wg.Wait()
	close(pullErrors)
	var wrappedErrors error
	for pullError := range pullErrors {
		if wrappedErrors == nil {
			wrappedErrors = pullError
		} else {
			wrappedErrors = fmt.Errorf(pullError.Error(), wrappedErrors)
		}
	}
	return wrappedErrors
}

func pullImageFromBackend(ctx context.Context, wg *sync.WaitGroup, backend *backend_interface.KurtosisBackend, image string, pullError chan<- error) {
	err := (*backend).PullImage(ctx, image)
	if err != nil {
		pullError <- stacktrace.Propagate(err, "Failed fetching the required image %v", image)
	}
	wg.Done()
}
