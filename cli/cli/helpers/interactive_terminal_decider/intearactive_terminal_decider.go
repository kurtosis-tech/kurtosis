package interactive_terminal_decider

import (
	"github.com/mattn/go-isatty"
	"os"
)

const (
	isCIEnvironmentVar = "CI"
)

// There are certain pieces of code that require the CLI to know if it's running in an interactive TTY or not
// This helper function abstracts the logic for checking
func IsInteractiveTerminal() bool {
	isStdoutTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	// Very frustratingly, steps that run in CircleCI run interactively! This is definitely wrong, but there's no
	//  way to disable it. See:
	//  - https://discuss.circleci.com/t/how-to-globally-turn-off-tty/42836
	//  - https://github.com/serverless/serverless/issues/10599#issuecomment-1029695246
	// This means that the IsTerminal check above will actually return true on CircleCI
	// Therefore, we add an extra fallback to see if a special environment variable is present
	//  that most CI systems set
	ciEnvvarValue := os.Getenv(isCIEnvironmentVar)
	isCi := ciEnvvarValue == "true"

	return !isCi && isStdoutTerminal
}
