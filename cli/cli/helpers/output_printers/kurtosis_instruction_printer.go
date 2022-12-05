package output_printers

import (
	"fmt"
	"github.com/bazelbuild/buildtools/build"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
	"math"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	bazelBuildDefaultFilename = ""
	spinnerDefaultSuffix      = ""

	instructionPrefixString = "> "
	resultPrefixString      = ""

	progressBarLength = 20       // in characters
	progressBarChar   = "\u2588" // unicode for: █

	codeCommentPrefix = "# "

	clearFromCurrentPositionChar = "\033[J"
	goToBeginningOfLineChar      = "\r"
	newlineChar                  = "\n"
)

var (
	colorizeInstruction      = color.New(color.FgCyan).SprintfFunc()
	colorizeResult           = color.New(color.FgWhite).SprintfFunc()
	colorizeError            = color.New(color.FgRed).SprintfFunc()
	colorizeRunSuccessfulMsg = color.New(color.FgGreen).SprintfFunc()

	colorizeProgressBarIsDone    = color.New(color.FgGreen).SprintfFunc()
	colorizeProgressBarRemaining = color.New(color.FgWhite).SprintfFunc()
)

var (
	writer               = logrus.StandardLogger().Out
	currentTerminalIndex = int(os.Stderr.Fd()) // logrus.StandardLogger().Out writes to os.Stderr

	spinnerChar  = spinner.CharSets[11]
	spinnerSpeed = 250 * time.Millisecond
	spinnerColor = spinner.WithColor("yellow")
)

type ExecutionPrinter struct {
	lock *sync.Mutex

	isStarted bool

	isSpinnerBeingUsed           bool
	spinner                      *spinner.Spinner
	progressInfoToPrintNext      *string
	progressInfoCurrentlyPrinted *string
}

func NewExecutionPrinter() *ExecutionPrinter {
	return &ExecutionPrinter{
		lock:                         &sync.Mutex{},
		isSpinnerBeingUsed:           false,
		spinner:                      nil,
		isStarted:                    false,
		progressInfoCurrentlyPrinted: nil,
		progressInfoToPrintNext:      nil,
	}
}

func (printer *ExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	if !interactive_terminal_decider.IsInteractiveTerminal() {
		printer.isSpinnerBeingUsed = false
		logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
		return nil
	}

	// spinner setup
	printer.isSpinnerBeingUsed = true
	printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer), spinner.WithSuffix(spinnerDefaultSuffix))
	// The spinner library naively assumes the spinner suffix is a single line on the terminal. This is quite
	// restrictive as 1. it is not true on small terminal windows, and 2. it prevents us from printing multi line
	// progress info. For now hack this by adding these PreUpdate and PostUpdate functions.
	printer.spinner.PreUpdate = func(s *spinner.Spinner) {
		// Clear all progress info currently being printed, if any
		printer.clearSpinnerInfoIfNecessary()
		// update the suffix if progressInfoToPrintNext is non nil
		if printer.progressInfoToPrintNext != nil {
			printer.spinner.Suffix = *printer.progressInfoToPrintNext
		}
	}
	printer.spinner.PostUpdate = func(s *spinner.Spinner) {
		// reset progressInfoToPrintNext to avoid updating the suffix every time
		printer.progressInfoToPrintNext = nil
		// store the suffix currently being printed to progressInfoCurrentlyPrinted
		printer.progressInfoCurrentlyPrinted = &printer.spinner.Suffix
	}
	printer.startSpinnerIfUsed()
	return nil
}

func (printer *ExecutionPrinter) Stop() {
	printer.stopSpinnerIfUsed()
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut.
func (printer *ExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	// Printing is a 3 phase operation:
	// 1. stop spinner to clear the ephemeral progress info
	// 2. print whatever needs to be printed, could be nothing
	// 3. restart the spinner, potentially with an updated content
	// To avoid conflicts, we take a lock out of cautiousness (this method shouldn't be called concurrently anyway)
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	// process response payload
	if responseLine.GetInstruction() != nil {
		formattedInstruction := formatInstruction(responseLine.GetInstruction(), verbosity)
		// we separate each tuple (instruction, result) with an additional newline
		formattedInstructionWithNewline := fmt.Sprintf("\n%s", formattedInstruction)
		if err := printer.printPersistentLineToStdOut(formattedInstructionWithNewline); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction: \n%v", formattedInstruction)
		}
	} else if responseLine.GetInstructionResult() != nil {
		formattedInstructionResult := formatInstructionResult(responseLine.GetInstructionResult())
		if err := printer.printPersistentLineToStdOut(formattedInstructionResult); err != nil {
			return stacktrace.Propagate(err, "Error printing Kurtosis instruction result: \n%v", formattedInstructionResult)
		}
	} else if responseLine.GetError() != nil {
		var errorMsg string
		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", responseLine.GetError().GetExecutionError().GetErrorMessage())
		}
		formattedError := formatError(errorMsg)
		if err := printer.printPersistentLineToStdOut(formattedError); err != nil {
			return stacktrace.Propagate(err, "An error happened executing Starlark code but the error couldn't be printed to the CLI output. Error message was: \n%v", errorMsg)
		}
	} else if responseLine.GetProgressInfo() != nil {
		if printer.isSpinnerBeingUsed {
			progress := responseLine.GetProgressInfo()
			progressMsessageStr := formatProgressMessage(progress.GetCurrentStepInfo())
			progressBarStr := formatProgressBar(progress.GetCurrentStepNumber(), progress.GetTotalSteps(), progressBarChar)
			spinnerInfoString := fmt.Sprintf("   %s %s", progressBarStr, progressMsessageStr)
			printer.progressInfoToPrintNext = &spinnerInfoString
		}
	} else if responseLine.GetRunFinishedEvent() != nil {
		formattedRunOutputMessage := formatRunOutput(responseLine.GetRunFinishedEvent(), dryRun)
		formattedRunOutputMessageWithNewline := fmt.Sprintf("\n%s", formattedRunOutputMessage)
		if err := printer.printPersistentLineToStdOut(formattedRunOutputMessageWithNewline); err != nil {
			return stacktrace.Propagate(err, "Unable to print the success output message containing the serialized output object. Message was: \n%v", formattedRunOutputMessage)
		}
	}
	return nil
}

func (printer *ExecutionPrinter) printPersistentLineToStdOut(lineToPrint string) error {
	// If spinner is being used, we have to stop spinner -> print -> start spinner in order to keep the spinner at the bottom of the output
	printer.stopSpinnerIfUsed()
	if _, err := fmt.Fprintln(writer, lineToPrint); err != nil {
		return stacktrace.Propagate(err, "An error happened printing a Starlark run response line. Line was:\n%s", lineToPrint)
	}
	printer.startSpinnerIfUsed()
	return nil
}

func formatError(errorMessage string) string {
	return colorizeError(errorMessage)
}

func formatInstruction(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, verbosity run.Verbosity) string {
	var serializedInstruction string
	switch verbosity {
	case run.Brief:
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	case run.Detailed:
		serializedInstruction = formatInstructionToReadableString(instruction, true)
	case run.Executable:
		serializedInstruction = formatInstructionToExecutable(instruction)
	default:
		logrus.Warnf("Unsupported verbosity flag: '%s'. Supported values are: (%s). The instruction will be printed with the default verbosity '%s'",
			verbosity.String(), strings.Join(run.VerbosityStrings(), ", "), run.Brief.String())
		serializedInstruction = formatInstructionToReadableString(instruction, false)
	}
	return colorizeInstruction(serializedInstruction)
}

func formatInstructionResult(instructionResult *kurtosis_core_rpc_api_bindings.StarlarkInstructionResult) string {
	serializedInstructionResult := fmt.Sprintf("%s%s", resultPrefixString, instructionResult.GetSerializedInstructionResult())
	return colorizeResult(serializedInstructionResult)
}

func formatInstructionToReadableString(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction, exhaustive bool) string {
	serializedInstructionComponents := []string{instruction.GetInstructionName()}
	for _, arg := range instruction.GetArguments() {
		if exhaustive || arg.GetIsRepresentative() {
			var serializedArg string
			if arg.ArgName != nil {
				serializedArg = fmt.Sprintf("%s=%s", arg.GetArgName(), arg.GetSerializedArgValue())
			} else {
				serializedArg = arg.GetSerializedArgValue()
			}
			serializedInstructionComponents = append(serializedInstructionComponents, serializedArg)
		}
	}

	var serializedInstruction string
	if exhaustive {
		separator := fmt.Sprintf("\n%s\t", instructionPrefixString)
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	} else {
		separator := " "
		serializedInstruction = strings.Join(serializedInstructionComponents, separator)
	}
	return fmt.Sprintf("%s%s", instructionPrefixString, serializedInstruction)
}

func formatInstructionToExecutable(instruction *kurtosis_core_rpc_api_bindings.StarlarkInstruction) string {
	serializedInstruction := fmt.Sprintf(
		"from %s[%d:%d]\n%s",
		instruction.GetPosition().GetFilename(),
		instruction.GetPosition().GetLine(),
		instruction.GetPosition().GetColumn(),
		instruction.GetExecutableInstruction(),
	)
	serializedInstructionWithComment := fmt.Sprintf("%s%s", codeCommentPrefix, serializedInstruction)

	parsedInstruction, err := build.ParseDefault(bazelBuildDefaultFilename, []byte(serializedInstructionWithComment))
	if err != nil {
		logrus.Warnf("Unable to format instruction. Will print it with no indentation. Problematic instruction was: \n%v", serializedInstructionWithComment)
		return serializedInstructionWithComment
	}

	multiLineInstruction := strings.Builder{}
	for _, statement := range parsedInstruction.Stmt {
		multiLineInstruction.WriteString(build.FormatString(statement))
	}
	return multiLineInstruction.String()
}

func formatProgressMessage(messageLines []string) string {
	return strings.Join(messageLines, "\n")
}

func formatProgressBar(currentStep uint32, totalSteps uint32, progressBarChar string) string {
	threshold := 0
	if totalSteps != 0 {
		threshold = int(currentStep * progressBarLength / totalSteps)
	}
	isDone := colorizeProgressBarIsDone(strings.Repeat(progressBarChar, threshold))
	remaining := colorizeProgressBarRemaining(strings.Repeat(progressBarChar, progressBarLength-threshold))
	return fmt.Sprintf("%s%s", isDone, remaining)
}

func formatRunOutput(runFinishedEvent *kurtosis_core_rpc_api_bindings.StarlarkRunFinishedEvent, dryRun bool) string {
	if !runFinishedEvent.GetIsRunSuccessful() {
		if dryRun {
			return colorizeError("Error encountered running Starlark code in dry-run mode.")
		}
		return colorizeError("Error encountered running Starlark code.")
	}
	// run was successful
	runSuccessMsg := strings.Builder{}
	runSuccessMsg.WriteString("Starlark code successfully run")
	if dryRun {
		runSuccessMsg.WriteString(" in dry-run mode")
	}
	if runFinishedEvent.GetSerializedOutput() != "" {
		runSuccessMsg.WriteString(fmt.Sprintf(". Output was:\n%v", runFinishedEvent.GetSerializedOutput()))
	} else {
		runSuccessMsg.WriteString(". No output was returned.")
	}
	return colorizeRunSuccessfulMsg(runSuccessMsg.String())
}

func (printer *ExecutionPrinter) startSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Start()
	}
}

func (printer *ExecutionPrinter) stopSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Stop()
		printer.clearSpinnerInfoIfNecessary()
		// reset progressInfoCurrentlyPrinted to store the fact that the spinner is not printing anything at the moment
		printer.progressInfoCurrentlyPrinted = nil
	}
}

func (printer *ExecutionPrinter) clearSpinnerInfoIfNecessary() {
	// if nothing is currently being printed, nothing to do
	if printer.progressInfoCurrentlyPrinted != nil {
		// compute the entire string currently being printed, including the spinner char
		spinnerCharPlusSuffix := spinnerChar[0] + *printer.progressInfoCurrentlyPrinted
		// from this string, compute the number of lines that are currently being taken to print it entirely
		spinnerSuffixNumberOfLines := computeNumberOfLinesPrintedToTerminal(spinnerCharPlusSuffix)
		// move the cursor up the required number of lines, and erase all content from this position
		fmt.Print(moveLinesUpLeftStr(spinnerSuffixNumberOfLines - 1))
		fmt.Print(clearFromCurrentPositionChar)
	}
}

// computeNumberOfLinesPrintedToTerminal computes the number of lines that were required to print the given string to
// the current terminal
func computeNumberOfLinesPrintedToTerminal(stringToPrint string) int {
	if term.IsTerminal(currentTerminalIndex) {
		terminalWidth, _, err := term.GetSize(currentTerminalIndex)
		if err != nil {
			// We assume infinite terminal width here because it's the less destructive approach. If we were assuming
			// terminal width = 80 chars by default, long suffix lines would cause us to erase multiple lines,
			// potentially clearing valuable printed lines that are not part of the progress info.
			// Assuming infinite means we will erase at most a single line, which we're sure contains progress info.
			logrus.Errorf("Unable to get width of terminal. Will assume infinite. Error was: %v", err.Error())
			return computeNumberOfLinesInString(stringToPrint, math.MaxInt)
		}
		return computeNumberOfLinesInString(stringToPrint, terminalWidth)
	}
	return computeNumberOfLinesInString(stringToPrint, math.MaxInt)
}

// computeNumberOfLinesInString computes the number of lines needed to print the string with a maxWidth allowed
// This is mostly to allow unit testing computeNumberOfLinesPrintedToTerminal
func computeNumberOfLinesInString(stringToPrint string, maxWidth int) int {
	if stringToPrint == "" {
		// empty string will necessarily take one line
		return 1
	}
	idxOfNewline := strings.Index(stringToPrint, newlineChar)
	if idxOfNewline < 0 {
		// we use utf8.RunCountInString() in place of len() because the string contains "complex" unicode chars that
		// might be represented by multiple individual bytes (such as the spinner char)
		return int(math.Ceil(float64(utf8.RuneCountInString(stringToPrint)) / float64(maxWidth)))
	} else {
		return computeNumberOfLinesInString(stringToPrint[:idxOfNewline], maxWidth) + computeNumberOfLinesInString(stringToPrint[idxOfNewline+1:], maxWidth)
	}
}

// moveLinesUpLeftStr generated the control string necessary to move the cursor numberOfLines up, and move it at the
// beginning of the line
func moveLinesUpLeftStr(numberOfLines int) string {
	if numberOfLines > 0 {
		// Character `\033[#A` moves the cursor `#` lines up. We here insert the right #
		return fmt.Sprintf("\033[%dA%s", numberOfLines, goToBeginningOfLineChar)
	} else {
		return goToBeginningOfLineChar
	}
}
