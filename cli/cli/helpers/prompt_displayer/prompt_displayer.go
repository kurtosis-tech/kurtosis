package prompt_displayer

import (
	"github.com/kurtosis-tech/stacktrace"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"net/mail"
)

const (
	maxLabelLength = 110
)

func DisplayConfirmationPromptAndGetBooleanResult(label string, defaultValue string) (string, error) {
	if len(label) > maxLabelLength {
		return "", stacktrace.NewError("Label '%v' is longer than the maximum allowed characters, '%v'", label, maxLabelLength)
	}

	prompt := promptui.Prompt{
		Label:       label,
		Default:     defaultValue,
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
		return "", stacktrace.Propagate(err, "An error occurred displaying the prompt")
	}
	logrus.Debugf("User input: '%v'", userInput)

	return userInput, nil
}

// ====================================================================================================
//
//	Private Helper Functions
//
// ====================================================================================================
func validateConfirmationInput(input string) error {
	if input == "" {
		return nil
	}
	_, parsingErr := mail.ParseAddress(input)
	return parsingErr
}
