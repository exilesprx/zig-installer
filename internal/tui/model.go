package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/exilesprx/zig-install/internal/config"
	"github.com/exilesprx/zig-install/internal/logger"
)

// Model represents the state of our Bubble Tea app
type Model struct {
	spinner      spinner.Model
	status       string
	err          error
	installing   bool
	zigDone      bool
	zlsDone      bool
	quit         bool
	config       *config.Config
	styles       *Styles
	detailOutput string         // Stores detailed command outputs
	logger       logger.ILogger // Logger for logging errors
}

// Custom message types for our app
type (
	StatusMsg          string
	ErrorMsg           error
	InstallCompleteMsg string
	ZigDoneMsg         struct{}
	ZLSDoneMsg         struct{}
	DetailOutputMsg    string // For showing command outputs
)

// NewModel creates a new TUI model
func NewModel(config *config.Config, styles *Styles, logger logger.ILogger) Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = styles.Spinner

	return Model{
		spinner:    s,
		status:     "Starting installation...",
		installing: true,
		config:     config,
		styles:     styles,
		logger:     logger,
	}
}

// Init initializes the Bubble Tea model
func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages and user input
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			m.quit = true
			return m, tea.Quit
		}
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case InstallCompleteMsg:
		m.installing = false
		m.status = string(msg)
		m.quit = true
		return m, tea.Quit
	case ErrorMsg:
		m.err = msg
		m.installing = false
		m.quit = true
		return m, tea.Quit
	case StatusMsg:
		m.status = string(msg)
		return m, nil
	case DetailOutputMsg:
		if m.config.Verbose {
			m.detailOutput += string(msg) + "\n"
		}
		return m, nil
	case ZigDoneMsg:
		m.zigDone = true
		return m, nil
	case ZLSDoneMsg:
		m.zlsDone = true
		return m, nil
	}
	return m, nil
}

// View renders the current UI
func (m Model) View() string {
	if m.config.NoColor {
		return m.plainView()
	}
	return m.colorView()
}

// plainView renders the UI without colors
func (m Model) plainView() string {
	if m.quit {
		if m.err != nil {
			return fmt.Sprintf("Error: %v\n", m.err)
		}
		return fmt.Sprintf("%s\n", m.status)
	}

	var view string
	view += " Zig & ZLS Installer \n\n"

	if m.installing {
		view += fmt.Sprintf("%s %s\n", m.spinner.View(), m.status)
	} else {
		view += m.status + "\n"
	}

	if m.config.Verbose && m.detailOutput != "" {
		view += "\nDetails:\n"
		view += m.detailOutput + "\n"
	}

	if m.zigDone {
		view += "✓ Zig installed successfully\n"
	}
	if m.zlsDone {
		view += "✓ ZLS installed successfully\n"
	}

	view += "\nPress q to quit\n"
	return view
}

// colorView renders the UI with colors
func (m Model) colorView() string {
	docStyle := m.styles.Document

	titleBar := m.styles.Title.Render(" ✨ Zig & ZLS Installer ✨ ")
	separator := m.styles.Separator.Render(strings.Repeat("─", 40))

	if m.quit {
		if m.err != nil {
			return docStyle.Render(
				titleBar + "\n\n" +
					m.styles.Error.Render(fmt.Sprintf("Error: %v", m.err)) + "\n")
		}
		return docStyle.Render(
			titleBar + "\n\n" +
				m.styles.Success.Render(m.status) + "\n")
	}

	var statusDisplay string
	if m.installing {
		statusDisplay = fmt.Sprintf("%s %s", m.spinner.View(), m.styles.Status.Render(m.status))
	} else {
		statusDisplay = m.styles.Status.Render(m.status)
	}

	var completionInfo string
	if m.zigDone {
		completionInfo += m.styles.Success.Render("✓ Zig installed successfully") + "\n"
	}
	if m.zlsDone {
		completionInfo += m.styles.Success.Render("✓ ZLS installed successfully") + "\n"
	}

	// Add detailed output if verbose mode is enabled
	var detailSection string
	if m.config.Verbose && m.detailOutput != "" {
		detailStyle := m.styles.Detail

		detailSection = m.styles.Subtitle.Render("Details:") + "\n" +
			detailStyle.Render(m.detailOutput) + "\n"
	}

	footerText := m.styles.Footer.Render("\nPress q to quit")

	return docStyle.Render(
		titleBar + "\n\n" +
			statusDisplay + "\n\n" +
			completionInfo +
			detailSection +
			separator + "\n" +
			footerText)
}
