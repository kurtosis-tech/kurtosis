package prompt_displayer

import (
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	//Valid confirm inputs
	yInput   validPromptInput = "y"
	yesInput validPromptInput = "yes"

	//Valid not confirm inputs
	nInput  validPromptInput = "n"
	noInput validPromptInput = "no"

	// Anything beyond this and it runs off the edge of the screen so the user doesn't see it
	maxLabelLength = 110
)

type validPromptInput string

var validConfirmInputs = []validPromptInput{yInput, yesInput}
var validRejectInputs = []validPromptInput{nInput, noInput}
var allValidDecisionInputs = append(validConfirmInputs, validRejectInputs...)

func DisplayConfirmationPromptAndGetBooleanResult(label string, defaultValue bool) (bool, error) {
	if len(label) > maxLabelLength {
		return false, stacktrace.NewError("Label '%v' is longer than the maximum allowed characters, '%v'", label, maxLabelLength)
	}

	defaultValueStr := string(nInput)
	if defaultValue {
		defaultValueStr = string(yInput)
	}

	labelWithValidInputs := fmt.Sprintf(
		"%v (%v/%v)",
		label,
		yInput,
		nInput,
	)
	prompt := promptui.Prompt{
		Label:       labelWithValidInputs,
		Default:     defaultValueStr,
		AllowEdit:   false,
		Validate:    validateConfirmationInput,
		Mask:        0,
		HideEntered: false,
		Templates:   nil,
		IsConfirm:   false,
		IsVimMode:   false,
		Pointer:     nil,
		Stdin:       nil,
		Stdout:      nil,
	}

	userInput, err := prompt.Run()
	if err != nil {
		return false, stacktrace.Propagate(err, "An error occurred displaying the prompt")
	}
	logrus.Debugf("User input: '%v'", userInput)

	didUserConfirm := isConfirmationInput(userInput)

	return didUserConfirm, nil
}

// ====================================================================================================
//                                       Private Helper Functions
// ====================================================================================================
func validateConfirmationInput(input string) error {
	isValid := contains(allValidDecisionInputs, input)
	if !isValid {
		return stacktrace.NewError(
			"You have entered an invalid input '%v'. "+
				"Valid inputs for confirmation: '%v' "+
				"Valid inputs for rejection: '%v'",
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
		if strings.EqualFold(vStr, str) {
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
