package screens

import (
	"fmt"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Bracket shows the knockout-stage ASCII tree.
type Bracket struct {
	w, h    int
	rounds  []wc.BracketRound
	scroll  int
	lines   []string
	cursor  int
}

func (b *Bracket) SetSize(w, h int) {
	b.w = w
	b.h = h
	b.lines = b.renderLines()
}

func (b *Bracket) Load(db *data.Store) {
	b.rounds = wc.BuildBracket(db.AllMatches())
	b.scroll = 0
	b.cursor = 0
	b.lines = b.renderLines()
}

func (b *Bracket) ScrollDown() {
	max := len(b.lines) - (b.h - 8)
	if max < 0 {
		max = 0
	}
	if b.scroll < max {
		b.scroll++
	}
}

func (b *Bracket) ScrollUp() {
	if b.scroll > 0 {
		b.scroll--
	}
}

func (b *Bracket) CursorDown() {
	if b.cursor < len(b.selectable())-1 {
		b.cursor++
		b.lines = b.renderLines()
		b.ensureVisible()
	}
}

func (b *Bracket) CursorUp() {
	if b.cursor > 0 {
		b.cursor--
		b.lines = b.renderLines()
		b.ensureVisible()
	}
}

func (b *Bracket) ensureVisible() {
	vis := b.h - 10
	if vis < 4 {
		vis = 4
	}
	line := 3 + b.cursor*2
	if line < b.scroll {
		b.scroll = line
	}
	if line >= b.scroll+vis {
		b.scroll = line - vis + 1
	}
}

func (b *Bracket) SelectedMatchID() int {
	sel := b.selectable()
	if b.cursor >= 0 && b.cursor < len(sel) {
		return sel[b.cursor]
	}
	return 0
}

func (b *Bracket) selectable() []int {
	var ids []int
	for _, r := range b.rounds {
		for _, sl := range r.Slots {
			if sl.Match.ID > 0 {
				ids = append(ids, sl.Match.ID)
			}
		}
	}
	return ids
}

func (b *Bracket) View() string {
	if len(b.lines) == 0 {
		return styles.DimText.Render("  No knockout data yet.\n")
	}
	vis := b.h - 8
	if vis < 4 {
		vis = 4
	}
	start := b.scroll
	end := start + vis
	if end > len(b.lines) {
		end = len(b.lines)
	}
	if start >= len(b.lines) {
		start = 0
		end = vis
		if end > len(b.lines) {
			end = len(b.lines)
		}
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(b.lines[start:end], "\n"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  chronological layout · j/k scroll · enter detail") + "\n")
	return sb.String()
}

func (b *Bracket) renderLines() []string {
	var lines []string
	lines = append(lines, styles.Heading.Render("  KNOCKOUT BRACKET")+"\n")
	lines = append(lines, "")

	selIdx := b.cursor

	idx := 0
	for _, round := range b.rounds {
		lines = append(lines, styles.GoldText.Render("  "+round.Label)+"  "+styles.DimText.Render("("+round.Stage+")"))
		for i, sl := range round.Slots {
			prefix := "├──"
			if i == len(round.Slots)-1 {
				prefix = "└──"
			}
			line := b.slotLine(sl, prefix, idx == selIdx)
			lines = append(lines, line)
			idx++
		}
		lines = append(lines, "")
	}
	return lines
}

func (b *Bracket) slotLine(sl wc.BracketSlot, prefix string, selected bool) string {
	m := sl.Match
	score := "– –"
	if m.HomeScore != nil && m.AwayScore != nil {
		score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
	}
	nameW := clamp((b.w-40)/2, 10, 18)
	row := fmt.Sprintf("  %s %s %s  %s  %s %s",
		prefix, m.HomeTeam.Flag, truncate(m.HomeTeam.Name, nameW),
		styles.GoldBold.Render(score),
		m.AwayTeam.Flag, truncate(m.AwayTeam.Name, nameW),
	)
		if sl.Winner != "" {
			row += "  " + styles.GoldText.Render("→ "+truncate(sl.Winner, 12))
		} else if m.Status == wc.StatusFinished && m.HomeScore != nil && m.AwayScore != nil {
			w := m.HomeTeam.Name
			if *m.AwayScore > *m.HomeScore {
				w = m.AwayTeam.Name
			} else if *m.HomeScore == *m.AwayScore {
				w = "draw"
			}
			row += "  " + styles.DimText.Render("if done: "+truncate(w, 12))
		}
	if selected {
		row = styles.GoldText.Render("> ") + row[2:]
	}
	return row
}