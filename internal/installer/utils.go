package installer

import tea "github.com/charmbracelet/bubbletea"

// sendDetailedOutputMsg sends detailed output messages to the program if verbose mode is enabled
func sendDetailedOutputMsg(p *tea.Program, msg string, verbose bool) {
	if !verbose || len(msg) == 0 {
		return
	}
	p.Send(msg)
}
