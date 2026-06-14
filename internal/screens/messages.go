package screens

import (
	_ "time/tzdata"
	"time"

	tea "charm.land/bubbletea/v2"
)

// cetLoc is the Europe/Amsterdam timezone (CET/CEST). All match times in the
// TUI are displayed in this zone since the user is based in the Netherlands.
// Falls back to UTC if the timezone database is unavailable.
var cetLoc = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// clamp constrains v to [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// truncate shortens s to at most n runes, adding "…" if cut.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

// TickMsg is sent every 30 seconds to refresh the live screen.
type TickMsg time.Time

// TickCmd returns a command that fires a TickMsg after 30 seconds.
func TickCmd() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
