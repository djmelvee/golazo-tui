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
	minute     int
	injuryTime int
	priority   int // lower = earlier when minutes are equal
	icon       string
	label      string
	score      string // gold score string, empty for non-goal events
}

type timelineGoal struct {
	minute     int
	injuryTime int
	priority   int
	side       string // "home" or "away"
	name       string
	flag       string
	penalty    bool
	ownGoal    bool
	homeScore  int
	awayScore  int
}

// MatchDetail is the per-match detail screen (opened from the live dashboard).
type MatchDetail struct {
	w, h        int
	match       wc.Match
	events      []wc.GoalEvent
	allMatches  []wc.Match
	eventLog    []displayEvent // rebuilt on every render, used for scroll bounds
	blink       bool
	scroll      int
	body        string
}

func (d *MatchDetail) SetSize(w, h int) {
	d.w = w
	d.h = h
	d.body = d.render()
}

// Set loads a new match and resets scroll.
func (d *MatchDetail) Set(m wc.Match, events []wc.GoalEvent, all []wc.Match) {
	d.match = m
	d.events = events
	d.allMatches = all
	d.scroll = 0
	d.body = d.render()
}

// Update refreshes match data without resetting scroll (for live tick updates).
func (d *MatchDetail) Update(m wc.Match, events []wc.GoalEvent, all []wc.Match) {
	d.match = m
	d.events = events
	d.allMatches = all
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
		m.HomeTeam.Flag+"  "+styles.GoldBold.Render(m.HomeTeam.Name),
	)
	awayTeam := lipgloss.NewStyle().Width(half).Align(lipgloss.Right).Render(
		styles.GoldBold.Render(m.AwayTeam.Name)+"  "+m.AwayTeam.Flag,
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
	heading := "MATCH TIMELINE"
	if m.Status == wc.StatusUpcoming {
		heading = "MATCH PREVIEW"
	}
	sb.WriteString(styles.Heading.Render("  "+heading) + "\n\n")

	d.eventLog = d.buildEventLog()

	if len(d.eventLog) == 0 {
		sb.WriteString(styles.DimText.Render("  No timeline data yet") + "\n")
	} else {
		visible := d.eventLog
		if d.scroll > 0 && d.scroll < len(visible) {
			visible = visible[d.scroll:]
		}
		visLines := d.h - 18
		if visLines < 4 {
			visLines = 4
		}
		if len(visible) > visLines {
			visible = visible[:visLines]
		}

		for _, ev := range visible {
			minStr := formatEventMinute(ev.minute, ev.injuryTime)
			if ev.minute < 0 {
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

	if summary := d.matchSummary(); summary != "" {
		sb.WriteString("\n" + styles.DimText.Render("  "+summary) + "\n")
	}

	return sb.String()
}

func (d *MatchDetail) buildEventLog() []displayEvent {
	m := d.match
	var evs []displayEvent

	evs = append(evs, d.buildContextEvents()...)

	if m.Status == wc.StatusUpcoming {
		sortTimeline(evs)
		return evs
	}

	currentMinute := 0
	if m.Minute != nil {
		currentMinute = *m.Minute
	}

	evs = append(evs, displayEvent{
		minute: 0, priority: 0,
		icon:  "🟢",
		label: "Kick-off",
	})

	goals := d.gatherGoals()
	htHome, htAway := halftimeScore(goals)

	for i, g := range goals {
		prevH, prevA := 0, 0
		if i > 0 {
			prevH, prevA = goals[i-1].homeScore, goals[i-1].awayScore
		}
		tag := goalTags(g)
		comment := goalCommentary(m.HomeTeam.Name, m.AwayTeam.Name, prevH, prevA, g)
		name := g.name
		if name == "" {
			name = teamNameForSide(m, g.side)
		}
		label := g.flag + "  " + styles.GoldText.Render(name)
		if tag != "" {
			label += "  " + styles.DimText.Render(tag)
		}
		if comment != "" {
			label += "  " + styles.DimText.Render("·  "+comment)
		}
		evs = append(evs, displayEvent{
			minute: g.minute, injuryTime: g.injuryTime, priority: g.priority,
			icon:  goalIcon(g),
			label: label,
			score: fmt.Sprintf("%d – %d", g.homeScore, g.awayScore),
		})
	}

	if currentMinute >= 45 || m.Status == wc.StatusFinished {
		htLabel := styles.DimText.Render("── HT ──")
		if htHome+htAway > 0 {
			htLabel = styles.DimText.Render(fmt.Sprintf("── HT  %d – %d ──", htHome, htAway))
		}
		evs = append(evs, displayEvent{
			minute: 45, priority: 2,
			icon:  "╌╌",
			label: htLabel,
		})
	}

	if currentMinute >= 46 || m.Status == wc.StatusFinished {
		evs = append(evs, displayEvent{
			minute: 46, priority: 0,
			icon:  "🟢",
			label: "Second half underway",
		})
	}

	if shHome, shAway := secondHalfGoals(goals); shHome+shAway >= 3 && (currentMinute >= 46 || m.Status == wc.StatusFinished) {
		evs = append(evs, displayEvent{
			minute: 46, priority: 1,
			icon:  "🔥",
			label: styles.DimText.Render(fmt.Sprintf("Explosive second half — %d goals after the break", shHome+shAway)),
		})
	}

	if m.Status == wc.StatusFinished {
		ftLabel := "Full time"
		if ftNote := fullTimeNote(m, htHome, htAway); ftNote != "" {
			ftLabel = ftNote
		}
		evs = append(evs, displayEvent{
			minute: 90, priority: 2,
			icon:  "🏁",
			label: styles.GoldText.Render(ftLabel),
		})
	} else if currentMinute >= 80 {
		evs = append(evs, displayEvent{
			minute: currentMinute, priority: 3,
			icon:  "⏱",
			label: styles.DimText.Render("Into the closing stages…"),
		})
	}

	sortTimeline(evs)
	return evs
}

func (d *MatchDetail) buildContextEvents() []displayEvent {
	m := d.match
	forms := wc.BuildTeamForms(d.allMatches)
	var evs []displayEvent

	homeForm := lookupForm(forms, m.HomeTeam)
	awayForm := lookupForm(forms, m.AwayTeam)
	if homeForm != nil && homeForm.Played > 0 {
		evs = append(evs, displayEvent{
			minute: -3, priority: 0, icon: "📋",
			label: fmt.Sprintf("%s  %s  ·  %s",
				m.HomeTeam.Flag,
				styles.GoldText.Render(m.HomeTeam.Name),
				styles.DimText.Render(formSummary(homeForm)),
			),
		})
	}
	if awayForm != nil && awayForm.Played > 0 {
		evs = append(evs, displayEvent{
			minute: -2, priority: 0, icon: "📋",
			label: fmt.Sprintf("%s  %s  ·  %s",
				m.AwayTeam.Flag,
				styles.GoldText.Render(m.AwayTeam.Name),
				styles.DimText.Render(formSummary(awayForm)),
			),
		})
	}

	if stakes := matchStakes(m); stakes != "" {
		evs = append(evs, displayEvent{
			minute: -1, priority: 0, icon: "🏆",
			label: styles.DimText.Render(stakes),
		})
	}

	if h2h := headToHeadLine(m, d.allMatches); h2h != "" {
		evs = append(evs, displayEvent{
			minute: -1, priority: 1, icon: "↔",
			label: styles.DimText.Render(h2h),
		})
	}

	if m.Status == wc.StatusUpcoming {
		if homeForm != nil && awayForm != nil && homeForm.Played > 0 && awayForm.Played > 0 {
			fav := "Evenly matched on paper"
			switch {
			case homeForm.Strength > awayForm.Strength+0.12:
				fav = m.HomeTeam.Flag + " " + m.HomeTeam.Name + " arrive with the sharper edge"
			case awayForm.Strength > homeForm.Strength+0.12:
				fav = m.AwayTeam.Flag + " " + m.AwayTeam.Name + " arrive with the sharper edge"
			}
			evs = append(evs, displayEvent{
				minute: -1, priority: 2, icon: "📊",
				label: styles.DimText.Render(fav),
			})
		}
		if !m.KickoffAt.IsZero() {
			evs = append(evs, displayEvent{
				minute: -1, priority: 3, icon: "🕐",
				label: styles.DimText.Render("Kick-off " + tz.FormatKickoff(m.KickoffAt)),
			})
		}
	}

	return evs
}

func (d *MatchDetail) gatherGoals() []timelineGoal {
	m := d.match
	var goals []timelineGoal

	if len(m.HomeScorers)+len(m.AwayScorers) > 0 {
		for _, s := range m.HomeScorers {
			goals = append(goals, timelineGoal{
				minute: s.Minute, injuryTime: s.InjuryTime, priority: 1,
				side: "home", name: s.Name, flag: m.HomeTeam.Flag,
				penalty: s.Penalty, ownGoal: s.OwnGoal,
			})
		}
		for _, s := range m.AwayScorers {
			goals = append(goals, timelineGoal{
				minute: s.Minute, injuryTime: s.InjuryTime, priority: 1,
				side: "away", name: s.Name, flag: m.AwayTeam.Flag,
				penalty: s.Penalty, ownGoal: s.OwnGoal,
			})
		}
	} else {
		for _, ev := range d.events {
			side := ev.ScoredBy
			if side != "home" && side != "away" {
				side = "home"
			}
			flag := m.HomeTeam.Flag
			name := ev.ScorerName
			if side == "away" {
				flag = m.AwayTeam.Flag
			}
			goals = append(goals, timelineGoal{
				minute: ev.Minute, priority: 1,
				side: side, name: name, flag: flag,
				homeScore: ev.HomeScore, awayScore: ev.AwayScore,
			})
		}
	}

	sort.Slice(goals, func(i, j int) bool {
		if goals[i].minute != goals[j].minute {
			return goals[i].minute < goals[j].minute
		}
		if goals[i].injuryTime != goals[j].injuryTime {
			return goals[i].injuryTime < goals[j].injuryTime
		}
		return goals[i].priority < goals[j].priority
	})

	if len(m.HomeScorers)+len(m.AwayScorers) > 0 {
		h, a := 0, 0
		for i := range goals {
			if goals[i].side == "home" {
				h++
			} else {
				a++
			}
			goals[i].homeScore = h
			goals[i].awayScore = a
		}
	}

	return goals
}

func (d *MatchDetail) matchSummary() string {
	m := d.match
	if m.Status == wc.StatusUpcoming {
		return "Cards, subs & VAR not in current API feed"
	}
	if m.Status != wc.StatusFinished {
		return "Live timeline · cards/subs not in feed"
	}
	if m.HomeScore == nil || m.AwayScore == nil {
		return ""
	}
	total := *m.HomeScore + *m.AwayScore
	parts := []string{"Final score locked in"}
	if total >= 5 {
		parts = append(parts, "goal fest")
	} else if total == 0 {
		parts = append(parts, "nil-nil stalemate")
	}
	if *m.HomeScore == 0 && *m.AwayScore > 0 {
		parts = append(parts, m.AwayTeam.Name+" clean sheet")
	} else if *m.AwayScore == 0 && *m.HomeScore > 0 {
		parts = append(parts, m.HomeTeam.Name+" clean sheet")
	}
	return strings.Join(parts, " · ")
}

func sortTimeline(evs []displayEvent) {
	sort.Slice(evs, func(i, j int) bool {
		if evs[i].minute != evs[j].minute {
			return evs[i].minute < evs[j].minute
		}
		return evs[i].priority < evs[j].priority
	})
}

func formatEventMinute(minute, injury int) string {
	if injury > 0 {
		return fmt.Sprintf("%3d+%d'", minute, injury)
	}
	return fmt.Sprintf("%3d'", minute)
}

func goalIcon(g timelineGoal) string {
	if g.penalty {
		return "🎯"
	}
	if g.ownGoal {
		return "😬"
	}
	return "⚽"
}

func goalTags(g timelineGoal) string {
	var tags []string
	if g.ownGoal {
		tags = append(tags, "OG")
	}
	if g.penalty {
		tags = append(tags, "pen")
	}
	if len(tags) == 0 {
		return ""
	}
	return "(" + strings.Join(tags, ", ") + ")"
}

func goalCommentary(homeName, awayName string, prevH, prevA int, g timelineGoal) string {
	h, a := g.homeScore, g.awayScore
	totalBefore := prevH + prevA

	if totalBefore == 0 {
		return "Opens the scoring!"
	}
	if h == a && prevH != prevA {
		return "Levels it up!"
	}
	if g.side == "home" && h > a && prevH <= prevA {
		return homeName + " take the lead!"
	}
	if g.side == "away" && a > h && prevA <= prevH {
		return awayName + " take the lead!"
	}
	if g.injuryTime > 0 || g.minute > 90 {
		return "Drama in stoppage time!"
	}
	if g.minute >= 80 {
		return "Late drama!"
	}
	if g.minute <= 15 {
		return "Early breakthrough"
	}
	if g.penalty {
		return "From the spot!"
	}
	if g.ownGoal {
		return "Cruel deflection"
	}
	if totalBefore >= 3 && h+a >= totalBefore+1 {
		return "Another twist!"
	}
	return ""
}

func halftimeScore(goals []timelineGoal) (home, away int) {
	for _, g := range goals {
		if g.minute > 45 {
			continue
		}
		if g.side == "home" {
			home++
		} else {
			away++
		}
	}
	return home, away
}

func secondHalfGoals(goals []timelineGoal) (home, away int) {
	for _, g := range goals {
		if g.minute < 46 {
			continue
		}
		if g.side == "home" {
			home++
		} else {
			away++
		}
	}
	return home, away
}

func fullTimeNote(m wc.Match, htHome, htAway int) string {
	if m.HomeScore == nil || m.AwayScore == nil {
		return ""
	}
	hs, as := *m.HomeScore, *m.AwayScore
	total := hs + as

	var notes []string
	switch {
	case hs > as:
		notes = append(notes, m.HomeTeam.Name+" win")
	case as > hs:
		notes = append(notes, m.AwayTeam.Name+" win")
	default:
		notes = append(notes, "Honours even")
	}

	if total >= 5 {
		notes = append(notes, "what a shootout")
	}
	if hs > htHome && as > htAway && htHome == htAway && (hs != as) {
		notes = append(notes, "both sides scored after HT")
	}
	if hs < as && htHome > htAway {
		notes = append(notes, m.HomeTeam.Name+" comeback!")
	} else if as < hs && htAway > htHome {
		notes = append(notes, m.AwayTeam.Name+" comeback!")
	}
	if m.Stage != "" && m.Stage != "group" {
		winner := m.HomeTeam.Name
		if as > hs {
			winner = m.AwayTeam.Name
		} else if hs == as {
			winner = ""
		}
		if winner != "" {
			notes = append(notes, winner+" advance")
		}
	}
	return strings.Join(notes, " · ")
}

func formSummary(f *wc.TeamForm) string {
	return fmt.Sprintf("%dW-%dD-%dL · %.1f gpg · %.0f%% clean sheets",
		f.W, f.D, f.L, f.GoalsPerGame, f.CleanSheetRate*100)
}

func lookupForm(forms map[int]*wc.TeamForm, team wc.Team) *wc.TeamForm {
	if f, ok := forms[team.ID]; ok {
		return f
	}
	for _, f := range forms {
		if f.Team.Name == team.Name {
			return f
		}
	}
	return nil
}

func matchStakes(m wc.Match) string {
	switch m.Stage {
	case "final":
		return "The biggest prize in football — World Cup final"
	case "sf":
		return "Semi-final — a final berth on the line"
	case "qf":
		return "Quarter-final — knockout tension"
	case "r16":
		return "Round of 16 — win or go home"
	case "r32":
		return "Round of 32 — first knockout hurdle"
	case "third":
		return "Third-place play-off"
	case "group", "":
		if m.Group != "" {
			return fmt.Sprintf("Group %s · matchday %d — every point counts", m.Group, m.Matchday)
		}
	}
	return ""
}

func headToHeadLine(m wc.Match, all []wc.Match) string {
	var meetings []wc.Match
	for _, o := range all {
		if o.ID == m.ID || o.Status != wc.StatusFinished {
			continue
		}
		if (o.HomeTeam.Name == m.HomeTeam.Name && o.AwayTeam.Name == m.AwayTeam.Name) ||
			(o.HomeTeam.Name == m.AwayTeam.Name && o.AwayTeam.Name == m.HomeTeam.Name) {
			meetings = append(meetings, o)
		}
	}
	if len(meetings) == 0 {
		return ""
	}
	homeW, awayW, draws := 0, 0, 0
	for _, o := range meetings {
		if o.HomeScore == nil || o.AwayScore == nil {
			continue
		}
		hs, as := *o.HomeScore, *o.AwayScore
		homeIsHome := o.HomeTeam.Name == m.HomeTeam.Name
		switch {
		case hs == as:
			draws++
		case hs > as:
			if homeIsHome {
				homeW++
			} else {
				awayW++
			}
		default:
			if homeIsHome {
				awayW++
			} else {
				homeW++
			}
		}
	}
	return fmt.Sprintf("Tournament H2H: %s %dW · %d draws · %s %dW",
		m.HomeTeam.Flag, homeW, draws, m.AwayTeam.Flag, awayW)
}

func teamNameForSide(m wc.Match, side string) string {
	if side == "away" {
		return m.AwayTeam.Name
	}
	return m.HomeTeam.Name
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