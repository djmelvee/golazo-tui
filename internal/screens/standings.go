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
	phase  string // "group" or "knockout"
	scroll int
	cursor int
	teams  []wc.Team
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
	s.phase = wc.TournamentPhase(db.AllMatches())
	s.lines = s.renderLines()
}

func (s *Standings) CursorDown() {
	if s.cursor < len(s.teams)-1 {
		s.cursor++
		s.lines = s.renderLines()
		s.ensureVisible()
	}
}

func (s *Standings) CursorUp() {
	if s.cursor > 0 {
		s.cursor--
		s.lines = s.renderLines()
		s.ensureVisible()
	}
}

func (s *Standings) SelectedTeam() *wc.Team {
	if s.cursor >= 0 && s.cursor < len(s.teams) {
		t := s.teams[s.cursor]
		return &t
	}
	return nil
}

func (s *Standings) ensureVisible() {
	// rough scroll: ~1 line per team row after headers
	row := 6 + s.cursor
	vis := s.h - 8
	if vis < 4 {
		vis = 4
	}
	if row < s.scroll {
		s.scroll = row
	}
	if row >= s.scroll+vis {
		s.scroll = row - vis + 1
	}
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
	s.teams = nil
	nameW := standingsNameWidth(s.w)
	var lines []string

	lines = append(lines, styles.Heading.Render("  FINAL GROUP STANDINGS")+"\n")
	if s.phase == "knockout" {
		lines = append(lines, styles.GoldText.Render("  Group stage complete — knockout phase underway")+"\n")
		lines = append(lines, styles.DimText.Render("  Historical tables only · see Bracket [b] or Predictions [p]")+"\n")
	} else {
		lines = append(lines, styles.DimText.Render("  Group stage in progress")+"\n")
	}
	lines = append(lines, "")

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
			s.teams = append(s.teams, row.Team)
			lines = append(lines, renderGroupRow(row, i < 2, nameW, len(s.teams)-1 == s.cursor))
		}

		anyPlayed := false
		for _, row := range rows {
			if row.Played > 0 {
				anyPlayed = true
				break
			}
		}
		if anyPlayed {
			if s.phase == "knockout" {
				lines = append(lines, styles.DimText.Render("  ✓ Top 2 advanced to Round of 32"))
			} else {
				lines = append(lines, styles.DimText.Render("  ✓ Top 2 advance to Round of 32"))
			}
		}
		lines = append(lines, "")
	}

	return lines
}

func renderGroupRow(row wc.GroupRow, advancing bool, nameW int, selected bool) string {
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
	if selected {
		return styles.GoldText.Render("> " + line[2:])
	}
	if advancing {
		return styles.Advancing.Render(line)
	}
	return styles.MainText.Render(line)
}
