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

// Digest is a newspaper-style match-day summary.
type Digest struct {
	w, h       int
	live       []wc.Match
	finished   []wc.Match
	upcoming   []wc.Match
	recentGoals []data.RecentGoal
	scroll     int
	cursor     int
	lines      []string
	matchIDs   []int
	lineForID  map[int]int
}

func (d *Digest) SetSize(w, h int) {
	d.w = w
	d.h = h
	d.lines = d.buildLines()
}

// DigestDay holds matches grouped for the digest screen.
type DigestDay struct {
	Live, Finished, Upcoming []wc.Match
}

// ClassifyDigestDay groups cache buckets into today's digest in the display timezone.
func ClassifyDigestDay(now time.Time, liveIn, finishedIn, upcomingIn []wc.Match) DigestDay {
	today := tz.StartOfDisplayDay(now)
	tomorrow := today.Add(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	lateCutoff := yesterday.Add(18 * time.Hour)
	earlyMorning := tz.DisplayIn(now).Before(today.Add(6 * time.Hour))

	seen := make(map[int]struct{})
	var day DigestDay

	day.Live = append([]wc.Match(nil), liveIn...)
	for _, m := range day.Live {
		seen[m.ID] = struct{}{}
	}

	for _, m := range finishedIn {
		if m.KickoffAt.IsZero() {
			continue
		}
		k := tz.DisplayIn(m.KickoffAt)
		onToday := !k.Before(today) && k.Before(tomorrow)
		carryOver := earlyMorning && !k.Before(lateCutoff) && k.Before(today)
		if onToday || carryOver {
			day.Finished = append(day.Finished, m)
			seen[m.ID] = struct{}{}
		}
	}

	for _, m := range upcomingIn {
		if _, ok := seen[m.ID]; ok {
			continue
		}
		if m.KickoffAt.IsZero() {
			continue
		}
		k := tz.DisplayIn(m.KickoffAt)
		if !k.Before(today) && k.Before(tomorrow) {
			day.Upcoming = append(day.Upcoming, m)
		}
	}

	sort.Slice(day.Live, func(i, j int) bool {
		return tz.DisplayIn(day.Live[i].KickoffAt).Before(tz.DisplayIn(day.Live[j].KickoffAt))
	})
	sort.Slice(day.Finished, func(i, j int) bool {
		return tz.DisplayIn(day.Finished[i].KickoffAt).After(tz.DisplayIn(day.Finished[j].KickoffAt))
	})
	sort.Slice(day.Upcoming, func(i, j int) bool {
		return tz.DisplayIn(day.Upcoming[i].KickoffAt).Before(tz.DisplayIn(day.Upcoming[j].KickoffAt))
	})
	return day
}

func (d *Digest) Load(db *data.Store) {
	day := ClassifyDigestDay(time.Now(), db.LiveMatches(), db.FinishedMatches(), db.UpcomingMatches())
	d.live = day.Live
	d.finished = day.Finished
	d.upcoming = day.Upcoming
	d.recentGoals = db.RecentGoals(5)
	d.scroll = 0
	d.cursor = 0
	d.lines = d.buildLines()
}

func (d *Digest) allMatches() []wc.Match {
	var all []wc.Match
	all = append(all, d.live...)
	all = append(all, d.finished...)
	all = append(all, d.upcoming...)
	return all
}

func (d *Digest) CursorDown() {
	if d.cursor < len(d.matchIDs)-1 {
		d.cursor++
		d.ensureVisible()
	}
}

func (d *Digest) CursorUp() {
	if d.cursor > 0 {
		d.cursor--
		d.ensureVisible()
	}
}

func (d *Digest) ScrollDown() {
	max := len(d.lines) - (d.h - 8)
	if max < 0 {
		max = 0
	}
	if d.scroll < max {
		d.scroll++
	}
}

func (d *Digest) ScrollUp() {
	if d.scroll > 0 {
		d.scroll--
	}
}

func (d *Digest) ensureVisible() {
	if d.cursor < 0 || d.cursor >= len(d.matchIDs) {
		return
	}
	id := d.matchIDs[d.cursor]
	line, ok := d.lineForID[id]
	if !ok {
		return
	}
	vis := d.h - 10
	if vis < 4 {
		vis = 4
	}
	if line < d.scroll {
		d.scroll = line
	}
	if line >= d.scroll+vis {
		d.scroll = line - vis + 1
	}
}

func (d *Digest) SelectedMatchID() int {
	if d.cursor >= 0 && d.cursor < len(d.matchIDs) {
		return d.matchIDs[d.cursor]
	}
	return 0
}

func (d *Digest) View() string {
	vis := d.h - 8
	if vis < 4 {
		vis = 4
	}
	end := d.scroll + vis
	if end > len(d.lines) {
		end = len(d.lines)
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(d.lines[d.scroll:end], "\n"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  j/k pick · enter detail · "+tz.DisplayLabel()+" time") + "\n")
	return sb.String()
}

func (d *Digest) buildLines() []string {
	date := tz.FormatDigestDate(time.Now())
	total := len(d.live) + len(d.finished) + len(d.upcoming)
	rule := styles.DimText.Render("  " + strings.Repeat("═", clamp(d.w-4, 30, 70)))
	var lines []string
	lines = append(lines, styles.Heading.Render("  MATCH DAY DIGEST")+"\n")
	lines = append(lines, styles.GoldText.Render("  "+date)+"  "+styles.DimText.Render(fmt.Sprintf("(%d fixtures)", total))+"\n")

	lines = append(lines, d.headlineBlock()...)
	lines = append(lines, rule)

	d.matchIDs = nil
	d.lineForID = make(map[int]int)

	lines = append(lines, d.section("LIVE NOW", d.live, true, false)...)
	lines = append(lines, rule)
	lines = append(lines, d.section("FULL TIME TODAY", d.finished, false, false)...)
	lines = append(lines, rule)
	lines = append(lines, d.section("STILL TO PLAY", d.upcoming, false, true)...)

	if len(d.recentGoals) > 0 {
		lines = append(lines, rule)
		lines = append(lines, "")
		lines = append(lines, styles.GoldBold.Render("  RECENT GOALS"))
		for _, rg := range d.recentGoals {
			g := rg.Goal
			who := rg.HomeTeam
			flag := rg.HomeFlag
			if g.ScoredBy == "away" {
				who = rg.AwayTeam
				flag = rg.AwayFlag
			}
			name := g.ScorerName
			if name == "" {
				name = who
			}
			lines = append(lines, fmt.Sprintf("  ⚽ %s %s  %d–%d  %s %s (%d')",
				flag, name, g.HomeScore, g.AwayScore, rg.HomeFlag, rg.AwayFlag, g.Minute))
		}
	}

	if total == 0 {
		lines = append(lines, "")
		lines = append(lines, styles.DimText.Render("  No World Cup fixtures on today's calendar.")+"\n")
	}
	return lines
}

func (d *Digest) headlineBlock() []string {
	var lines []string
	lines = append(lines, "")
	if len(d.live) > 0 {
		m := d.live[0]
		score := "0–0"
		if m.HomeScore != nil && m.AwayScore != nil {
			score = fmt.Sprintf("%d–%d", *m.HomeScore, *m.AwayScore)
		}
		lines = append(lines, styles.GoldBold.Render(fmt.Sprintf(
			"  TOP STORY: %s %s vs %s %s (%s)",
			m.HomeTeam.Flag, truncate(m.HomeTeam.Name, 12), m.AwayTeam.Flag, truncate(m.AwayTeam.Name, 12), score)))
	}
	if upset := d.biggestUpset(); upset != nil {
		m := *upset
		lines = append(lines, styles.MainText.Render(fmt.Sprintf(
			"  Biggest result today: %s %d–%d %s",
			m.HomeTeam.Flag, *m.HomeScore, *m.AwayScore, m.AwayTeam.Flag)))
	}
	if len(d.upcoming) > 0 {
		nx := d.upcoming[0]
		lines = append(lines, styles.DimText.Render(fmt.Sprintf(
			"  Next up: %s vs %s at %s",
			nx.HomeTeam.Name, nx.AwayTeam.Name, tz.FormatClock(nx.KickoffAt))))
	}
	return lines
}

func (d *Digest) biggestUpset() *wc.Match {
	var best *wc.Match
	for i := range d.finished {
		m := d.finished[i]
		if m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		margin := absInt(*m.HomeScore - *m.AwayScore)
		if margin >= 2 {
			if best == nil || margin > absInt(*best.HomeScore-*best.AwayScore) {
				cp := m
				best = &cp
			}
		}
	}
	return best
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (d *Digest) section(title string, matches []wc.Match, live, showKick bool) []string {
	var lines []string
	lines = append(lines, "")
	lines = append(lines, styles.GoldBold.Render("  "+title))
	if len(matches) == 0 {
		msg := "  — none —"
		if live {
			msg = "  — nothing live right now —"
		}
		lines = append(lines, styles.DimText.Render(msg))
		return lines
	}
	for i, m := range matches {
		d.matchIDs = append(d.matchIDs, m.ID)
		cursorHere := len(d.matchIDs) - 1 == d.cursor
		lineIdx := len(lines)
		d.lineForID[m.ID] = lineIdx
		lines = append(lines, d.matchRow(m, live, showKick, cursorHere))
		_ = i
	}
	return lines
}

func (d *Digest) matchRow(m wc.Match, live, showKick, selected bool) string {
	score := "– –"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	if live {
		score = styles.GoldBold.Render(score)
	} else {
		score = styles.MainText.Render(score)
	}

	prefix := "  "
	if selected {
		prefix = styles.GoldText.Render("> ")
	} else if live {
		prefix = "  " + styles.LiveBadge.Render("●") + " "
	}

	meta := ""
	if live && m.Minute != nil {
		meta = styles.LiveBadge.Render(fmt.Sprintf(" %d'", *m.Minute))
	} else if showKick && !m.KickoffAt.IsZero() {
		meta = styles.GoldText.Render(" " + tz.FormatClock(m.KickoffAt))
	} else if !m.KickoffAt.IsZero() {
		meta = styles.DimText.Render(" " + tz.FormatClock(m.KickoffAt))
	}

	stage := ""
	if m.Stage != "" && m.Stage != "group" {
		stage = styles.DimText.Render("  " + wc.StageLabel(m.Stage))
	}

	return fmt.Sprintf("%s%s %s  %s  %s %s%s%s",
		prefix,
		m.HomeTeam.Flag, truncate(m.HomeTeam.Name, 14),
		score,
		m.AwayTeam.Flag, truncate(m.AwayTeam.Name, 14),
		meta, stage,
	)
}