package validator_state

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/stacktrace"
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
	pullErrors := make(chan error)
	for image := range validatorState.requiredDockerImages {
		go pullImageFromBackend(ctx, validatorState.kurtosisBackend, image, pullErrors)
	}
	for range validatorState.requiredDockerImages {
		err := <-pullErrors
		if err != nil {
			return err
		}
	}
	return nil
}

func pullImageFromBackend(ctx context.Context, backend *backend_interface.KurtosisBackend, image string, pullError chan error) {
	err := (*backend).PullImage(ctx, image)
	if err != nil {
		pullError <- stacktrace.Propagate(err, "Failed fetching the required image %v", image)
	} else {
		pullError <- nil
	}
}
