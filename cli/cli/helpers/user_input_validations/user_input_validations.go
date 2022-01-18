package user_input_validations

import (
	"github.com/kurtosis-tech/stacktrace"
	"strings"
)

var userAcceptSendingMetricsValidInputs = []string{"y", "yes", "accept-sending-metrics"}
var userDoNotAcceptSendingMetricsValidInputs = []string{"n", "no", "do-not-accept-sending-metrics"}
var allAcceptSendingMetricsValidInputs = append(userAcceptSendingMetricsValidInputs, userDoNotAcceptSendingMetricsValidInputs...)


func ValidateMetricsConsentInput(input string) error {
	input = strings.ToLower(input)
	isValid := contains(allAcceptSendingMetricsValidInputs, input)
	if !isValid {
		return stacktrace.NewError(
			"Yo have entered an invalid input '%v'. " +
				"Valid inputs to accept sending metrics: '%+v' " +
				"Valid inputs to not accept sending metrics: '%+v'",
				input,
				strings.Join(userAcceptSendingMetricsValidInputs, `','`),
				strings.Join(userDoNotAcceptSendingMetricsValidInputs, `','`))
	}
	return nil
}

func IsAcceptedSendingMetricsValidInput(input string) bool {
	return contains(userAcceptSendingMetricsValidInputs, input)
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func contains(s []string, str string) bool {
	for _, v := range s {
		if strings.ToLower(v) == strings.ToLower(str) {
			return true
		}
	}
	return false
}
