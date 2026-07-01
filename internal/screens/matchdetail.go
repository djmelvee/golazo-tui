package screens

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// displayEvent is a single liveblog entry — goal, lifecycle marker, etc.
type displayEvent struct {
	minute   int
	priority int // lower = earlier when minutes are equal
	icon     string
	label    string
	score    string // gold score string, empty for non-goal events
}

// MatchDetail is the per-match detail screen (opened from the live dashboard).
type MatchDetail struct {
	w, h     int
	match    wc.Match
	events   []wc.GoalEvent
	eventLog []displayEvent // rebuilt on every render, used for scroll bounds
	blink    bool
	scroll   int
	body     string
}

func (d *MatchDetail) SetSize(w, h int) {
	d.w = w
	d.h = h
	d.body = d.render()
}

// Set loads a new match and resets scroll.
func (d *MatchDetail) Set(m wc.Match, events []wc.GoalEvent) {
	d.match = m
	d.events = events
	d.scroll = 0
	d.body = d.render()
}

// Update refreshes match data without resetting scroll (for live tick updates).
func (d *MatchDetail) Update(m wc.Match, events []wc.GoalEvent) {
	d.match = m
	d.events = events
	d.body = d.render()
}

func (d *MatchDetail) MatchID() int { return d.match.ID }
func (d *MatchDetail) View() string { return d.body }

func (d *MatchDetail) ToggleBlink() {
	d.blink = !d.blink
	d.body = d.render()
}

func (d *MatchDetail) ScrollDown() {
	if d.scroll < len(d.eventLog)-1 {
		d.scroll++
		d.body = d.render()
	}
}

func (d *MatchDetail) ScrollUp() {
	if d.scroll > 0 {
		d.scroll--
		d.body = d.render()
	}
}

func (d *MatchDetail) render() string {
	m := d.match
	contentW := d.w
	if contentW < 48 {
		contentW = 48
	}

	rule := styles.DimText.Render(strings.Repeat("─", contentW))

	var sb strings.Builder
	sb.WriteString("\n")

	// ── TEAMS ─────────────────────────────────────────────────────────────────
	half := contentW / 2
	homeTeam := lipgloss.NewStyle().Width(half).Align(lipgloss.Left).Render(
		m.HomeTeam.Flag + "  " + styles.GoldBold.Render(m.HomeTeam.Name),
	)
	awayTeam := lipgloss.NewStyle().Width(half).Align(lipgloss.Right).Render(
		styles.GoldBold.Render(m.AwayTeam.Name) + "  " + m.AwayTeam.Flag,
	)
	sb.WriteString("  " + homeTeam + awayTeam + "\n")
	sb.WriteString("  " + rule + "\n")

	// ── SCORE + MINUTE ────────────────────────────────────────────────────────
	homeStr, awayStr := "--", "--"
	if m.HomeScore != nil {
		homeStr = fmt.Sprintf("%d", *m.HomeScore)
	}
	if m.AwayScore != nil {
		awayStr = fmt.Sprintf("%d", *m.AwayScore)
	}

	dot := styles.DimText.Render("●")
	if d.blink && m.Status == wc.StatusLive {
		dot = styles.LiveBadge.Render("●")
	}
	var minuteStr string
	switch {
	case m.Minute != nil:
		minuteStr = dot + "  " + styles.DimText.Render(fmt.Sprintf("%d'", *m.Minute))
	case m.Status == wc.StatusLive:
		minuteStr = dot + "  " + styles.DimText.Render("--")
	case m.Status == wc.StatusFinished:
		minuteStr = styles.DimText.Render("FT")
	}

	scoreContent := fmt.Sprintf("  %s    %s    %s  ",
		styles.GoldBold.Render(homeStr),
		minuteStr,
		styles.GoldBold.Render(awayStr),
	)
	scoreLine := lipgloss.NewStyle().Width(contentW).Align(lipgloss.Center).Render(scoreContent)
	sb.WriteString("\n  " + scoreLine + "\n\n")
	sb.WriteString("  " + rule + "\n")

	// ── MATCH INFO ────────────────────────────────────────────────────────────
	stageLabel := stageString(m.Stage, m.Group, m.Matchday)
	info := m.Venue
	if !m.KickoffAt.IsZero() {
		kick := tz.FormatKickoff(m.KickoffAt)
		if info != "" {
			info += "  ·  "
		}
		info += kick
	}
	if stageLabel != "" {
		if info != "" {
			info += "  ·  "
		}
		info += stageLabel
	}
	sb.WriteString(styles.DimText.Render("  "+info) + "\n\n")

	// ── EVENTS ────────────────────────────────────────────────────────────────
	sb.WriteString(styles.Heading.Render("  MATCH EVENTS") + "\n\n")

	d.eventLog = d.buildEventLog()

	if len(d.eventLog) == 0 {
		sb.WriteString(styles.DimText.Render("  Match has not started yet") + "\n")
	} else {
		visible := d.eventLog
		if d.scroll > 0 && d.scroll < len(visible) {
			visible = visible[d.scroll:]
		}
		visLines := d.h - 16
		if visLines < 3 {
			visLines = 3
		}
		if len(visible) > visLines {
			visible = visible[:visLines]
		}

		for _, ev := range visible {
			minStr := fmt.Sprintf("%3d'", ev.minute)
			if ev.minute == 0 {
				minStr = "    "
			}
			line := fmt.Sprintf("  %s  %s  %s",
				ev.icon,
				styles.DimText.Render(minStr),
				ev.label,
			)
			if ev.score != "" {
				line += "  " + styles.GoldBold.Render(ev.score)
			}
			sb.WriteString(line + "\n")
		}


		if len(d.eventLog) > visLines || d.scroll > 0 {
			sb.WriteString(styles.DimText.Render(
				fmt.Sprintf("\n  %d / %d  ·  j/k scroll", d.scroll+len(visible), len(d.eventLog)),
			) + "\n")
		}
	}

	// ── PROGRESS BAR ─────────────────────────────────────────────────────────
	if m.Minute != nil && m.Status == wc.StatusLive {
		sb.WriteString("\n")
		barW := contentW - 14
		if barW < 10 {
			barW = 10
		}
		min := *m.Minute
		cap := 90
		if min > cap {
			cap = min
		}
		filled := (min * barW) / cap
		if filled > barW {
			filled = barW
		}
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
		sb.WriteString(fmt.Sprintf("  [%s] %d' / 90'\n",
			styles.GoldText.Render(bar), *m.Minute,
		))
	}

	if m.HomeScore != nil && m.AwayScore != nil && (*m.HomeScore+*m.AwayScore) > 0 {
		sb.WriteString("\n")
		barW := contentW - 20
		if barW < 10 {
			barW = 10
		}
		total := *m.HomeScore + *m.AwayScore
		homeW := (*m.HomeScore * barW) / total
		awayW := barW - homeW
		homeBar := strings.Repeat("█", homeW)
		awayBar := strings.Repeat("█", awayW)
		sb.WriteString(styles.DimText.Render("  Goals  ") +
			styles.GoldText.Render(homeBar) +
			styles.DimText.Render("│") +
			styles.GoldBold.Render(awayBar) +
			fmt.Sprintf("  %d – %d\n", *m.HomeScore, *m.AwayScore))
	}

	sb.WriteString(styles.DimText.Render("\n  Lineups / cards / VAR — not available from current API feed.\n"))

	return sb.String()
}

// buildEventLog merges goal events with derived lifecycle markers into a
// chronological liveblog list. Lifecycle events (kick-off, half-time, full
// time) are synthesised from match state; goals come from the fetcher cache.
func (d *MatchDetail) buildEventLog() []displayEvent {
	m := d.match
	if m.Status == wc.StatusUpcoming {
		return nil
	}

	var evs []displayEvent

	currentMinute := 0
	if m.Minute != nil {
		currentMinute = *m.Minute
	}

	// Kick-off
	evs = append(evs, displayEvent{
		minute: 0, priority: 0,
		icon:  "🟢",
		label: "Kick-off",
	})

	// Prefer API scorer timeline when available.
	if len(m.HomeScorers)+len(m.AwayScorers) > 0 {
		for _, s := range m.HomeScorers {
			evs = append(evs, displayEvent{
				minute: s.Minute, priority: 1, icon: "⚽",
				label: m.HomeTeam.Flag + "  " + styles.GoldText.Render(s.Name),
			})
		}
		for _, s := range m.AwayScorers {
			evs = append(evs, displayEvent{
				minute: s.Minute, priority: 1, icon: "⚽",
				label: m.AwayTeam.Flag + "  " + styles.GoldText.Render(s.Name),
			})
		}
	} else {
		for _, ev := range d.events {
			team := m.HomeTeam
			label := team.Flag + "  " + styles.GoldText.Render(team.Name)
			if ev.ScoredBy == "away" {
				team = m.AwayTeam
				label = team.Flag + "  " + styles.GoldText.Render(team.Name)
			}
			if ev.ScorerName != "" {
				label = team.Flag + "  " + styles.GoldText.Render(ev.ScorerName)
			}
			evs = append(evs, displayEvent{
				minute: ev.Minute, priority: 1,
				icon:  "⚽",
				label: label,
				score: fmt.Sprintf("%d – %d", ev.HomeScore, ev.AwayScore),
			})
		}
	}

	if currentMinute >= 45 || m.Status == wc.StatusFinished {
		evs = append(evs, displayEvent{
			minute: 45, priority: 2,
			icon:  "╌╌",
			label: styles.DimText.Render("── HT ──"),
		})
	}

	// Full time
	if m.Status == wc.StatusFinished {
		evs = append(evs, displayEvent{
			minute: 90, priority: 2,
			icon:  "🏁",
			label: styles.DimText.Render("Full time"),
		})
	}

	sort.Slice(evs, func(i, j int) bool {
		if evs[i].minute != evs[j].minute {
			return evs[i].minute < evs[j].minute
		}
		return evs[i].priority < evs[j].priority
	})

	return evs
}

func stageString(stage, group string, matchday int) string {
	switch stage {
	case "group", "":
		if group != "" {
			return fmt.Sprintf("Group %s  ·  Matchday %d", group, matchday)
		}
	case "r32":
		return "Round of 32"
	case "r16":
		return "Round of 16"
	case "qf":
		return "Quarter-final"
	case "sf":
		return "Semi-final"
	case "final":
		return "Final"
	}
	return stage
}
