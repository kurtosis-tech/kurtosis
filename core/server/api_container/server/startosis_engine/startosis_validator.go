package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/kurtosis_instruction"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/sirupsen/logrus"
)

const (
	validationInProgressMsg = "Validating Starlark code and downloading container images - execution will begin shortly"
)

type StartosisValidator struct {
	dockerImagesValidator *startosis_validator.DockerImagesValidator

	serviceNetwork service_network.ServiceNetwork
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend, serviceNetwork service_network.ServiceNetwork) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		dockerImagesValidator,
		serviceNetwork,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructions []kurtosis_instruction.KurtosisInstruction) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer close(starlarkRunResponseLineStream)
		isValidationFailure := false

		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromProgressInfo(
			validationInProgressMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)
		environment := startosis_validator.NewValidatorEnvironment(validator.serviceNetwork.GetServiceIDs())

		isValidationFailure = isValidationFailure ||
			validator.validateAnUpdateEnvironment(instructions, environment, starlarkRunResponseLineStream)
		logrus.Debug("Finished validating environment. Moving to container image validation.")

		isValidationFailure = isValidationFailure ||
			validator.downloadAndValidateImagesAccountingForProgress(ctx, environment, starlarkRunResponseLineStream)

		if isValidationFailure {
			logrus.Debug("All image validated with errors")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
		} else {
			logrus.Debug("All image successfully validated")
		}
	}()
	return starlarkRunResponseLineStream
}

func (validator *StartosisValidator) validateAnUpdateEnvironment(instructions []kurtosis_instruction.KurtosisInstruction, environment *startosis_validator.ValidatorEnvironment, starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isValidationFailure := false
	for _, instruction := range instructions {
		err := instruction.ValidateAndUpdateEnvironment(environment)
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Error while validating instruction %v. The instruction can be found at %v", instruction.String(), instruction.GetPositionInOriginalScript().String())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			isValidationFailure = true
		}
	}
	return isValidationFailure
}

func (validator *StartosisValidator) downloadAndValidateImagesAccountingForProgress(ctx context.Context, environment *startosis_validator.ValidatorEnvironment, starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isValidationFailure := false
	var imageCurrentlyBeingValidated []string
	imageValidationStarted, imageValidationFinished, errors := validator.dockerImagesValidator.Validate(ctx, environment)

	logrus.Debug("Waiting for all images to be validated")
ReadAllChan:
	for {
		select {
		case image, isChanOpen := <-imageValidationStarted:
			if !isChanOpen {
				break ReadAllChan
			}
			logrus.Debugf("Received image validation started event: '%s'", image)
			imageCurrentlyBeingValidated = append(imageCurrentlyBeingValidated, image)
			updateProgressWithDownloadInfo(starlarkRunResponseLineStream, imageCurrentlyBeingValidated)
		case image, isChanOpen := <-imageValidationFinished:
			if !isChanOpen {
				break ReadAllChan
			}
			logrus.Debugf("Received image validation finished event: '%s'", image)
			imageCurrentlyBeingValidated = removeIfPresent(imageCurrentlyBeingValidated, image)
			updateProgressWithDownloadInfo(starlarkRunResponseLineStream, imageCurrentlyBeingValidated)
		case err, isChanOpen := <-errors:
			if !isChanOpen {
				break ReadAllChan
			}
			logrus.Debugf("Received an error during image validation: '%s'", err.Error())
			isValidationFailure = true
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Error while validating final environment of script")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
		}
	}
	return isValidationFailure
}

func updateProgressWithDownloadInfo(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, imageCurrentlyInProgress []string) {
	msgLines := []string{validationInProgressMsg}
	for _, imageName := range imageCurrentlyInProgress {
		msgLines = append(msgLines, fmt.Sprintf("Downloading %s", imageName))
	}
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromMultilineProgressInfo(
		msgLines, defaultCurrentStepNumber, defaultTotalStepsNumber)
}

func removeIfPresent(slice []string, valueToRemove string) []string {
	valueToRemoveIndex := -1
	for idx, value := range slice {
		if value == valueToRemove {
			valueToRemoveIndex = idx
			break
		}
	}
	if valueToRemoveIndex < 0 {
		logrus.Warnf("Removing a value that was not present in the slice (value: '%s', slice: '%v')", valueToRemove, slice)
		return slice
	}
	newSlice := make([]string, 0)
	newSlice = append(newSlice, slice[:valueToRemoveIndex]...)
	return append(newSlice, slice[valueToRemoveIndex+1:]...)
}
