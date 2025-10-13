package output_printers

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/bubbles/progress"
	bubbleteaSpinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/kurtosis_core_rpc_api_bindings"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/cli/cli/out"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
)

type ParallelExecutionPrinter struct {
	lock *sync.Mutex

	isStarted bool

	isSpinnerBeingUsed bool
	spinner            *spinner.Spinner

	// bubbletea library components for rendering the TUI
	bubbleteaModel   *ExecutionModel
	bubbleteaProgram *tea.Program
	bubbleteaMsgChan chan tea.Msg
}

func NewParallelExecutionPrinter() *ParallelExecutionPrinter {
	return &ParallelExecutionPrinter{
		lock:               &sync.Mutex{},
		isStarted:          false,
		isSpinnerBeingUsed: false,
		spinner:            nil,

		bubbleteaMsgChan: make(chan tea.Msg, 100),
	}
}

func (printer *ParallelExecutionPrinter) Start() error {
	if printer.isStarted {
		return stacktrace.NewError("printer already started")
	}
	printer.isStarted = true
	// TODO: figure out if we need this?
	// if !interactive_terminal_decider.IsInteractiveTerminal() {
	// 	printer.isSpinnerBeingUsed = false
	// 	logrus.Infof("Kurtosis CLI is running in a non interactive terminal. Everything will work but progress information and the progress bar will not be displayed.")
	// 	return nil
	// }

	// Initialize bubbletea model and program
	printer.bubbleteaModel = NewExecutionModel()
	printer.bubbleteaProgram = tea.NewProgram(
		printer.bubbleteaModel,
		tea.WithMouseCellMotion(),
	)

	// Start the bubbletea program in a goroutine
	go func() {
		if _, err := printer.bubbleteaProgram.Run(); err != nil {
			logrus.Errorf("Error running bubbletea program: %v", err)
		}
	}()

	// Start message processing goroutine
	go printer.processMessages()

	// spinner setup
	printer.isSpinnerBeingUsed = true
	printer.spinner = spinner.New(spinnerChar, spinnerSpeed, spinnerColor, spinner.WithWriter(writer), spinner.WithSuffix(spinnerDefaultSuffix))
	printer.startSpinnerIfUsed()
	return nil
}

func (printer *ParallelExecutionPrinter) Stop() {
	printer.stopSpinnerIfUsed()
	// Give a moment for final messages to be processed and rendered
	time.Sleep(100 * time.Millisecond)

	// Properly quit the bubbletea program to release terminal control
	printer.bubbleteaProgram.Quit()

	// Close the message channel
	close(printer.bubbleteaMsgChan)
	printer.isStarted = false
}

// PrintKurtosisExecutionResponseLineToStdOut format and prints the instruction to StdOut.
// TODO: consider refactoring this?
func (printer *ParallelExecutionPrinter) PrintKurtosisExecutionResponseLineToStdOut(responseLine *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, verbosity run.Verbosity, dryRun bool) error {
	printer.lock.Lock()
	defer printer.lock.Unlock()

	if !printer.isStarted {
		return stacktrace.NewError("Cannot print with a non started printer")
	}

	var msg tea.Msg
	if responseLine.GetInstruction() != nil && verbosity != run.OutputOnly {
		instruction := responseLine.GetInstruction()
		instructionId := instruction.GetInstructionId()
		msg = InstructionStartedMsg{
			ID:   instructionId,
			Name: formatInstruction(instruction, verbosity),
		}
	} else if responseLine.GetInstructionResult() != nil {
		result := responseLine.GetInstructionResult()
		instructionId := result.GetInstructionId()
		msg = InstructionCompletedMsg{
			ID:     instructionId,
			Result: formatInstructionResult(result, verbosity),
		}
	} else if responseLine.GetError() != nil {
		var errorMsg string
		var instructionId string

		if responseLine.GetError().GetInterpretationError() != nil {
			errorMsg = fmt.Sprintf("There was an error interpreting Starlark code \n%v", responseLine.GetError().GetInterpretationError().GetErrorMessage())
			instructionId = "interpretation"
		} else if responseLine.GetError().GetValidationError() != nil {
			errorMsg = fmt.Sprintf("There was an error validating Starlark code \n%v", responseLine.GetError().GetValidationError().GetErrorMessage())
			instructionId = "validation"
		} else if responseLine.GetError().GetExecutionError() != nil {
			errorMsgWithStackTrace := errors.New(responseLine.GetError().GetExecutionError().GetErrorMessage())
			cleanedErrorFromStarlark := out.GetErrorMessageToBeDisplayedOnCli(errorMsgWithStackTrace)
			errorMsg = fmt.Sprintf("There was an error executing Starlark code \n%v", cleanedErrorFromStarlark)
			instructionId = "execution"
		}

		msg = InstructionFailedMsg{
			ID:    instructionId,
			Error: errorMsg,
		}
	} else if responseLine.GetProgressInfo() != nil {
		progress := responseLine.GetProgressInfo()
		progressRatio := float64(progress.GetCurrentStepNumber()) / float64(progress.GetTotalSteps())
		instructionId := progress.GetInstructionId()
		msg = InstructionProgressMsg{
			ID:       instructionId,
			Progress: progressRatio,
			Message:  formatProgressMessage(progress.GetCurrentStepInfo()),
		}
	} else if responseLine.GetRunFinishedEvent() != nil {
		runFinished := responseLine.GetRunFinishedEvent()
		output := formatRunOutput(runFinished, dryRun, verbosity)
		// err := os.WriteFile("/Users/tewodrosmitiku/craft/kurtosis/output.txt", []byte(output), 0644)
		// if err != nil {
		// 	logrus.Errorf("Error writing output to file: %v", err)
		// }
		msg = ExecutionCompleteMsg{
			ID:      "execution",
			Result:  output,
			Success: runFinished.GetIsRunSuccessful(),
			Error:   nil,
		}
	} else if responseLine.GetWarning() != nil {
		warning := responseLine.GetWarning()
		instructionId := "execution"
		msg = InstructionWarningMsg{
			ID:      instructionId,
			Warning: formatWarning(warning.GetWarningMessage()),
		}
	} else if responseLine.GetInfo() != nil {
		info := responseLine.GetInfo()
		instructionId := "execution"
		msg = InstructionInfoMsg{
			ID:   instructionId,
			Info: formatInfo(info.GetInfoMessage()),
		}
	}
	if msg != nil {
		select {
		case printer.bubbleteaMsgChan <- msg:
			// Message sent successfully
		default:
			// Channel full, log warning but don't block
			logrus.Warnf("Message channel full, dropping message")
		}
	}

	// No message to send for unknown response types
	return nil
}

// processMessages handles bubbletea messages from the channel
func (printer *ParallelExecutionPrinter) processMessages() {
	for msg := range printer.bubbleteaMsgChan {
		if printer.bubbleteaProgram != nil {
			printer.bubbleteaProgram.Send(msg)
		}
	}
}

func (printer *ParallelExecutionPrinter) startSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Start()
	}
}

func (printer *ParallelExecutionPrinter) stopSpinnerIfUsed() {
	if printer.isSpinnerBeingUsed {
		printer.spinner.Stop()
	}
}

// InstructionStatus represents the current state of an instruction
type InstructionStatus int

const (
	StatusPending InstructionStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

// InstructionState holds the state of a single instruction
type InstructionState struct {
	ID              string
	Name            string
	Status          InstructionStatus
	Progress        float64
	ProgressBar     progress.Model
	Spinner         bubbleteaSpinner.Model
	ErrorMessage    string
	Result          string
	WarningMessages []string
	InfoMessages    []string
}

// ExecutionModel is the main bubbletea model for parallel instruction display
type ExecutionModel struct {
	// Instruction tracking
	instructions     map[string]*InstructionState
	instructionOrder []string

	// UI state
	width, height int
	isInteractive bool

	// Program control
	done bool
}

// Message types for bubbletea

// InstructionStartedMsg is sent when a new instruction begins
type InstructionStartedMsg struct {
	ID   string
	Name string
}

// InstructionProgressMsg is sent when an instruction reports progress
type InstructionProgressMsg struct {
	ID       string
	Progress float64
	Message  string
}

// InstructionCompletedMsg is sent when an instruction completes successfully
type InstructionCompletedMsg struct {
	ID     string
	Result string
}

// InstructionFailedMsg is sent when an instruction fails
type InstructionFailedMsg struct {
	ID    string
	Error string
}

// InstructionWarningMsg is sent when an instruction reports a warning
type InstructionWarningMsg struct {
	ID      string
	Warning string
}

// InstructionInfoMsg is sent when an instruction reports info
type InstructionInfoMsg struct {
	ID   string
	Info string
}

// WindowSizeMsg is sent when the terminal window is resized
type WindowSizeMsg struct {
	Width, Height int
}

// ExecutionCompleteMsg is sent when the entire execution is complete
type ExecutionCompleteMsg struct {
	ID      string
	Result  string
	Success bool
	Error   error
}

// ProgressTickMsg is sent periodically to update progress bars
type ProgressTickMsg struct {
	Time time.Time
}

// NewExecutionModel creates a new ExecutionModel
func NewExecutionModel() *ExecutionModel {
	instructions := make(map[string]*InstructionState)
	s := bubbleteaSpinner.New()
	s.Spinner = bubbleteaSpinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	instructions["execution"] = &InstructionState{
		ID:              "execution",
		Name:            "Executing Starlark code",
		Status:          StatusRunning,
		Progress:        0.1, // Start with 10% progress
		Result:          "",
		ProgressBar:     progress.New(progress.WithGradient("#008000", "#C0C0C0")),
		Spinner:         bubbleteaSpinner.New(),
		WarningMessages: make([]string, 0),
		InfoMessages:    make([]string, 0),
		ErrorMessage:    "",
	}
	instructions["validation"] = &InstructionState{
		ID:              "validation",
		Name:            "Validating Starlark code",
		Status:          StatusRunning,
		Progress:        0.1, // Start with 10% progress
		Result:          "",
		ProgressBar:     progress.New(progress.WithGradient("#008000", "#C0C0C0")),
		Spinner:         s,
		WarningMessages: make([]string, 0),
		InfoMessages:    make([]string, 0),
		ErrorMessage:    "",
	}
	instructions["interpretation"] = &InstructionState{
		ID:              "interpretation",
		Name:            "Interpreting Starlark code",
		Status:          StatusRunning,
		Progress:        0.1, // Start with 10% progress
		Result:          "",
		ProgressBar:     progress.New(progress.WithGradient("#008000", "#C0C0C0")),
		Spinner:         s,
		WarningMessages: make([]string, 0),
		InfoMessages:    make([]string, 0),
		ErrorMessage:    "",
	}
	return &ExecutionModel{
		instructions:     instructions,
		instructionOrder: []string{"execution"},
		done:             false,
		isInteractive:    interactive_terminal_decider.IsInteractiveTerminal(),
	}
}

// Init implements tea.Model
func (m *ExecutionModel) Init() tea.Cmd {
	// Start the execution spinner and progress ticker
	var cmds []tea.Cmd

	if execution, exists := m.instructions["execution"]; exists {
		cmds = append(cmds, execution.Spinner.Tick)
	}

	// Start periodic progress updates
	cmds = append(cmds, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return ProgressTickMsg{Time: t}
	}))

	return tea.Batch(cmds...)
}

// Update implements tea.Model
func (m *ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Update all running spinners
	for _, instruction := range m.instructions {
		if instruction.Status == StatusRunning {
			var cmd tea.Cmd
			instruction.Spinner, cmd = instruction.Spinner.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, tea.Batch(cmds...)

	case InstructionStartedMsg:
		s := bubbleteaSpinner.New()
		s.Spinner = bubbleteaSpinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		instruction := &InstructionState{
			ID:              msg.ID,
			Name:            msg.Name,
			Status:          StatusRunning,
			Progress:        0.3, // Start at 30% for visual feedback
			ProgressBar:     progress.New(progress.WithGradient("#008000", "#C0C0C0")),
			Spinner:         s,
			WarningMessages: make([]string, 0),
			InfoMessages:    make([]string, 0),
		}
		m.instructions[msg.ID] = instruction
		m.instructionOrder = append(m.instructionOrder, msg.ID)
		cmds = append(cmds, instruction.Spinner.Tick)
		return m, tea.Batch(cmds...)

	case InstructionProgressMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Progress = msg.Progress
			if msg.Message != "" { // what is message?
				instruction.InfoMessages = append(instruction.InfoMessages, msg.Message)
			}
		}
		return m, tea.Batch(cmds...)

	case InstructionCompletedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusCompleted
			instruction.Progress = 1.0
			instruction.Result = msg.Result
		} // should we handle case where instruction doesn't exist? this means that the instruction didn't have an instruction started message?
		return m, tea.Batch(cmds...)

	case InstructionFailedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusFailed
			instruction.ErrorMessage = msg.Error
		}
		return m, tea.Batch(cmds...)

	case InstructionWarningMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.WarningMessages = append(instruction.WarningMessages, msg.Warning)
		}
		return m, tea.Batch(cmds...)

	case InstructionInfoMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.InfoMessages = append(instruction.InfoMessages, msg.Info)
		}
		return m, tea.Batch(cmds...)

	case ExecutionCompleteMsg:
		if instruction, exists := m.instructions["execution"]; exists {
			instruction.Status = StatusCompleted
			instruction.Result = msg.Result
			instruction.Progress = 1.0
		}

		// m.done = true
		// m.error = msg.Error
		return m, tea.Batch(cmds...)

	case ProgressTickMsg:
		// Gradually increase progress for running instructions
		for _, instruction := range m.instructions {
			if instruction.Status == StatusRunning && instruction.Progress < 0.9 {
				// Slowly increase progress, max out at 90% until completion
				instruction.Progress += 0.02
				if instruction.Progress > 0.9 {
					instruction.Progress = 0.9
				}
			}
		}
		// Schedule next progress tick
		return m, tea.Batch(append(cmds, tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
			return ProgressTickMsg{Time: t}
		}))...)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *ExecutionModel) View() string {
	if !m.isInteractive {
		// For non-interactive terminals, fall back to simple output
		return m.renderNonInteractive()
	}

	return m.renderInteractive()
}

// renderInteractive renders the full TUI interface
func (m *ExecutionModel) renderInteractive() string {
	var content []string

	// Instructions with special ordering: execution always last
	orderedInstructions := m.getOrderedInstructions()
	for _, instruction := range orderedInstructions {
		content = append(content, m.renderInstruction(instruction))
	}

	// Summary if execution is complete
	if m.done {
		content = append(content, m.renderSummary())
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// renderNonInteractive renders simple output for non-interactive terminals
func (m *ExecutionModel) renderNonInteractive() string {
	return ""
}

// renderInstruction renders a single instruction
func (m *ExecutionModel) renderInstruction(instruction *InstructionState) string {
	var style lipgloss.Style
	var statusText string

	switch instruction.Status {
	case StatusPending:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		statusText = "⏳ " + instruction.Name
	case StatusRunning:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
		statusText = instruction.Spinner.View() + " " + instruction.Name
	case StatusCompleted:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
		statusText = "✅ " + instruction.Name
	case StatusFailed:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
		statusText = "❌ " + instruction.Name
	}

	// Build instruction line with status
	line := style.Render(statusText)

	// Add progress bar if running
	if instruction.Status == StatusRunning {
		progressDisplay := instruction.ProgressBar.ViewAs(instruction.Progress)
		line += "\n" + progressDisplay + "\n\n"
	}

	if instruction.Status == StatusFailed {
		line += "\n" + instruction.ErrorMessage
	}

	if instruction.Status == StatusCompleted {
		if instruction.Result != "" {
			line += "\n" + instruction.Result
		} else {
			line += "\n" + "Execution completed successfully"
		}
	}

	return line
}

// renderSummary renders the execution summary
func (m *ExecutionModel) renderSummary() string {
	completed := 0
	failed := 0

	for _, instruction := range m.instructions {
		switch instruction.Status {
		case StatusCompleted:
			completed++
		case StatusFailed:
			failed++
		}
	}

	var summaryStyle lipgloss.Style
	var message string

	if failed > 0 {
		summaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
		message = "❌ Execution failed"
	} else {
		summaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
		message = "✅ Execution completed successfully"
	}

	stats := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render(" (" + string(rune(completed)) + " completed, " + string(rune(failed)) + " failed)")

	return summaryStyle.Render(message) + stats
}

// getOrderedInstructions returns instructions in the proper display order:
// interpretation
// validation
// regular instructions
// execution (always last)
func (m *ExecutionModel) getOrderedInstructions() []*InstructionState {
	var orderedInstructions []*InstructionState
	var executionInstruction *InstructionState

	// First pass: collect all non-execution instructions in original order
	for _, id := range m.instructionOrder {
		instruction := m.instructions[id]
		if id == "execution" {
			// Save execution instruction for last
			executionInstruction = instruction
		} else {
			orderedInstructions = append(orderedInstructions, instruction)
		}
	}

	// Add execution instruction at the end if it exists
	if executionInstruction != nil {
		orderedInstructions = append(orderedInstructions, executionInstruction)
	}

	return orderedInstructions
}
