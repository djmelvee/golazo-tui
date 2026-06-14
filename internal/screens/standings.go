package screens

import (
	"fmt"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Standings shows all 12 groups A–L with W/D/L/GF/GA/GD/Pts.
type Standings struct {
	w, h   int
	groups map[string][]wc.GroupRow
	scroll int
	lines  []string // pre-rendered lines for scroll
}

func (s *Standings) SetSize(w, h int) {
	s.w = w
	s.h = h
	s.lines = s.renderLines() // re-render at new width
}

func (s *Standings) Load(db *data.Store) {
	s.groups = db.Standings()
	s.scroll = 0
	s.lines = s.renderLines()
}

func (s *Standings) ScrollDown() {
	max := len(s.lines) - (s.h - 6)
	if max < 0 {
		max = 0
	}
	if s.scroll < max {
		s.scroll++
	}
}

func (s *Standings) ScrollUp() {
	if s.scroll > 0 {
		s.scroll--
	}
}

func (s Standings) View() string {
	if len(s.lines) == 0 {
		return styles.DimText.Render("  No standings data. Run golazo-seed first.\n")
	}
	visible := s.h - 6
	if visible < 1 {
		visible = 20
	}

	start := s.scroll
	end := start + visible
	if end > len(s.lines) {
		end = len(s.lines)
	}

	var sb strings.Builder
	for _, line := range s.lines[start:end] {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if len(s.lines) > visible {
		sb.WriteString("\n")
		sb.WriteString(styles.DimText.Render(
			fmt.Sprintf("  j/k or ↑↓ to scroll  (%d/%d)", s.scroll+1, len(s.lines)),
		))
	}
	return sb.String()
}

var groupOrder = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L"}

// standingsNameWidth computes the team name column width from the content area
// width. Flag emoji (~2 visual cols) + space + name + 8 stat columns ≈ 41 chars
// overhead. Names grow up to 22 chars (longest: "Bosnia and Herzegovina").
func standingsNameWidth(contentW int) int {
	if contentW <= 0 {
		contentW = 62 // sensible default before first WindowSizeMsg
	}
	// overhead: 2 indent + 2 flag + 1 space + 8 stat cols (36) = 41
	return clamp((contentW-41)/1, 10, 22)
}

func (s *Standings) renderLines() []string {
	nameW := standingsNameWidth(s.w)
	var lines []string

	for _, grp := range groupOrder {
		rows, ok := s.groups[grp]
		if !ok {
			continue
		}

		header := fmt.Sprintf("  ─── GROUP %s  ·  FIFA WORLD CUP 2026", grp)
		lines = append(lines, styles.Heading.Render(header))
		lines = append(lines, "")

		// Column header: "Team" spans flag(≈2vis) + space + nameW; +2 compensates
		// for flag byte-width vs visual-width so the header aligns with data rows.
		colHdr := fmt.Sprintf("  %-*s %3s %3s %3s %3s %4s %4s %4s %4s",
			nameW+2, "Team", "P", "W", "D", "L", "GF", "GA", "GD", "Pts")
		lines = append(lines, styles.DimText.Render(colHdr))

		for i, row := range rows {
			lines = append(lines, renderGroupRow(row, i < 2, nameW))
		}

		anyPlayed := false
		for _, row := range rows {
			if row.Played > 0 {
				anyPlayed = true
				break
			}
		}
		if anyPlayed {
			lines = append(lines, styles.DimText.Render("  ✓ Top 2 advance to Round of 32"))
		}
		lines = append(lines, "")
	}

	return lines
}

func renderGroupRow(row wc.GroupRow, advancing bool, nameW int) string {
	gd := fmt.Sprintf("%+d", row.GD)
	if row.GD == 0 {
		gd = "0"
	}
	name := truncate(row.Team.Name, nameW)
	line := fmt.Sprintf("  %s %-*s %3d %3d %3d %3d %4d %4d %4s %4d",
		row.Team.Flag, nameW, name,
		row.Played, row.W, row.D, row.L,
		row.GF, row.GA, gd, row.Pts,
	)
	if advancing {
		return styles.Advancing.Render(line)
	}
	return styles.MainText.Render(line)
}
