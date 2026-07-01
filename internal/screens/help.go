package screens

import (
	"strings"

	"github.com/djmelvee/golazo-tui/internal/styles"
)

// Help is the keybinding reference overlay.
type Help struct {
	w, h int
}

func (h *Help) SetSize(w, height int) {
	h.w = w
	h.h = height
}

func (Help) View() string {
	lines := []string{
		styles.Heading.Render("  KEYBOARD SHORTCUTS"),
		"",
		styles.GoldText.Render("  Navigation"),
		"  h live        p predictions   b bracket",
		"  d digest      s scorers       g groups",
		"  f fixtures    c changelog     t team",
		"  ? help        q quit",
		"",
		styles.GoldText.Render("  Lists"),
		"  j/k or ↑↓     move cursor or scroll (screen-dependent)",
		"  enter         open match or prediction detail",
		"  esc           back from detail screens",
		"",
		styles.GoldText.Render("  Live & goals"),
		"  !             toggle goal bell + desktop alerts",
		"  Golazo!       10s celebration overlay on any screen",
		"",
		styles.GoldText.Render("  Fixtures filter"),
		"  1 all         2 group stage   3 knockout",
		"",
		styles.GoldText.Render("  Preferences"),
		"  z             cycle timezone (Amsterdam / local / UTC)",
		"  GOLAZO_NO_SPLASH=1  skip boot animation",
		"",
		styles.DimText.Render("  b opens bracket from main screens; esc backs out of detail"),
	}
	return strings.Join(lines, "\n") + "\n"
}