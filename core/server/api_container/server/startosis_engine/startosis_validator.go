package startosis_engine

import (
	"context"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/service"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/service_network"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/instructions_plan"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_validator"
	"github.com/kurtosis-tech/kurtosis/core/server/commons/enclave_data_directory"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	validationInProgressMsg = "Validating Starlark code and downloading container images - execution will begin shortly"
)

type StartosisValidator struct {
	dockerImagesValidator *startosis_validator.DockerImagesValidator

	serviceNetwork    service_network.ServiceNetwork
	fileArtifactStore *enclave_data_directory.FilesArtifactStore
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend, serviceNetwork service_network.ServiceNetwork, fileArtifactStore *enclave_data_directory.FilesArtifactStore) *StartosisValidator {
	dockerImagesValidator := startosis_validator.NewDockerImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		dockerImagesValidator,
		serviceNetwork,
		fileArtifactStore,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructionsSequence []*instructions_plan.ScheduledInstruction) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer close(starlarkRunResponseLineStream)
		isValidationFailure := false

		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			validationInProgressMsg, defaultCurrentStepNumber, defaultTotalStepsNumber)

		serviceNames := validator.serviceNetwork.GetServiceNames()
		serviceNamePortIdMapping, err := getServiceNameToPortIDsMap(serviceNames, validator.serviceNetwork)
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Couldn't create validator environment as we ran into errors fetching existing services and ports")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		environment := startosis_validator.NewValidatorEnvironment(
			validator.serviceNetwork.IsNetworkPartitioningEnabled(),
			serviceNames,
			validator.fileArtifactStore.ListFiles(),
			serviceNamePortIdMapping)

		isValidationFailure = isValidationFailure ||
			validator.validateAndUpdateEnvironment(instructionsSequence, environment, starlarkRunResponseLineStream)
		logrus.Debug("Finished validating environment. Validating and downloading container images.")

		isValidationFailure = isValidationFailure ||
			validator.downloadAndValidateImagesAccountingForProgress(ctx, environment, starlarkRunResponseLineStream)

		if isValidationFailure {
			logrus.Debug("Errors encountered downloading and validating container images.")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
		} else {
			logrus.Debug("All images successfully downloaded and validated.")
		}
	}()
	return starlarkRunResponseLineStream
}

func (validator *StartosisValidator) validateAndUpdateEnvironment(instructionsSequence []*instructions_plan.ScheduledInstruction, environment *startosis_validator.ValidatorEnvironment, starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isValidationFailure := false
	for _, scheduledInstruction := range instructionsSequence {
		if scheduledInstruction.IsExecuted() {
			// no need to validate the instruction as it won't be executed in this round
			continue
		}
		instruction := scheduledInstruction.GetInstruction()
		err := instruction.ValidateAndUpdateEnvironment(environment)
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err,
				"Error while validating instruction %v. The instruction can be found at %v",
				instruction.String(),
				instruction.GetPositionInOriginalScript().String())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			isValidationFailure = true
		}
	}
	return isValidationFailure
}

func (validator *StartosisValidator) downloadAndValidateImagesAccountingForProgress(ctx context.Context, environment *startosis_validator.ValidatorEnvironment, starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isValidationFailure := false

	errors := make(chan error)
	imageValidationStarted := make(chan string)
	imageValidationFinished := make(chan string)
	go validator.dockerImagesValidator.Validate(ctx, environment, imageValidationStarted, imageValidationFinished, errors)

	numberOfImageValidated := uint32(0)
	totalImageNumberToValidate := environment.GetNumberOfContainerImages()

	waitForErrorChannelToBeClosed := make(chan bool)
	defer close(waitForErrorChannelToBeClosed)

	go func() {
		var imageCurrentlyBeingValidated []string
		// we read the three channels to update imageCurrentlyBeingValidated and return progress info back to the CLI
		// it returns when the error channel is closed. The error channel is the reference here as we don't want to
		// hide an error from the user. I.e. we don't want this function to return before the error channel is closed
		for {
			select {
			case image, isChanOpen := <-imageValidationStarted:
				if !isChanOpen {
					// the subroutine returns when the error channel is closed
					continue
				}
				logrus.Debugf("Received image validation started event: '%s'", image)
				imageCurrentlyBeingValidated = append(imageCurrentlyBeingValidated, image)
				updateProgressWithDownloadInfo(starlarkRunResponseLineStream, imageCurrentlyBeingValidated, numberOfImageValidated, totalImageNumberToValidate)
			case image, isChanOpen := <-imageValidationFinished:
				if !isChanOpen {
					// the subroutine returns when the error channel is closed
					continue
				}
				numberOfImageValidated++
				logrus.Debugf("Received image validation finished event: '%s'", image)
				imageCurrentlyBeingValidated = removeIfPresent(imageCurrentlyBeingValidated, image)
				updateProgressWithDownloadInfo(starlarkRunResponseLineStream, imageCurrentlyBeingValidated, numberOfImageValidated, totalImageNumberToValidate)
			case err, isChanOpen := <-errors:
				if !isChanOpen {
					// the error channel is the important. If it's closed, then all errors have been forwarded to
					// starlarkRunResponseLineStream and this method can return
					waitForErrorChannelToBeClosed <- true
					return
				}
				logrus.Debugf("Received an error during image validation: '%s'", err.Error())
				isValidationFailure = true
				wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Error while validating final environment of script")
				starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			}
		}
	}()

	logrus.Debug("Waiting for all images to be downloaded and validated")

	// block until the error channel is closed to make sure we forward _all_ errors to
	// starlarkRunResponseLineStream before returning
	<-waitForErrorChannelToBeClosed

	return isValidationFailure
}

func updateProgressWithDownloadInfo(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, imageCurrentlyInProgress []string, numberOfImageValidated uint32, totalNumberOfImagesToValidate uint32) {
	msgLines := []string{validationInProgressMsg}
	for _, imageName := range imageCurrentlyInProgress {
		msgLines = append(msgLines, fmt.Sprintf("Downloading %s", imageName))
	}
	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromMultilineProgressInfo(
		msgLines, numberOfImageValidated, totalNumberOfImagesToValidate)
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

func getServiceNameToPortIDsMap(serviceNames map[service.ServiceName]bool, network service_network.ServiceNetwork) (map[service.ServiceName][]string, error) {
	serviceToPrivatePortIds := make(map[service.ServiceName][]string, len(serviceNames))
	ctx := context.Background()
	for serviceName := range serviceNames {
		service, err := network.GetService(ctx, string(serviceName))
		if err != nil {
			return nil, stacktrace.NewError("An error occurred while fetching service '%s' for its private port mappings", serviceName)
		}
		serviceToPrivatePortIds[serviceName] = []string{}
		privatePorts := service.GetPrivatePorts()
		for portId := range privatePorts {
			serviceToPrivatePortIds[serviceName] = append(serviceToPrivatePortIds[serviceName], portId)
		}
	}
	return serviceToPrivatePortIds, nil
}
