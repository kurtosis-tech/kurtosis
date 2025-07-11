package output_printers

import (
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kurtosis-tech/kurtosis/cli/cli/command_args/run"
)

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
	Spinner         spinner.Model
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
	verbosity     run.Verbosity
	dryRun        bool
	isInteractive bool

	// Program control
	done  bool
	error error
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
	Success bool
	Error   error
}

// NewExecutionModel creates a new ExecutionModel
func NewExecutionModel(verbosity run.Verbosity, dryRun bool, isInteractive bool) *ExecutionModel {
	return &ExecutionModel{
		instructions:     make(map[string]*InstructionState),
		instructionOrder: make([]string, 0),
		verbosity:        verbosity,
		dryRun:           dryRun,
		isInteractive:    isInteractive,
		done:             false,
	}
}

// Init implements tea.Model
func (m *ExecutionModel) Init() tea.Cmd {
	return nil
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
		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		instruction := &InstructionState{
			ID:              msg.ID,
			Name:            msg.Name,
			Status:          StatusRunning,
			Progress:        0.25,
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
		m.done = true
		m.error = msg.Error
		return m, tea.Quit

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

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Render("Kurtosis Execution")
	content = append(content, header)

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
	// This will be implemented to provide fallback for non-interactive terminals
	// For now, return empty string as we'll handle this in the existing printer
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

	if instruction.Status == StatusCompleted {
		line += "\n" + instruction.Result
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
// interpretation, validation, regular instructions, execution (always last)
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
