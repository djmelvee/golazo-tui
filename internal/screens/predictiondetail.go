package screens

import (
	"fmt"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// PredictionDetail shows the full forecast breakdown for one fixture.
type PredictionDetail struct {
	w, h   int
	pred   wc.MatchPrediction
	scroll int
	lines  []string
}

func (d *PredictionDetail) SetSize(w, h int) {
	d.w = w
	d.h = h
}

func (d *PredictionDetail) Set(pred wc.MatchPrediction) {
	d.pred = pred
	d.scroll = 0
	d.lines = d.buildLines()
}

func (d *PredictionDetail) ScrollDown() {
	max := len(d.lines) - (d.h - 6)
	if max < 0 {
		max = 0
	}
	if d.scroll < max {
		d.scroll++
	}
}

func (d *PredictionDetail) ScrollUp() {
	if d.scroll > 0 {
		d.scroll--
	}
}

func (d *PredictionDetail) View() string {
	if len(d.lines) == 0 {
		return ""
	}
	vis := d.h - 6
	if vis < 8 {
		vis = 8
	}
	end := d.scroll + vis
	if end > len(d.lines) {
		end = len(d.lines)
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(d.lines[d.scroll:end], "\n"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  b/esc back · j/k scroll"))
	return sb.String()
}

func (d *PredictionDetail) buildLines() []string {
	p := d.pred
	m := p.Match
	var lines []string

	lines = append(lines, styles.Heading.Render("  ⚽ PREDICTION BREAKDOWN"))
	lines = append(lines, "")
	lines = append(lines, styles.GoldBold.Render(fmt.Sprintf("  %s %s  vs  %s %s",
		m.HomeTeam.Flag, m.HomeTeam.Name, m.AwayTeam.Name, m.AwayTeam.Flag)))
	lines = append(lines, styles.DimText.Render("  "+wc.StageLabel(m.Stage)+" · "+tz.FormatKickoff(m.KickoffAt)+" · confidence "+p.Confidence))
	lines = append(lines, "")

	lines = append(lines, styles.GoldText.Render("  SUMMARY"))
	if p.Summary != "" {
		for _, l := range wrapLines("  "+p.Summary, d.w) {
			lines = append(lines, styles.GoldBold.Render(l))
		}
	}
	lines = append(lines, "")

	lines = append(lines, styles.GoldText.Render("  FORECAST"))
	lines = append(lines, fmt.Sprintf("  Half-time     %s %d – %d %s",
		m.HomeTeam.Flag, p.HTHome, p.HTAway, m.AwayTeam.Flag))
	lines = append(lines, fmt.Sprintf("  Full-time     %s %d – %d %s",
		m.HomeTeam.Flag, p.FTHome, p.FTAway, m.AwayTeam.Flag))
	lines = append(lines, fmt.Sprintf("  First goal    %s (~%d')", p.FirstScorer, p.FirstScorerMin))
	lines = append(lines, fmt.Sprintf("  Expected goals %.1f – %.1f", p.HomeXG, p.AwayXG))
	lines = append(lines, fmt.Sprintf("  Win / Draw / Loss   %.0f%% / %.0f%% / %.0f%%",
		p.HomeWinProb*100, p.DrawProb*100, p.AwayWinProb*100))

	if len(p.TopScores) > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.DimText.Render("  Other likely scores:"))
		for i, s := range p.TopScores {
			if i >= 5 {
				break
			}
			mark := " "
			if s.Home == p.FTHome && s.Away == p.FTAway {
				mark = "▶"
			}
			lines = append(lines, fmt.Sprintf("  %s %d–%d  (%.0f%%)", mark, s.Home, s.Away, s.Prob*100))
		}
	}

	if m.Stage != "" && m.Stage != "group" {
		lines = append(lines, "")
		lines = append(lines, styles.GoldText.Render("  KNOCKOUT"))
		if p.ExtraTime {
			lines = append(lines, fmt.Sprintf("  Extra time    %d – %d", p.ETHome, p.ETAway))
		}
		if p.Penalties {
			lines = append(lines, fmt.Sprintf("  Penalties     %d – %d → %s", p.PensHome, p.PensAway, p.PenWinner))
		}
	}

	lines = append(lines, "")
	lines = append(lines, styles.GoldText.Render("  WHY THIS PREDICTION"))
	for _, r := range p.Reasons {
		for _, l := range wrapLines("    · "+r, d.w) {
			lines = append(lines, l)
		}
	}

	if len(p.RecentHome) > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.DimText.Render("  "+m.HomeTeam.Name+" recent WC results:"))
		for _, r := range p.RecentHome {
			lines = append(lines, "    · "+r)
		}
	}
	if len(p.RecentAway) > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.DimText.Render("  "+m.AwayTeam.Name+" recent WC results:"))
		for _, r := range p.RecentAway {
			lines = append(lines, "    · "+r)
		}
	}

	if len(p.H2H) > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.GoldText.Render("  HEAD-TO-HEAD (tournament)"))
		for _, h := range p.H2H {
			lines = append(lines, "    · "+h)
		}
	}

	if len(p.Facts) > 0 {
		lines = append(lines, "")
		lines = append(lines, styles.DimText.Render("  Tournament stats:"))
		for _, f := range p.Facts {
			lines = append(lines, "    · "+f)
		}
	}

	return lines
}