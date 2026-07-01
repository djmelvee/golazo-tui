package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// Predictions shows data-driven match forecasts.
type Predictions struct {
	w, h    int
	preds   []wc.MatchPrediction
	cursor  int
	scroll  int
	lines   []string
	itemAt  []int
}

func (p *Predictions) SetSize(w, h int) {
	p.w = w
	p.h = h
	p.lines = p.renderLines()
}

func (p *Predictions) Load(db *data.Store) {
	now := time.Now()
	all := db.AllMatches()
	if cached, ok := db.GetPredictions(); ok && len(cached) > 0 {
		p.preds = enrichPredictions(cached, all)
	} else {
		p.preds = wc.BuildPredictions(all, now)
		_ = db.SetPredictions(p.preds)
	}
	if p.cursor >= len(p.preds) {
		p.cursor = max(0, len(p.preds)-1)
	}
	p.lines = p.renderLines()
}

func enrichPredictions(preds []wc.MatchPrediction, all []wc.Match) []wc.MatchPrediction {
	byID := make(map[int]wc.Match)
	for _, m := range all {
		byID[m.ID] = m
	}
	for i := range preds {
		if actual, ok := byID[preds[i].Match.ID]; ok && actual.Status == wc.StatusFinished {
			preds[i].Accuracy = wc.EvaluateAccuracy(preds[i], actual)
		}
	}
	return preds
}

func (p *Predictions) CursorDown() {
	if p.cursor < len(p.preds)-1 {
		p.cursor++
		p.ensureVisible()
		p.lines = p.renderLines()
	}
}

func (p *Predictions) CursorUp() {
	if p.cursor > 0 {
		p.cursor--
		p.ensureVisible()
		p.lines = p.renderLines()
	}
}

func (p *Predictions) ensureVisible() {
	vis := p.h - 10
	if vis < 6 {
		vis = 6
	}
	if p.cursor < len(p.itemAt) {
		starts := []int{p.itemAt[p.cursor]}
		heights := []int{estimatePredLines(p.preds[p.cursor])}
		p.scroll = ScrollToItem(0, starts, heights, vis, p.scroll)
	}
}

func estimatePredLines(pr wc.MatchPrediction) int {
	n := 3
	if p := pr; p.Summary != "" {
		n += 2
	}
	return n
}

func (p *Predictions) SelectedMatch() *wc.Match {
	if pr := p.SelectedPrediction(); pr != nil {
		m := pr.Match
		return &m
	}
	return nil
}

func (p *Predictions) SelectedPrediction() *wc.MatchPrediction {
	if p.cursor < 0 || p.cursor >= len(p.preds) {
		return nil
	}
	pr := p.preds[p.cursor]
	return &pr
}

func (p *Predictions) View() string {
	if len(p.preds) == 0 {
		return styles.DimText.Render("  No upcoming matches to predict.\n")
	}
	vis := p.h - 8
	if vis < 6 {
		vis = 6
	}
	end := p.scroll + vis
	if end > len(p.lines) {
		end = len(p.lines)
	}
	var sb strings.Builder
	sb.WriteString(strings.Join(p.lines[p.scroll:end], "\n"))
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  j/k pick · enter full breakdown · "+tz.DisplayLabel()+" time") + "\n")
	return sb.String()
}

func (p *Predictions) renderLines() []string {
	var lines []string
	p.itemAt = make([]int, len(p.preds))
	lines = append(lines, styles.Heading.Render("  ⚽ MATCH PREDICTIONS")+"\n")
	lines = append(lines, styles.GoldText.Render("  Form-weighted forecasts for upcoming fixtures")+"\n")
	lines = append(lines, styles.DimText.Render(fmt.Sprintf("  %d fixtures · enter for reasoning", len(p.preds)))+"\n")
	lines = append(lines, "")

	for i, pr := range p.preds {
		p.itemAt[i] = len(lines)
		m := pr.Match
		kick := tz.FormatKickoff(m.KickoffAt)
		sel := i == p.cursor
		prefix := "  "
		if sel {
			prefix = styles.GoldText.Render("▶ ")
		}
		score := styles.GoldBold.Render(fmt.Sprintf("%d–%d", pr.FTHome, pr.FTAway))
		badge := ""
		if pr.LowData {
			badge = " " + styles.DimText.Render("[low data]")
		}
		if pr.Accuracy != "" {
			badge += " " + styles.GoldText.Render(pr.Accuracy)
		}
		row := fmt.Sprintf("%s%s  %s  %s %s %s %s %s%s",
			prefix, kick, wc.StageLabel(m.Stage),
			m.HomeTeam.Flag, truncate(m.HomeTeam.Name, 11),
			score,
			m.AwayTeam.Flag, truncate(m.AwayTeam.Name, 11), badge,
		)
		lines = append(lines, row)
		conf := pr.Confidence
		if pr.LowData {
			conf = "insufficient data"
		}
		meta := fmt.Sprintf("     xG %.1f–%.1f · HT %d–%d · win %.0f/%.0f/%.0f%% · %s",
			pr.HomeXG, pr.AwayXG, pr.HTHome, pr.HTAway,
			pr.HomeWinProb*100, pr.DrawProb*100, pr.AwayWinProb*100, conf)
		if sel {
			lines = append(lines, styles.GoldText.Render(meta))
			if pr.Summary != "" {
				for _, l := range wrapLines("     "+pr.Summary, p.w) {
					lines = append(lines, styles.MainText.Render(l))
				}
			}
		} else {
			lines = append(lines, styles.DimText.Render(meta))
		}
		lines = append(lines, "")
	}
	return lines
}