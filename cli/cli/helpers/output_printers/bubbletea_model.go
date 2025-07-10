package output_printers

import (
	"time"

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
	StartTime       time.Time
	EndTime         *time.Time
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
	program *tea.Program
	done    bool
	error   error
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
		dryRun:          dryRun,
		isInteractive:   isInteractive,
		done:            false,
	}
}

// Init implements tea.Model
func (m *ExecutionModel) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update implements tea.Model
func (m *ExecutionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case InstructionStartedMsg:
		instruction := &InstructionState{
			ID:              msg.ID,
			Name:            msg.Name,
			Status:          StatusRunning,
			Progress:        0.0,
			StartTime:       time.Now(),
			WarningMessages: make([]string, 0),
			InfoMessages:    make([]string, 0),
		}
		m.instructions[msg.ID] = instruction
		m.instructionOrder = append(m.instructionOrder, msg.ID)
		return m, nil

	case InstructionProgressMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Progress = msg.Progress
			if msg.Message != "" {
				instruction.InfoMessages = append(instruction.InfoMessages, msg.Message)
			}
		}
		return m, nil

	case InstructionCompletedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusCompleted
			instruction.Progress = 1.0
			now := time.Now()
			instruction.EndTime = &now
			instruction.Result = msg.Result
		}
		return m, nil

	case InstructionFailedMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.Status = StatusFailed
			now := time.Now()
			instruction.EndTime = &now
			instruction.ErrorMessage = msg.Error
		}
		return m, nil

	case InstructionWarningMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.WarningMessages = append(instruction.WarningMessages, msg.Warning)
		}
		return m, nil

	case InstructionInfoMsg:
		if instruction, exists := m.instructions[msg.ID]; exists {
			instruction.InfoMessages = append(instruction.InfoMessages, msg.Info)
		}
		return m, nil

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

	return m, nil
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

	// Instructions
	for _, id := range m.instructionOrder {
		instruction := m.instructions[id]
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
	var statusIcon string

	switch instruction.Status {
	case StatusPending:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("8")) // Gray
		statusIcon = "⏳"
	case StatusRunning:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // Yellow
		statusIcon = "🔄"
	case StatusCompleted:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // Green
		statusIcon = "✅"
	case StatusFailed:
		style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // Red
		statusIcon = "❌"
	}

	// Build instruction line
	line := style.Render(statusIcon + " " + instruction.Name)

	// Add progress if running
	if instruction.Status == StatusRunning && instruction.Progress > 0 {
		progressBar := m.renderProgressBar(instruction.Progress)
		line += " " + progressBar
	}

	// Add timing if completed
	if instruction.EndTime != nil {
		duration := instruction.EndTime.Sub(instruction.StartTime)
		timing := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Render(" (" + duration.Round(time.Millisecond).String() + ")")
		line += timing
	}

	return line
}

// renderProgressBar renders a progress bar for running instructions
func (m *ExecutionModel) renderProgressBar(progress float64) string {
	const barWidth = 20
	filled := int(progress * barWidth)
	
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"

	percentage := int(progress * 100)
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("12")).
		Render(bar + " " + string(rune(percentage)) + "%")
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