package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Fixtures shows upcoming matches grouped by date with scroll and selection.
type Fixtures struct {
	w, h      int
	matches   []wc.Match
	phase     string
	filter    int // 0=all 1=group 2=knockout
	cursor    int
	scroll    int
	lines     []string
	matchIdx  []int
}

func (f *Fixtures) SetSize(w, h int) {
	f.w = w
	f.h = h
	f.rebuild()
}

func (f *Fixtures) Load(db *data.Store) {
	f.matches = db.FutureMatches()
	f.phase = wc.TournamentPhase(db.AllMatches())
	f.rebuild()
}

func (f *Fixtures) SetFilter(mode int) {
	f.filter = mode % 3
	f.cursor = 0
	f.scroll = 0
	f.rebuild()
}

func (f *Fixtures) filtered() []wc.Match {
	var out []wc.Match
	for _, m := range f.matches {
		switch f.filter {
		case 1:
			if m.Stage != "" && m.Stage != "group" {
				continue
			}
		case 2:
			if m.Stage == "" || m.Stage == "group" {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func (f *Fixtures) CursorDown() {
	if f.cursor < len(f.matchIdx)-1 {
		f.cursor++
		f.ensureVisible()
	}
}

func (f *Fixtures) CursorUp() {
	if f.cursor > 0 {
		f.cursor--
		f.ensureVisible()
	}
}

func (f *Fixtures) ScrollDown() {
	max := len(f.lines) - (f.h - 8)
	if max < 0 {
		max = 0
	}
	if f.scroll < max {
		f.scroll++
	}
}

func (f *Fixtures) ScrollUp() {
	if f.scroll > 0 {
		f.scroll--
	}
}

func (f *Fixtures) ensureVisible() {
	vis := f.h - 10
	if vis < 4 {
		vis = 4
	}
	starts := make([]int, len(f.matchIdx))
	heights := make([]int, len(f.matchIdx))
	for i, lineIdx := range f.matchIdx {
		starts[i] = lineIdx
		heights[i] = 1
	}
	f.scroll = ScrollToItem(f.cursor, starts, heights, vis, f.scroll)
}

func (f *Fixtures) SelectedMatch() *wc.Match {
	if f.cursor < 0 || f.cursor >= len(f.matchIdx) {
		return nil
	}
	line := f.matchIdx[f.cursor]
	for i, idx := range f.matchIdx {
		if i == f.cursor && idx == line {
			sorted := f.sortedFiltered()
			if f.cursor < len(sorted) {
				m := sorted[f.cursor]
				return &m
			}
		}
	}
	sorted := f.sortedFiltered()
	if f.cursor < len(sorted) {
		m := sorted[f.cursor]
		return &m
	}
	return nil
}

func (f *Fixtures) sortedFiltered() []wc.Match {
	sorted := make([]wc.Match, len(f.filtered()))
	copy(sorted, f.filtered())
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].KickoffAt.Before(sorted[j].KickoffAt)
	})
	return sorted
}

func (f *Fixtures) rebuild() {
	nameW := clamp((f.w-34)/2, 10, 22)
	venueW := f.w - 34 - 2*nameW
	if venueW < 6 {
		venueW = 0
	}

	sorted := f.sortedFiltered()
	var lines []string
	filterLabel := "all"
	switch f.filter {
	case 1:
		filterLabel = "group"
	case 2:
		filterLabel = "knockout"
	}
	stageLabel := "KNOCKOUT STAGE"
	if f.phase != "knockout" {
		stageLabel = "GROUP STAGE"
	}
	lines = append(lines, styles.Heading.Render("  UPCOMING FIXTURES  ·  "+stageLabel))
	lines = append(lines, styles.DimText.Render(fmt.Sprintf("  filter: %s · 1/2/3 switch · %s time", filterLabel, tz.DisplayLabel())))
	lines = append(lines, "")

	f.matchIdx = nil
	if len(sorted) == 0 {
		lines = append(lines, styles.DimText.Render("  No fixtures for this filter."))
		f.lines = lines
		return
	}

	var lastDate string
	for i, m := range sorted {
		dateKey := tz.DisplayIn(m.KickoffAt).Format("2006-01-02")
		if dateKey != lastDate {
			lastDate = dateKey
			matchday := ""
			if m.Matchday > 0 {
				matchday = fmt.Sprintf("  ·  MATCHDAY %d", m.Matchday)
			}
			lines = append(lines, styles.GoldBold.Render(fmt.Sprintf("  %s%s", tz.FormatDateHeader(m.KickoffAt), matchday)))
		}
		f.matchIdx = append(f.matchIdx, len(lines))
		prefix := "  "
		if i == f.cursor {
			prefix = styles.GoldText.Render("> ")
		}
		row := prefix + strings.TrimPrefix(renderFixtureRow(m, nameW, venueW), "  ")
		lines = append(lines, row)
	}

	f.lines = lines
}

func (f *Fixtures) View() string {
	if len(f.lines) == 0 {
		return styles.DimText.Render("  No upcoming fixtures.\n")
	}
	vis := f.h - 8
	if vis < 4 {
		vis = 4
	}
	end := f.scroll + vis
	if end > len(f.lines) {
		end = len(f.lines)
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(f.lines[f.scroll:end], "\n"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  j/k pick · enter detail · 1/2/3 filter") + "\n")
	return sb.String()
}

func renderFixtureRow(m wc.Match, nameW, venueW int) string {
	kickoff := tz.FormatClock(m.KickoffAt)
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