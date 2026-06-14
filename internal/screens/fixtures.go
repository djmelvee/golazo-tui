package screens

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Fixtures shows upcoming matches grouped by date.
type Fixtures struct {
	w, h    int
	matches []wc.Match
}

func (f *Fixtures) SetSize(w, h int) {
	f.w = w
	f.h = h
}

func (f *Fixtures) Load(db *data.Store) {
	f.matches = db.UpcomingMatches()
}

func (f Fixtures) View() string {
	if len(f.matches) == 0 {
		return styles.DimText.Render("  No upcoming fixtures. Run golazo-seed first.\n")
	}

	// Sort by kickoff
	sorted := make([]wc.Match, len(f.matches))
	copy(sorted, f.matches)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].KickoffAt.Before(sorted[j].KickoffAt)
	})

	var sb strings.Builder
	sb.WriteString(styles.Heading.Render("  ─── UPCOMING FIXTURES  ·  GROUP STAGE"))
	sb.WriteString("\n\n")

	// Group by date
	type dateGroup struct {
		date    time.Time
		matches []wc.Match
	}
	var groups []dateGroup
	dateMap := make(map[string]int)

	for _, m := range sorted {
		dateKey := m.KickoffAt.Local().Format("2006-01-02")
		if idx, exists := dateMap[dateKey]; exists {
			groups[idx].matches = append(groups[idx].matches, m)
		} else {
			dateMap[dateKey] = len(groups)
			groups = append(groups, dateGroup{
				date:    m.KickoffAt.Local(),
				matches: []wc.Match{m},
			})
		}
	}

	for _, g := range groups {
		matchday := ""
		if g.matches[0].Matchday > 0 {
			matchday = fmt.Sprintf("  ·  MATCHDAY %d", g.matches[0].Matchday)
		}
		dateHeader := fmt.Sprintf("  %s%s", g.date.Format("Mon 02 Jan 2026"), matchday)
		sb.WriteString(styles.GoldBold.Render(dateHeader))
		sb.WriteString("\n")

		for _, m := range g.matches {
			sb.WriteString(renderFixtureRow(m))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func renderFixtureRow(m wc.Match) string {
	kickoff := m.KickoffAt.Local().Format("15:04")
	return fmt.Sprintf("  %s %-18s vs  %s %-18s  %s  %s",
		m.HomeTeam.Flag, m.HomeTeam.Name,
		m.AwayTeam.Flag, m.AwayTeam.Name,
		styles.GoldText.Render(kickoff),
		styles.DimText.Render(m.Venue),
	)
}
