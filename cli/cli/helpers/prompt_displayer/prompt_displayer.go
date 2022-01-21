package prompt_displayer

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	//Valid confirm inputs
	yInput    validPromptInput = "y"
	YesInput  validPromptInput = "yes"

	//Valid not confirm inputs
	nInput   validPromptInput = "n"
	NoInput  validPromptInput = "no"
)

type validPromptInput string

var validConfirmInputs = []validPromptInput{yInput, YesInput}
var validRejectInputs = []validPromptInput{nInput, NoInput}
var allValidDecisionInputs = append(validConfirmInputs, validRejectInputs...)


func DisplayConfirmationPromptAndGetBooleanResult(label string, defaultValue validPromptInput) (bool, error) {
	prompt := promptui.Prompt{
		Label:    label,
		Default:  string(defaultValue),
		Validate: validateConfirmationInput,
	}

	userInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred displaying the prompt")
	}
	logrus.Debugf("User input: '%v'", userInput)

	userConfirmOverrideKurtosisConfig := isConfirmationInput(userInput)

	return userConfirmOverrideKurtosisConfig, nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func validateConfirmationInput(input string) error {
	isValid := contains(allValidDecisionInputs, input)
	if !isValid {
		return stacktrace.NewError(
			"Yo have entered an invalid input '%v'. "+
				"Valid inputs for confirmation: '%+v' "+
				"Valid inputs for not confirmation: '%+v'",
			input,
			getValidInputsListStrFromValidPromptInputsSlice(validConfirmInputs),
			getValidInputsListStrFromValidPromptInputsSlice(validRejectInputs))
	}

	return nil
}

func isConfirmationInput(input string) bool {
	return contains(validConfirmInputs, input)
}

func contains(s []validPromptInput, str string) bool {
	for _, v := range s {
		vStr := string(v)
		if strings.ToLower(vStr) == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func getValidInputsListStrFromValidPromptInputsSlice(validUserInputsSlice []validPromptInput) string {
	var validInputsSliceStr []string

	for _, validInput := range validUserInputsSlice {
		validInputsSliceStr = append(validInputsSliceStr, string(validInput))
	}
	validInputsListStr := strings.Join(validInputsSliceStr, `','`)
	return validInputsListStr
}
