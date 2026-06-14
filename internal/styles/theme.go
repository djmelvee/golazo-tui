package styles

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette — FIFA World Cup 2026 brand colours
var (
	Gold       = lipgloss.Color("#f5c518")
	PitchGreen = lipgloss.Color("#1a472a")
	LiveRed    = lipgloss.Color("#e63946")
	TextMain   = lipgloss.Color("#e6edf3")
	TextDim    = lipgloss.Color("#6e7681")
	SidebarBg  = lipgloss.Color("#0d1117")
)

// Named styles — everything imports from here; no inline colours elsewhere.
var (
	AppHeader = lipgloss.NewStyle().
			Background(PitchGreen).
			Foreground(Gold).
			Bold(true).
			Padding(0, 2)

	GoldText = lipgloss.NewStyle().
			Foreground(Gold)

	GoldBold = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	DimText = lipgloss.NewStyle().
		Foreground(TextDim)

	MainText = lipgloss.NewStyle().
			Foreground(TextMain)

	LiveBadge = lipgloss.NewStyle().
			Foreground(LiveRed).
			Bold(true)

	ActiveNav = lipgloss.NewStyle().
			Foreground(Gold).
			Bold(true)

	InactiveNav = lipgloss.NewStyle().
			Foreground(TextDim)

	Heading = lipgloss.NewStyle().
		Foreground(Gold).
		Bold(true)

	Advancing = lipgloss.NewStyle().
			Foreground(Gold)

	Bold = lipgloss.NewStyle().
		Bold(true).
		Foreground(TextMain)

	Sidebar = lipgloss.NewStyle().
		Foreground(TextDim).
		Padding(0, 1)
)

// Render helpers — wraps a string with the given colour and no side-effects on
// surrounding layout.

func Gold_(s string) string    { return GoldText.Render(s) }
func Dim(s string) string      { return DimText.Render(s) }
func Live(s string) string     { return LiveBadge.Render(s) }
func Green(s string) string    { return lipgloss.NewStyle().Foreground(PitchGreen).Render(s) }
func BoldWhite(s string) string { return Bold.Render(s) }

// PaddedWidth returns a style set to the given width.
func PaddedWidth(w int) lipgloss.Style {
	return lipgloss.NewStyle().Width(w)
}

// HeaderBar renders the full-width WC2026 app header at the given width.
func HeaderBar(w int) string {
	content := "⚽  FIFA WORLD CUP 2026  ·  🇺🇸 USA  🇨🇦 CANADA  🇲🇽 MEXICO"
	return AppHeader.
		Width(w).
		Render(content)
}

// Ensure color.Color is used (avoid import cycle with image/color)
var _ color.Color = Gold
