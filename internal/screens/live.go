package screens

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Live is the live-match dashboard screen.
type Live struct {
	w, h       int
	live       []wc.Match
	finished   []wc.Match
	upcoming   []wc.Match
	updatedAt  time.Time
	blink      bool
	cursor     int // selected row: live, FT, then upcoming
	fetchNote  string
	favorite   string
	body       string
}

// SetFetchNote sets the status note shown next to the refresh timestamp.
func (l *Live) SetFetchNote(note string) {
	l.fetchNote = note
}

func (l *Live) SetSize(w, h int) {
	l.w = w
	l.h = h
	l.body = l.render(l.live, l.finished, l.upcoming)
}

// Load fetches data from the cache and rebuilds the rendered body.
// Live matches come only from the API live bucket — never inferred from kickoff time.
func (l *Live) Load(db *data.Store) {
	l.live = db.LiveMatches()
	l.finished = db.FinishedMatches()
	l.upcoming = db.FutureMatches()
	l.updatedAt = db.LastUpdated("matches:live")
	l.favorite = FavoriteName(db)
	total := l.selectableCount()
	if total == 0 {
		l.cursor = 0
	} else if l.cursor >= total {
		l.cursor = total - 1
	}
	l.body = l.render(l.live, l.finished, l.upcoming)
}

func (l *Live) selectableCount() int {
	up := len(l.upcoming)
	if up > 12 {
		up = 12
	}
	return len(l.live) + len(l.finished) + up
}

func (l *Live) View() string {
	return l.body
}

// ToggleBlink flips the blink state and re-renders only when live matches are present.
func (l *Live) ToggleBlink() {
	l.blink = !l.blink
	if len(l.live) > 0 {
		l.body = l.render(l.live, l.finished, l.upcoming)
	}
}

// LiveCount returns how many matches are currently live.
func (l *Live) LiveCount() int { return len(l.live) }

// Blink returns the current blink state.
func (l *Live) Blink() bool { return l.blink }

// CursorUp moves the selection up in the live match list.
func (l *Live) CursorUp() {
	if l.cursor > 0 {
		l.cursor--
		l.body = l.render(l.live, l.finished, l.upcoming)
	}
}

// CursorDown moves selection down through live, FT, and upcoming rows.
func (l *Live) CursorDown() {
	if l.cursor < l.selectableCount()-1 {
		l.cursor++
		l.body = l.render(l.live, l.finished, l.upcoming)
	}
}

// SelectedMatch returns the highlighted match, or nil if none.
func (l *Live) SelectedMatch() *wc.Match {
	all := l.orderedSelectable()
	if len(all) == 0 {
		return nil
	}
	idx := l.cursor
	if idx >= len(all) {
		idx = len(all) - 1
	}
	m := all[idx]
	return &m
}

func (l *Live) orderedSelectable() []wc.Match {
	live := append([]wc.Match{}, l.live...)
	finished := append([]wc.Match{}, l.finished...)
	upcoming := append([]wc.Match{}, l.upcoming...)
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
	sort.Slice(finished, func(i, j int) bool {
		return finished[i].KickoffAt.After(finished[j].KickoffAt)
	})
	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].KickoffAt.Before(upcoming[j].KickoffAt)
	})
	if len(upcoming) > 12 {
		upcoming = upcoming[:12]
	}
	var all []wc.Match
	all = append(all, live...)
	all = append(all, finished...)
	all = append(all, upcoming...)
	return all
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
		line := fmt.Sprintf("  Updated %s (%s)  ·  data polls every 5s · UI tick 1s", tz.FormatClock(l.updatedAt), tz.DisplayLabel())
		if l.fetchNote != "" {
			line += "  ·  " + l.fetchNote
		}
		sb.WriteString(styles.DimText.Render(line + "\n"))
	} else {
		line := "  Loading match data..."
		if l.fetchNote != "" {
			line += "  ·  " + l.fetchNote
		}
		sb.WriteString(styles.DimText.Render(line + "\n"))
	}
	sb.WriteString("\n")

	// ── LIVE ──────────────────────────────────────────────────────────────
	if len(live) > 0 {
		dot := styles.DimText.Render("●")
		if l.blink {
			dot = styles.LiveBadge.Render("●")
		}
		sb.WriteString("  " + dot + styles.LiveBadge.Render(" LIVE MATCHES"))
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

		for i, m := range live {
			sb.WriteString(renderLiveRow(m, nameW, venueW, l.blink, i == l.cursor, l.favorite))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// ── FULL TIME ─────────────────────────────────────────────────────────
	if len(finished) > 0 {
		sb.WriteString(styles.DimText.Render("  FULL TIME"))
		sb.WriteString("\n\n")
		sort.Slice(finished, func(i, j int) bool {
			return finished[i].KickoffAt.After(finished[j].KickoffAt)
		})
		for i, m := range finished {
			sb.WriteString(renderFTRow(m, nameW, venueW, (len(live)+i) == l.cursor, l.favorite))
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
		if len(shown) > 12 {
			shown = shown[:12]
		}
		base := len(live) + len(finished)
		for i, m := range shown {
			sb.WriteString(renderUpcomingRow(m, upNameW, upVenueW, (base+i) == l.cursor, l.favorite))
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

func renderLiveRow(m wc.Match, nameW, venueW int, blink, selected bool, favorite string) string {
	minute := "  --"
	if m.Minute != nil {
		minute = fmt.Sprintf("%3d'", *m.Minute)
	}
	dot := styles.DimText.Render("●")
	if blink {
		dot = styles.LiveBadge.Render("●")
	}
	min := dot + " " + styles.DimText.Render(minute)
	score := "--"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	teams := fmt.Sprintf("  %s %-*s %s  %s %-*s",
		m.HomeTeam.Flag, nameW, homeName,
		styles.GoldBold.Render(score),
		m.AwayTeam.Flag, nameW, awayName,
	)
	prefix := "  "
	if selected {
		prefix = styles.GoldText.Render("> ")
	}
	row := prefix + min + " " + teams
	if favorite != "" && (m.HomeTeam.Name == favorite || m.AwayTeam.Name == favorite) {
		row += " " + styles.GoldText.Render("★")
	}
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

func renderFTRow(m wc.Match, nameW, venueW int, selected bool, favorite string) string {
	score := "– –"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	prefix := "  "
	if selected {
		prefix = styles.GoldText.Render("> ")
	}
	row := prefix + fmt.Sprintf("%s  %s %-*s %s  %s %-*s",
		styles.DimText.Render("FT"),
		m.HomeTeam.Flag, nameW, homeName,
		styles.DimText.Render(score),
		m.AwayTeam.Flag, nameW, awayName,
	)
	if favorite != "" && (m.HomeTeam.Name == favorite || m.AwayTeam.Name == favorite) {
		row += " " + styles.GoldText.Render("★")
	}
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

func renderUpcomingRow(m wc.Match, nameW, venueW int, selected bool, favorite string) string {
	kickoff := tz.FormatKickoffShort(m.KickoffAt)
	homeName := truncate(m.HomeTeam.Name, nameW)
	awayName := truncate(m.AwayTeam.Name, nameW)
	prefix := "  "
	if selected {
		prefix = styles.GoldText.Render("> ")
	}
	row := fmt.Sprintf("%s%s %-*s vs  %s %-*s  %s",
		prefix, m.HomeTeam.Flag, nameW, homeName,
		m.AwayTeam.Flag, nameW, awayName,
		styles.GoldText.Render(kickoff),
	)
	if favorite != "" && (m.HomeTeam.Name == favorite || m.AwayTeam.Name == favorite) {
		row += " " + styles.GoldText.Render("★")
	}
	if venueW > 0 {
		row += "  " + styles.DimText.Render(venueShort(m.Venue, venueW))
	}
	return row
}

// FindMatch returns the promoted match by ID (live, finished, or upcoming).
func (l *Live) FindMatch(id int) *wc.Match {
	for i := range l.live {
		if l.live[i].ID == id {
			m := l.live[i]
			return &m
		}
	}
	for i := range l.finished {
		if l.finished[i].ID == id {
			m := l.finished[i]
			return &m
		}
	}
	for i := range l.upcoming {
		if l.upcoming[i].ID == id {
			m := l.upcoming[i]
			return &m
		}
	}
	return nil
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
