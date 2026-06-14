package screens

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

// TickMsg is sent every 30 seconds to refresh the live screen.
type TickMsg time.Time

// TickCmd returns a command that fires a TickMsg after 30 seconds.
func TickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
