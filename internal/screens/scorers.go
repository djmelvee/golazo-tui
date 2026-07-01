package screens

import (
	"fmt"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Scorers shows the Golden Boot leaderboard.
type Scorers struct {
	w, h     int
	rows     []wc.ScorerRow
	scroll   int
	cursor   int
}

func (s *Scorers) SetSize(w, h int) {
	s.w = w
	s.h = h
}

func (s *Scorers) Load(db *data.Store) {
	s.rows = wc.BuildTopScorers(db.AllMatches())
	s.scroll = 0
	s.cursor = 0
}

func (s *Scorers) CursorDown() {
	if s.cursor < len(s.rows)-1 {
		s.cursor++
		s.ensureVisible()
	}
}

func (s *Scorers) CursorUp() {
	if s.cursor > 0 {
		s.cursor--
		s.ensureVisible()
	}
}

func (s *Scorers) ensureVisible() {
	vis := s.h - 12
	if vis < 5 {
		vis = 5
	}
	if s.cursor < s.scroll {
		s.scroll = s.cursor
	}
	if s.cursor >= s.scroll+vis {
		s.scroll = s.cursor - vis + 1
	}
}

func (s *Scorers) SelectedTeam() string {
	if s.cursor >= 0 && s.cursor < len(s.rows) {
		return s.rows[s.cursor].Team
	}
	return ""
}

func (s *Scorers) ScrollDown() { s.CursorDown() }
func (s *Scorers) ScrollUp()   { s.CursorUp() }

func (s *Scorers) View() string {
	if len(s.rows) == 0 {
		return styles.DimText.Render("  No scorer data yet (finished matches with API scorers).\n")
	}

	nameW := clamp((s.w-28)/2, 12, 22)
	var sb strings.Builder
	sb.WriteString(styles.Heading.Render("  GOLDEN BOOT") + "\n\n")
	sb.WriteString(styles.DimText.Render(fmt.Sprintf("  %-4s  %-*s  %-14s  %s\n", "#", nameW, "Player", "Team", "Goals")))
	sb.WriteString(styles.DimText.Render("  "+strings.Repeat("─", clamp(s.w-4, 30, 80))) + "\n")

	vis := s.h - 12
	if vis < 5 {
		vis = 5
	}
	end := s.scroll + vis
	if end > len(s.rows) {
		end = len(s.rows)
	}
	for i := s.scroll; i < end; i++ {
		r := s.rows[i]
		rank := fmt.Sprintf("%-4d", i+1)
		name := truncate(r.Name, nameW)
		team := truncate(r.Team, 14)
		goals := fmt.Sprintf("%d", r.Goals)
		prefix := "  "
		if i == s.cursor {
			prefix = styles.GoldText.Render("> ")
		}
		line := fmt.Sprintf("%s%s  %-*s  %s %-14s  %s", prefix, rank, nameW, name, r.Flag, team, goals)
		if i < 3 {
			sb.WriteString(styles.GoldBold.Render(line) + "\n")
		} else {
			sb.WriteString(line + "\n")
		}
	}
	sb.WriteString(styles.DimText.Render("\n  j/k pick · enter team hub") + "\n")
	return sb.String()
}