package tz

import (
	"sync"
	"time"
)

// DisplayMode controls which timezone user-facing labels use.
type DisplayMode string

const (
	DisplayAmsterdam DisplayMode = "amsterdam"
	DisplayLocal     DisplayMode = "local"
	DisplayUTC       DisplayMode = "utc"
)

var (
	displayMu   sync.RWMutex
	displayMode = DisplayAmsterdam
)

// SetDisplayMode sets the active display timezone (amsterdam, local, utc).
func SetDisplayMode(mode DisplayMode) {
	displayMu.Lock()
	displayMode = mode
	displayMu.Unlock()
}

// GetDisplayMode returns the active display timezone mode.
func GetDisplayMode() DisplayMode {
	displayMu.RLock()
	defer displayMu.RUnlock()
	return displayMode
}

func displayLoc() *time.Location {
	displayMu.RLock()
	mode := displayMode
	displayMu.RUnlock()
	switch mode {
	case DisplayLocal:
		return time.Local
	case DisplayUTC:
		return time.UTC
	default:
		return Amsterdam
	}
}

// DisplayIn converts t to the active display timezone.
func DisplayIn(t time.Time) time.Time {
	return t.In(displayLoc())
}

// StartOfDisplayDay returns midnight on t's calendar day in the display timezone.
func StartOfDisplayDay(t time.Time) time.Time {
	at := DisplayIn(t)
	loc := displayLoc()
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, loc)
}

// DisplayLabel returns a short label for the active display zone.
func DisplayLabel() string {
	switch GetDisplayMode() {
	case DisplayLocal:
		return "local"
	case DisplayUTC:
		return "UTC"
	default:
		return "Amsterdam"
	}
}