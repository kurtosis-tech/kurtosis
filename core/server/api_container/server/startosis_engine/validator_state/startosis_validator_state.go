package validator_state

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
	"sync"
)

type StartosisValidatorState struct {
	requiredDockerImages map[string]bool
	kurtosisBackend      *backend_interface.KurtosisBackend
}

func NewStartosisValidatorState(kurtosisBackend *backend_interface.KurtosisBackend) *StartosisValidatorState {
	return &StartosisValidatorState{
		map[string]bool{},
		kurtosisBackend,
	}
}

func (validatorState *StartosisValidatorState) AppendRequiredDockerImage(dockerImage string) {
	validatorState.requiredDockerImages[dockerImage] = true
}

func (validatorState *StartosisValidatorState) Validate(ctx context.Context) error {
	err := validatorState.validateDockerImages(ctx)
	if err != nil {
		return stacktrace.Propagate(err, "Failed while validating images")
	}
	return nil
}

func (validatorState *StartosisValidatorState) validateDockerImages(ctx context.Context) error {
	pullErrors := make(chan error, len(validatorState.requiredDockerImages))
	var wg sync.WaitGroup
	for image := range validatorState.requiredDockerImages {
		wg.Add(1)
		go pullImageFromBackend(ctx, &wg, validatorState.kurtosisBackend, image, pullErrors)
	}
	wg.Wait()
	close(pullErrors)
	for pullError := range pullErrors {
		// TODO(victor.colombo): ValidationError
		return pullError
	}
	return nil
}

func pullImageFromBackend(ctx context.Context, wg *sync.WaitGroup, backend *backend_interface.KurtosisBackend, image string, pullError chan<- error) {
	err := (*backend).PullImage(ctx, image)
	if err != nil {
		pullError <- stacktrace.Propagate(err, "Failed fetching the required image %v", image)
	}
	wg.Done()
}
