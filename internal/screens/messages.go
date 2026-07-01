package screens

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

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

// TickMsg is sent every second to refresh the live screen.
type TickMsg time.Time

// TickCmd returns a command that fires a TickMsg every second.
func TickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// BlinkMsg drives the live-indicator animation (fires every second).
type BlinkMsg time.Time

// BlinkCmd returns a command that fires a BlinkMsg every second.
func BlinkCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return BlinkMsg(t)
	})
}

// SplashTickMsg advances the boot animation frame.
type SplashTickMsg time.Time

// SplashTickCmd fires every 150ms during splash.
func SplashTickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return SplashTickMsg(t)
	})
}

// SplashDoneMsg ends the splash screen.
type SplashDoneMsg struct{}

// SplashDoneCmd fires after the splash duration.
func SplashDoneCmd() tea.Cmd {
	return tea.Tick(3500*time.Millisecond, func(t time.Time) tea.Msg {
		return SplashDoneMsg{}
	})
}

// CelebrationDoneMsg clears the goal celebration overlay.
type CelebrationDoneMsg struct{}

// CelebrationDuration is how long the goal celebration overlay stays visible.
const CelebrationDuration = 10 * time.Second

// CelebrationDoneCmd clears celebration after CelebrationDuration.
func CelebrationDoneCmd() tea.Cmd {
	return tea.Tick(CelebrationDuration, func(t time.Time) tea.Msg {
		return CelebrationDoneMsg{}
	})
}
