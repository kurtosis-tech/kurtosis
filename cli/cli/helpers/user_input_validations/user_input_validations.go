package user_input_validations

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

const (
	//Valid confirm inputs
	yInput validUserInput = "y"
	YesInput validUserInput = "yes"

	//Valid not confirm inputs
	nInput validUserInput = "n"
	NotInput validUserInput = "no"

	//Valid accept sending metrics inputs
	acceptSendingMetricsInput validUserInput = "accept-sending-metrics"
	doNotAcceptSendingMetricsInputs validUserInput = "do-not-accept-sending-metrics"
)

type validUserInput string

var validConfirmInputs = []validUserInput{yInput, YesInput}
var validNotConfirmInputs = []validUserInput{nInput, NotInput}
var allConfirmationValidInputs = append(validConfirmInputs, validNotConfirmInputs...)
var userAcceptSendingMetricsValidInputs = append(validConfirmInputs, acceptSendingMetricsInput)
var userDoNotAcceptSendingMetricsValidInputs = append(validNotConfirmInputs, doNotAcceptSendingMetricsInputs)
var allAcceptSendingMetricsValidInputs = append(userAcceptSendingMetricsValidInputs, userDoNotAcceptSendingMetricsValidInputs...)

func ValidateConfirmationInput(input string) error {
	isValid := contains(allConfirmationValidInputs, input)
	if !isValid {
		return stacktrace.NewError(
			"Yo have entered an invalid input '%v'. " +
				"Valid inputs for confirmation: '%+v' " +
				"Valid inputs for not confirmation: '%+v'",
			input,
			getValidInputsListStrFromValidUserInputsSlice(validConfirmInputs),
			getValidInputsListStrFromValidUserInputsSlice(validNotConfirmInputs))
	}

	return nil
}

func IsConfirmationInput(input string) bool {
	return contains(validConfirmInputs, input)
}

func ValidateMetricsConsentInput(input string) error {
	isValid := contains(allAcceptSendingMetricsValidInputs, input)
	if !isValid {
		return stacktrace.NewError(
			"Yo have entered an invalid input '%v'. " +
				"Valid inputs to accept sending metrics: '%+v' " +
				"Valid inputs to not accept sending metrics: '%+v'",
				input,
				getValidInputsListStrFromValidUserInputsSlice(userAcceptSendingMetricsValidInputs),
				getValidInputsListStrFromValidUserInputsSlice(userDoNotAcceptSendingMetricsValidInputs))
	}
	return nil
}

func IsAcceptSendingMetricsInput(input string) bool {
	return contains(userAcceptSendingMetricsValidInputs, input)
}
// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func contains(s []validUserInput, str string) bool {
	for _, v := range s {
		vStr := string(v)
		if strings.ToLower(vStr) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func getValidInputsListStrFromValidUserInputsSlice(validUserInputsSlice []validUserInput) string{
	var validInputsSliceStr []string

	for _, validInput := range validUserInputsSlice {
		validInputsSliceStr = append(validInputsSliceStr, string(validInput))
	}
	validInputsListStr := strings.Join(validInputsSliceStr, `','`)
	return validInputsListStr
}
