package startosis_engine

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/binding_constructors"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface"
	"github.com/kurtosis-tech/kurtosis/container-engine-lib/lib/backend_interface/objects/image_download_mode"
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
	validationInProgressMsg = "Validating plan and preparing container images - execution will begin shortly"

	containerImageValidationMsgHeader     = "Container images used in this run:"
	containerImageValidationMsgFromLocal  = "locally cached"
	containerImageValidationMsgFromRemote = "remotely downloaded"
	containerImageValidationBuilt         = "locally built"
	containerImageValidationMsgLineFormat = "> %s - %s"

	containerImageArchWarningHeaderFormat   = "WARNING: Container images with different architecture than expected(%s):"
	containerImageArchitectureMsgLineFormat = "> %s - %s"

	linebreak = "\n"

	emptyImageArchitecture = ""
)

type StartosisValidator struct {
	imagesValidator *startosis_validator.ImagesValidator

	serviceNetwork    service_network.ServiceNetwork
	fileArtifactStore *enclave_data_directory.FilesArtifactStore

	backend *backend_interface.KurtosisBackend
}

func NewStartosisValidator(kurtosisBackend *backend_interface.KurtosisBackend, serviceNetwork service_network.ServiceNetwork, fileArtifactStore *enclave_data_directory.FilesArtifactStore) *StartosisValidator {
	imagesValidator := startosis_validator.NewImagesValidator(kurtosisBackend)
	return &StartosisValidator{
		imagesValidator,
		serviceNetwork,
		fileArtifactStore,
		kurtosisBackend,
	}
}

func (validator *StartosisValidator) Validate(ctx context.Context, instructionsSequence []*instructions_plan.ScheduledInstruction, imageDownloadMode image_download_mode.ImageDownloadMode) <-chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine {
	starlarkRunResponseLineStream := make(chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine)
	go func() {
		defer close(starlarkRunResponseLineStream)
		isValidationFailure := false

		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromSinglelineProgressInfo(
			validationInProgressMsg, defaultCurrentStepNumber, defaultTotalStepsNumber, "validatation")

		serviceNames, err := validator.serviceNetwork.GetServiceNames()
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "An error occurred getting all service names")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		serviceNamePortIdMapping, err := getServiceNameToPortIDsMap(serviceNames, validator.serviceNetwork)
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Couldn't create validator environment as we ran into errors fetching existing services and ports")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		availableMemoryInMegaBytes, availableCpuInMilliCores, isResourceInformationComplete, err := (*validator.backend).GetAvailableCPUAndMemory(ctx)
		if err != nil {
			wrappedValidationError := startosis_errors.WrapWithValidationError(err, "Couldn't create validator environment as we ran into errors fetching information about available cpu & memory")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromValidationError(wrappedValidationError.ToAPIType())
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
			return
		}

		environment := startosis_validator.NewValidatorEnvironment(
			serviceNames,
			validator.fileArtifactStore.ListFiles(),
			serviceNamePortIdMapping,
			availableCpuInMilliCores,
			availableMemoryInMegaBytes,
			isResourceInformationComplete,
			imageDownloadMode)

		isValidationFailure = isValidationFailure ||
			validator.validateAndUpdateEnvironment(instructionsSequence, environment, starlarkRunResponseLineStream)
		logrus.Debug("Finished validating environment. Validating container images...")

		isValidationFailure = isValidationFailure ||
			validator.validateImagesAccountingForProgress(ctx, environment, starlarkRunResponseLineStream)

		if isValidationFailure {
			logrus.Debug("Errors encountered validating container images.")
			starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromRunFailureEvent()
		} else {
			logrus.Debug("All images successfully validated.")
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

func (validator *StartosisValidator) validateImagesAccountingForProgress(ctx context.Context, environment *startosis_validator.ValidatorEnvironment, starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine) bool {
	isValidationFailure := false

	errors := make(chan error)
	imageValidationStarted := make(chan string)
	imageValidationFinished := make(chan *startosis_validator.ValidatedImage)
	go validator.imagesValidator.Validate(ctx, environment, imageValidationStarted, imageValidationFinished, errors)

	numberOfImageValidated := uint32(0)
	totalImageNumberToValidate := environment.GetNumberOfContainerImagesToProcess()

	waitForErrorChannelToBeClosed := make(chan bool)
	defer close(waitForErrorChannelToBeClosed)

	go func() {
		var imageCurrentlyBeingValidated []string
		imageSuccessfullyValidated := map[string]*startosis_validator.ValidatedImage{}
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
			case validatedImage, isChanOpen := <-imageValidationFinished:
				if !isChanOpen {
					// the subroutine returns when the error channel is closed
					continue
				}
				numberOfImageValidated++

				imageName := validatedImage.GetName()
				logrus.Debugf("Received image validation finished event: '%s'", imageName)

				imageCurrentlyBeingValidated = removeIfPresent(imageCurrentlyBeingValidated, imageName)

				imageSuccessfullyValidated[imageName] = validatedImage

				updateProgressWithDownloadInfo(starlarkRunResponseLineStream, imageCurrentlyBeingValidated, numberOfImageValidated, totalImageNumberToValidate)
			case err, isChanOpen := <-errors:
				if !isChanOpen {
					sendContainerImageSummaryInfoMsg(imageSuccessfullyValidated, starlarkRunResponseLineStream)
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

func sendContainerImageSummaryInfoMsg(
	imageSuccessfullyValidated map[string]*startosis_validator.ValidatedImage,
	starlarkRunResponseLineStream chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine,
) {
	if len(imageSuccessfullyValidated) == 0 {
		return
	}

	imageLines := []string{}
	imagesWithIncorrectArchLines := []string{}

	for image, validatedImage := range imageSuccessfullyValidated {
		imageValidationStr := containerImageValidationMsgFromLocal
		if validatedImage.IsPulledFromRemote() {
			imageValidationStr = containerImageValidationMsgFromRemote
		}

		if validatedImage.IsBuiltLocally() {
			imageValidationStr = containerImageValidationBuilt
		}

		imageLine := fmt.Sprintf(containerImageValidationMsgLineFormat, image, imageValidationStr)
		imageLines = append(imageLines, imageLine)

		architecture := validatedImage.GetArchitecture()

		if architecture == emptyImageArchitecture {
			logrus.Debugf("Couldn't fetch image architecture for '%v'; this is expected on k8s backend but not docker", image)
		} else if architecture != runtime.GOARCH {
			imageWithIncorrectArchLine := fmt.Sprintf(containerImageArchitectureMsgLineFormat, image, architecture)
			imagesWithIncorrectArchLines = append(imagesWithIncorrectArchLines, imageWithIncorrectArchLine)
		}
	}

	msgLines := []string{containerImageValidationMsgHeader}
	msgLines = append(msgLines, imageLines...)

	msg := strings.Join(msgLines, linebreak)

	starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromInfoMsg(msg)

	if len(imagesWithIncorrectArchLines) > 0 {
		imageWarningHeader := fmt.Sprintf(containerImageArchWarningHeaderFormat, runtime.GOARCH)
		imagesWithArchMsgLines := []string{imageWarningHeader}
		imagesWithArchMsgLines = append(imagesWithArchMsgLines, imagesWithIncorrectArchLines...)
		imagesWithDiffArchWarningMessage := strings.Join(imagesWithArchMsgLines, linebreak)
		starlarkRunResponseLineStream <- binding_constructors.NewStarlarkRunResponseLineFromWarning(imagesWithDiffArchWarningMessage)
	}
}

func updateProgressWithDownloadInfo(starlarkRunResponseLineStream chan<- *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, imageCurrentlyInProgress []string, numberOfImageValidated uint32, totalNumberOfImagesToValidate uint32) {
	msgLines := []string{validationInProgressMsg}
	for _, imageName := range imageCurrentlyInProgress {
		msgLines = append(msgLines, fmt.Sprintf("Validating %s", imageName))
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
