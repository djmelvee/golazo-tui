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

// Live is the live-match dashboard screen.
type Live struct {
	w, h      int
	body      string
	updatedAt time.Time
}

func (l *Live) SetSize(w, h int) {
	l.w = w
	l.h = h
}

// Load fetches data from the cache and rebuilds the rendered body.
func (l *Live) Load(db *data.Store) {
	live := db.LiveMatches()
	finished := db.FinishedMatches()
	upcoming := db.UpcomingMatches()
	l.updatedAt = db.LastUpdated("matches:live")
	l.body = l.render(live, finished, upcoming)
}

func (l *Live) View() string {
	return l.body
}

func (l *Live) render(live, finished, upcoming []wc.Match) string {
	var sb strings.Builder

	// Updated-at line
	if !l.updatedAt.IsZero() {
		sb.WriteString(styles.DimText.Render(
			fmt.Sprintf("  Updated %s CET  ·  auto-refreshes every 30s\n", l.updatedAt.In(cetLoc).Format("15:04")),
		))
	} else {
		sb.WriteString(styles.DimText.Render("  Loading match data...\n"))
	}
	sb.WriteString("\n")

	// ── LIVE ──────────────────────────────────────────────────────────────
	if len(live) > 0 {
		sb.WriteString(styles.LiveBadge.Render("  ● LIVE MATCHES"))
		sb.WriteString("\n\n")

		sort.Slice(live, func(i, j int) bool {
			mi, mj := 0, 0
			if live[i].Minute != nil {
				mi = *live[i].Minute
			}
			if live[j].Minute != nil {
				mj = *live[j].Minute
			}
			return mi > mj // highest minute first
		})

		for _, m := range live {
			sb.WriteString(renderLiveRow(m))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// ── FULL TIME ─────────────────────────────────────────────────────────
	if len(finished) > 0 {
		sb.WriteString(styles.DimText.Render("  FULL TIME"))
		sb.WriteString("\n\n")
		for _, m := range finished {
			sb.WriteString(renderFTRow(m))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// ── UPCOMING ──────────────────────────────────────────────────────────
	if len(upcoming) > 0 {
		sb.WriteString(styles.DimText.Render("  UPCOMING"))
		sb.WriteString("\n\n")

		sort.Slice(upcoming, func(i, j int) bool {
			return upcoming[i].KickoffAt.Before(upcoming[j].KickoffAt)
		})

		shown := upcoming
		if len(shown) > 6 {
			shown = shown[:6]
		}
		for _, m := range shown {
			sb.WriteString(renderUpcomingRow(m))
			sb.WriteString("\n")
		}
	}

	switch {
	case len(live)+len(finished)+len(upcoming) == 0:
		sb.WriteString(styles.DimText.Render("  No match data yet. Run golazo-fetcher or golazo-seed first.\n"))
	case len(live) == 0:
		sb.WriteString(styles.DimText.Render("  No matches currently live  ·  check back during match hours\n"))
	}

	return sb.String()
}

func renderLiveRow(m wc.Match) string {
	minute := "  '"
	if m.Minute != nil {
		minute = fmt.Sprintf("%3d'", *m.Minute)
	}
	min := styles.LiveBadge.Render("●") + " " + styles.DimText.Render(minute)
	score := "--"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	teams := fmt.Sprintf("  %s %-18s %s  %s %-18s",
		m.HomeTeam.Flag, m.HomeTeam.Name,
		styles.Bold.Render(score),
		m.AwayTeam.Flag, m.AwayTeam.Name,
	)
	venue := styles.DimText.Render("  " + venueShort(m.Venue))
	return "  " + min + " " + teams + venue
}

func renderFTRow(m wc.Match) string {
	score := "– –"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	return fmt.Sprintf("  %s  %s %-18s %s  %s %-18s  %s",
		styles.DimText.Render("FT"),
		m.HomeTeam.Flag, m.HomeTeam.Name,
		styles.DimText.Render(score),
		m.AwayTeam.Flag, m.AwayTeam.Name,
		styles.DimText.Render(venueShort(m.Venue)),
	)
}

func renderUpcomingRow(m wc.Match) string {
	kickoff := m.KickoffAt.In(cetLoc).Format("Mon 02 Jan  15:04")
	return fmt.Sprintf("  %s %-18s vs  %s %-18s  %s  %s",
		m.HomeTeam.Flag, m.HomeTeam.Name,
		m.AwayTeam.Flag, m.AwayTeam.Name,
		styles.GoldText.Render(kickoff),
		styles.DimText.Render(venueShort(m.Venue)),
	)
}

func venueShort(venue string) string {
	// Truncate to ~30 chars for the live row
	if len(venue) > 32 {
		return venue[:29] + "…"
	}
	return venue
}
