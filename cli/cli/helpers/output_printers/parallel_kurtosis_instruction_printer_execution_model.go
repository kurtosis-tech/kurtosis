package output_printers

import (
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurtosis-tech/kurtosis/cli/cli/helpers/interactive_terminal_decider"
	"github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine"
)

var (
	progressBarGradient   = progress.WithGradient("#008000", "#C0C0C0")
	bubbleteaSpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))

	validationInstructionId     = startosis_engine.ValidationInstructionId
	interpretationInstructionId = startosis_engine.InterpretationInstructionId
	executionInstructionId      = startosis_engine.ExecutionInstructionId

	gray   = lipgloss.Color("8")
	yellow = lipgloss.Color("11")
	green  = lipgloss.Color("10")
	red    = lipgloss.Color("9")

	periodicTickInterval = 500 * time.Millisecond
)

const (
	ExecutionInstructionName      = "Executing Starlark code"
	ValidationInstructionName     = "Validating Starlark code"
	InterpretationInstructionName = "Interpreting Starlark code"

	ExecutionCompletedMessage = "Execution completed successfully"
	ExecutionFailedMessage    = "Execution failed"
)

type InstructionStatus int

const (
	StatusPending InstructionStatus = iota
	StatusRunning
	StatusCompleted
	StatusFailed
)

type InstructionState struct {
	ID              string
	Name            string
	Status          InstructionStatus
	Progress        float64
	ProgressBar     progress.Model
	Spinner         spinner.Model
	ErrorMessage    string
	Result          string
	WarningMessages []string
	InfoMessages    []string
}

// ExecutionModel is responsible for tracking state of the Terminal UI as StarlarkRunResponseLine objects are received during instruction execution
// The execution printer sends messages to the execution model as StarlarkRunResponseLine objects are received
// Init initializes the execution model, keeping track of the state of each instruction, and UI objects
// Update updates the state of execution model as new msg's updating instructions states are received
// View renders the UI based on the state of the execution model on some interval
type ExecutionModel struct {
	instructions     map[string]*InstructionState
	instructionOrder []string

	width, height int
	isInteractive bool

	done bool
}

type InstructionStartedMsg struct {
	ID   string
	Name string
}

type InstructionProgressMsg struct {
	ID       string
	Progress float64
	Message  string
}

type InstructionCompletedMsg struct {
	ID     string
	Result string
}

type InstructionFailedMsg struct {
	ID    string
	Error string
}

type InstructionWarningMsg struct {
	ID      string
	Warning string
}

type InstructionInfoMsg struct {
	ID   string
	Info string
}

type WindowSizeMsg struct {
	Width, Height int
}

type ExecutionCompleteMsg struct {
	ID      string
	Result  string
	Success bool
	Error   error
}

type ProgressTickMsg struct {
	Time time.Time
}

func NewExecutionModel() *ExecutionModel {
	instructions := make(map[string]*InstructionState)

	// Start with interpretation, execution, validation instructions (really these are phases but we treat them as instructions printing)
	instructions[executionInstructionId] = getInitialInstructionState(executionInstructionId, ExecutionInstructionName)
	instructions[validationInstructionId] = getInitialInstructionState(validationInstructionId, ValidationInstructionName)
	instructions[interpretationInstructionId] = getInitialInstructionState(interpretationInstructionId, InterpretationInstructionName)

	return &ExecutionModel{
		instructions:     instructions,
		instructionOrder: []string{executionInstructionId},
		done:             false,
		isInteractive:    interactive_terminal_decider.IsInteractiveTerminal(),
	}
}

func (m *ExecutionModel) Init() tea.Cmd {
	var cmds []tea.Cmd

	if execution, exists := m.instructions[executionInstructionId]; exists {
		cmds = append(cmds, execution.Spinner.Tick)
	}

	cmds = append(cmds, tea.Tick(periodicTickInterval, func(t time.Time) tea.Msg {
		return ProgressTickMsg{Time: t}
	}))

	return tea.Batch(cmds...)
}

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
		instruction := getInitialInstructionState(msg.ID, msg.Name)

		m.instructions[msg.ID] = instruction
		m.instructionOrder = append(m.instructionOrder, msg.ID)
		cmds = append(cmds, instruction.Spinner.Tick)
		return m, tea.Batch(cmds...)

	case InstructionProgressMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Progress = msg.Progress
			if msg.Message != "" {
				instruction.InfoMessages = append(instruction.InfoMessages, msg.Message)
			}
		}

		return m, tea.Batch(cmds...)
	case InstructionCompletedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusCompleted
			instruction.Progress = 1.0
			instruction.Result = msg.Result
		}

		return m, tea.Batch(cmds...)
	case InstructionFailedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusFailed
			instruction.ErrorMessage = msg.Error
		}

		m.done = true
		return m, tea.Quit
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
		if instruction, exists := m.instructions[executionInstructionId]; exists {
			instruction.Status = StatusCompleted
			instruction.Result = msg.Result
			instruction.Progress = 1.0
		}

		m.done = true
		return m, tea.Quit
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

		return m, tea.Batch(append(cmds, tea.Tick(periodicTickInterval, func(t time.Time) tea.Msg {
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

func (m *ExecutionModel) View() string {
	if !m.isInteractive {
		return m.renderNonInteractive()
	}

	return m.renderInteractive()
}

// renderInteractive renders the full TUI interface
func (m *ExecutionModel) renderInteractive() string {
	var content []string

	orderedInstructions := m.getOrderedInstructions()
	for _, instruction := range orderedInstructions {
		content = append(content, m.renderInstruction(instruction))
	}

	if m.done {
		content = append(content, m.renderSummary())
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

func (m *ExecutionModel) renderNonInteractive() string {
	return ""
}

// renderInstruction renders a single instruction
func (m *ExecutionModel) renderInstruction(instruction *InstructionState) string {
	var style lipgloss.Style
	var statusText string

	switch instruction.Status {
	case StatusPending:
		style = lipgloss.NewStyle().Foreground(gray) // Gray
		statusText = "⏳ " + instruction.Name
	case StatusRunning:
		style = lipgloss.NewStyle().Foreground(yellow) // Yellow
		statusText = instruction.Spinner.View() + " " + instruction.Name
	case StatusCompleted:
		style = lipgloss.NewStyle().Foreground(green) // Green
		statusText = "✅ " + instruction.Name
	case StatusFailed:
		style = lipgloss.NewStyle().Foreground(red) // Red
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
			line += "\n" + ExecutionCompletedMessage
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
		summaryStyle = lipgloss.NewStyle().Foreground(red).Bold(true)
		message = "❌ " + ExecutionFailedMessage
	} else {
		summaryStyle = lipgloss.NewStyle().Foreground(green).Bold(true)
		message = "✅ " + ExecutionCompletedMessage
	}

	stats := lipgloss.NewStyle().
		Foreground(gray).
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
		if id == executionInstructionId {
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

func getInitialInstructionState(id, name string) *InstructionState {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = bubbleteaSpinnerStyle

	return &InstructionState{
		ID:              id,
		Name:            name,
		Status:          StatusRunning,
		Progress:        0.1, // Start with 10% progress
		Result:          "",
		ProgressBar:     progress.New(progressBarGradient),
		Spinner:         s,
		WarningMessages: make([]string, 0),
		InfoMessages:    make([]string, 0),
		ErrorMessage:    "",
	}
}
