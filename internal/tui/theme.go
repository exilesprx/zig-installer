package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors contains the Catppuccin Mocha theme colors
type Colors struct {
	Rosewater lipgloss.Color
	Flamingo  lipgloss.Color
	Pink      lipgloss.Color
	Mauve     lipgloss.Color
	Red       lipgloss.Color
	Maroon    lipgloss.Color
	Peach     lipgloss.Color
	Yellow    lipgloss.Color
	Green     lipgloss.Color
	Teal      lipgloss.Color
	Sky       lipgloss.Color
	Sapphire  lipgloss.Color
	Blue      lipgloss.Color
	Lavender  lipgloss.Color
	Text      lipgloss.Color
	Subtext1  lipgloss.Color
	Subtext0  lipgloss.Color
	Overlay2  lipgloss.Color
	Overlay1  lipgloss.Color
	Overlay0  lipgloss.Color
	Surface2  lipgloss.Color
	Surface1  lipgloss.Color
	Surface0  lipgloss.Color
	Base      lipgloss.Color
	Mantle    lipgloss.Color
	Crust     lipgloss.Color
}

// Styles contains lipgloss styles used in the application
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Success  lipgloss.Style
	Error    lipgloss.Style
	Info     lipgloss.Style
	Header   lipgloss.Style
	Status   lipgloss.Style
}

// NewMochaColors creates a new theme colors instance with Catppuccin Mocha palette
func NewMochaColors() *Colors {
	return &Colors{
		Rosewater: lipgloss.Color("#f5e0dc"),
		Flamingo:  lipgloss.Color("#f2cdcd"),
		Pink:      lipgloss.Color("#f5c2e7"),
		Mauve:     lipgloss.Color("#cba6f7"),
		Red:       lipgloss.Color("#f38ba8"),
		Maroon:    lipgloss.Color("#eba0ac"),
		Peach:     lipgloss.Color("#fab387"),
		Yellow:    lipgloss.Color("#f9e2af"),
		Green:     lipgloss.Color("#a6e3a1"),
		Teal:      lipgloss.Color("#94e2d5"),
		Sky:       lipgloss.Color("#89dceb"),
		Sapphire:  lipgloss.Color("#74c7ec"),
		Blue:      lipgloss.Color("#89b4fa"),
		Lavender:  lipgloss.Color("#b4befe"),
		Text:      lipgloss.Color("#cdd6f4"),
		Subtext1:  lipgloss.Color("#bac2de"),
		Subtext0:  lipgloss.Color("#a6adc8"),
		Overlay2:  lipgloss.Color("#9399b2"),
		Overlay1:  lipgloss.Color("#7f849c"),
		Overlay0:  lipgloss.Color("#6c7086"),
		Surface2:  lipgloss.Color("#585b70"),
		Surface1:  lipgloss.Color("#45475a"),
		Surface0:  lipgloss.Color("#313244"),
		Base:      lipgloss.Color("#1e1e2e"),
		Mantle:    lipgloss.Color("#181825"),
		Crust:     lipgloss.Color("#11111b"),
	}
}

// NewStyles creates styles using theme colors
func NewStyles(colors *Colors) *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(colors.Text).
			Padding(0, 1),
		Subtitle: lipgloss.NewStyle().
			Foreground(colors.Mauve).
			Bold(true),
		Success: lipgloss.NewStyle().
			Foreground(colors.Green),
		Error: lipgloss.NewStyle().
			Foreground(colors.Red),
		Info: lipgloss.NewStyle().
			Foreground(colors.Blue),
		Header: lipgloss.NewStyle().
			Foreground(colors.Lavender).
			Bold(true),
		Status: lipgloss.NewStyle().
			Foreground(colors.Peach),
	}
}

// PrintWithStyles formats a message with styled output if colors are enabled
func PrintWithStyles(message string, style lipgloss.Style, noColor bool) string {
	if noColor {
		return message
	}
	return style.Render(message)
}
