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
	live      []wc.Match
	finished  []wc.Match
	upcoming  []wc.Match
	updatedAt time.Time
	body      string
}

func (l *Live) SetSize(w, h int) {
	l.w = w
	l.h = h
	l.body = l.render(l.live, l.finished, l.upcoming)
}

// Load fetches data from the cache and rebuilds the rendered body.
func (l *Live) Load(db *data.Store) {
	l.live = db.LiveMatches()
	l.finished = db.FinishedMatches()
	l.upcoming = db.UpcomingMatches()
	l.updatedAt = db.LastUpdated("matches:live")
	l.body = l.render(l.live, l.finished, l.upcoming)
}

func (l *Live) View() string {
	return l.body
}

// liveWidths returns the name column width and venue max length for the
// live/FT rows given the content area width.
// FT row fixed overhead: "  FT  " + 2×(flag≈2 + space) + score(5) + spaces ≈ 22 chars.
// Names get priority; venue gets whatever remains (0 if none).
func liveWidths(contentW int) (nameW, venueW int) {
	if contentW <= 0 {
		contentW = 62
	}
	nameW = clamp((contentW-22)/2, 10, 22)
	venueW = contentW - 22 - 2*nameW
	if venueW < 6 {
		venueW = 0
	}
	return
}

// upcomingWidths returns name/venue widths for upcoming rows.
// Overhead: "  " + 2×(flag + space) + " vs  " + "  " + kickoff(17) + "  " ≈ 34 chars.
func upcomingWidths(contentW int) (nameW, venueW int) {
	if contentW <= 0 {
		contentW = 62
	}
	nameW = clamp((contentW-34)/2, 10, 22)
	venueW = contentW - 34 - 2*nameW
	if venueW < 6 {
		venueW = 0
	}
	return
}

func (l *Live) render(live, finished, upcoming []wc.Match) string {
	nameW, venueW := liveWidths(l.w)
	upNameW, upVenueW := upcomingWidths(l.w)

	var sb strings.Builder

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
			return mi > mj
		})

		for _, m := range live {
			sb.WriteString(renderLiveRow(m, nameW, venueW))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// ── FULL TIME ─────────────────────────────────────────────────────────
	if len(finished) > 0 {
		sb.WriteString(styles.DimText.Render("  FULL TIME"))
		sb.WriteString("\n\n")
		for _, m := range finished {
			sb.WriteString(renderFTRow(m, nameW, venueW))
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
			sb.WriteString(renderUpcomingRow(m, upNameW, upVenueW))
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

func renderLiveRow(m wc.Match, nameW, venueW int) string {
	minute := "   '"
	if m.Minute != nil {
		minute = fmt.Sprintf("%3d'", *m.Minute)
	}
	min := styles.LiveBadge.Render("●") + " " + styles.DimText.Render(minute)
	score := "--"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	teams := fmt.Sprintf("  %s %-*s %s  %s %-*s",
		m.HomeTeam.Flag, nameW, homeName,
		styles.Bold.Render(score),
		m.AwayTeam.Flag, nameW, awayName,
	)
	row := "  " + min + " " + teams
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

func renderFTRow(m wc.Match, nameW, venueW int) string {
	score := "– –"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	row := fmt.Sprintf("  %s  %s %-*s %s  %s %-*s",
		styles.DimText.Render("FT"),
		m.HomeTeam.Flag, nameW, homeName,
		styles.DimText.Render(score),
		m.AwayTeam.Flag, nameW, awayName,
	)
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

func renderUpcomingRow(m wc.Match, nameW, venueW int) string {
	kickoff := m.KickoffAt.In(cetLoc).Format("Mon 02 Jan  15:04")
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	row := fmt.Sprintf("  %s %-*s vs  %s %-*s  %s",
		m.HomeTeam.Flag, nameW, homeName,
		m.AwayTeam.Flag, nameW, awayName,
		styles.GoldText.Render(kickoff),
	)
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

func venueShort(venue string, maxLen int) string {
	r := []rune(venue)
	if len(r) <= maxLen {
		return venue
	}
	if maxLen <= 1 {
		return "…"
	}
	return string(r[:maxLen-1]) + "…"
}
