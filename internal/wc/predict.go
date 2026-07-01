package wc

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// MatchPrediction is a data-driven forecast for one fixture.
type MatchPrediction struct {
	Match          Match
	HTHome         int
	HTAway         int
	FTHome         int
	FTAway         int
	HomeXG         float64
	AwayXG         float64
	HomeWinProb    float64
	DrawProb       float64
	AwayWinProb    float64
	FirstScorer    string
	FirstScorerMin int
	EndsDrawFT     bool
	ExtraTime      bool
	ETGoals        bool
	ETHome         int
	ETAway         int
	Penalties      bool
	PensHome       int
	PensAway       int
	PenWinner      string
	Confidence     string
	Summary        string
	TopScores      []ScoreLine
	Facts          []string
	Reasons        []string
	RecentHome     []string
	RecentAway     []string
	LowData        bool
	Accuracy       string // set when a finished result exists for comparison
	H2H            []string
}

// ScoreLine is a candidate scoreline with model probability.
type ScoreLine struct {
	Home int
	Away int
	Prob float64
}

// BuildPredictions generates forecasts for future fixtures with known teams.
func BuildPredictions(matches []Match, now time.Time) []MatchPrediction {
	scored := ScoredMatches(matches)
	forms := BuildTeamForms(scored)
	topScorers := BuildTopScorers(scored)
	avgFirstGoal := tournamentAvgFirstGoalMinute(scored)
	upcoming := UpcomingForPredict(matches, now)

	var preds []MatchPrediction
	for _, m := range upcoming {
		preds = append(preds, predictMatch(m, forms, scored, topScorers, avgFirstGoal))
	}
	return preds
}

// ScoredMatches returns every match with a recorded score (group + knockout).
func ScoredMatches(matches []Match) []Match {
	var out []Match
	for _, m := range matches {
		if m.HomeScore != nil && m.AwayScore != nil {
			out = append(out, m)
		}
	}
	return out
}

func predictMatch(m Match, forms map[int]*TeamForm, scored []Match, topScorers []ScorerRow, avgFirstGoal int) MatchPrediction {
	homeF := lookupForm(forms, m.HomeTeam)
	awayF := lookupForm(forms, m.AwayTeam)
	if homeF == nil {
		homeF = &TeamForm{Team: m.HomeTeam, GoalsPerGame: 1.1, ConcededPerGame: 1.1, FirstGoalRate: 0.5, Strength: 0.45}
	}
	if awayF == nil {
		awayF = &TeamForm{Team: m.AwayTeam, GoalsPerGame: 1.1, ConcededPerGame: 1.1, FirstGoalRate: 0.5, Strength: 0.45}
	}

	knockout := m.Stage != "" && m.Stage != "group"
	avgGoals := tournamentAvgGoals(scored)
	homeXG, awayXG := expectedGoals(homeF, awayF, avgGoals, knockout)

	homeWin, draw, awayWin := outcomeProbs(homeXG, awayXG)
	topScores := topScorelines(homeXG, awayXG, 5)
	ftHome, ftAway := scoreFromXG(homeXG, awayXG, homeWin, draw, awayWin)
	htHome, htAway := scoreFromXG(homeXG*htShare(homeF), awayXG*htShare(awayF), homeWin, draw, awayWin)
	ftProb := scoreProbability(homeXG, awayXG, ftHome, ftAway)

	drawProb := draw
	endsDraw := ftHome == ftAway

	firstTeam, firstPlayer, firstMin := pickFirstScorer(m, homeF, awayF, topScorers, avgFirstGoal, htHome+htAway == 0)

	p := MatchPrediction{
		Match:          m,
		HTHome:         htHome,
		HTAway:         htAway,
		FTHome:         ftHome,
		FTAway:         ftAway,
		HomeXG:         homeXG,
		AwayXG:         awayXG,
		HomeWinProb:    homeWin,
		DrawProb:       drawProb,
		AwayWinProb:    awayWin,
		FirstScorerMin: firstMin,
		EndsDrawFT:     endsDraw,
		RecentHome:     recentResults(m.HomeTeam, scored, 3),
		RecentAway:     recentResults(m.AwayTeam, scored, 3),
		TopScores:      topScores,
	}
	if firstPlayer == "" {
		p.FirstScorer = firstTeam
	} else {
		p.FirstScorer = firstPlayer + " (" + firstTeam + ")"
	}

	p.Facts = buildFacts(m, homeF, awayF, knockout, scored)
	p.Reasons = buildReasons(m, homeF, awayF, homeXG, awayXG, ftHome, ftAway, ftProb, homeWin, draw, awayWin, knockout, topScores)
	p.Summary = buildSummary(m, homeF, awayF, ftHome, ftAway, homeXG, awayXG, homeWin, draw, awayWin, knockout)
	p.Confidence = confidenceLevel(homeF, awayF, knockout)
	p.LowData = homeF.Played < 2 || awayF.Played < 2
	if p.LowData {
		p.Confidence = "low data"
	}
	p.H2H = formatH2H(m.HomeTeam, m.AwayTeam, scored)

	if knockout {
		if endsDraw || drawProb >= 0.24 {
			p.ExtraTime = true
			etHXG, etAXG := homeXG*0.35, awayXG*0.35
			p.ETHome, p.ETAway = scoreFromXG(float64(ftHome)+etHXG, float64(ftAway)+etAXG, homeWin, draw, awayWin)
			p.ETGoals = p.ETHome+p.ETAway > ftHome+ftAway
			if p.ETHome == p.ETAway {
				p.Penalties = true
				p.PensHome, p.PensAway, p.PenWinner = predictPens(m, homeF, awayF)
				p.Reasons = append(p.Reasons, fmt.Sprintf(
					"Knockout draw likely (%.0f%%) — extra time then pens favoured for %s",
					drawProb*100, p.PenWinner))
			}
		}
	}

	return p
}

func lookupForm(forms map[int]*TeamForm, t Team) *TeamForm {
	if t.ID != 0 {
		if f := forms[t.ID]; f != nil {
			return f
		}
	}
	for _, f := range forms {
		if f.Team.Name == t.Name {
			return f
		}
	}
	return nil
}

func expectedGoals(homeF, awayF *TeamForm, avgGoals float64, knockout bool) (home, away float64) {
	if avgGoals < 1.0 {
		avgGoals = 2.5
	}
	// Floor conceded rate — clean sheets shouldn't zero-out opponent xG entirely.
	homeDef := math.Max(homeF.ConcededPerGame, avgGoals*0.38)
	awayDef := math.Max(awayF.ConcededPerGame, avgGoals*0.38)

	homeAdv := 1.12
	if knockout {
		homeAdv = 1.06
	}

	// Blend each side's attack with how much the opponent concedes.
	home = (homeF.GoalsPerGame*0.55 + awayDef*0.45) * homeAdv
	away = (awayF.GoalsPerGame*0.55 + homeDef*0.45) * 0.94

	home = clampF(home, 0.35, 4.2)
	away = clampF(away, 0.35, 4.0)
	return home, away
}

// scoreFromXG maps expected goals to a realistic scoreline (not always 1-0).
func scoreFromXG(homeXG, awayXG, homeWin, draw, awayWin float64) (int, int) {
	h := int(math.Round(homeXG))
	a := int(math.Round(awayXG))

	if h == 0 && a == 0 && homeXG+awayXG >= 0.75 {
		switch {
		case homeWin >= awayWin && homeWin >= draw:
			h = int(math.Ceil(homeXG))
			if h == 0 {
				h = 1
			}
		case awayWin >= homeWin && awayWin >= draw:
			a = int(math.Ceil(awayXG))
			if a == 0 {
				a = 1
			}
		default:
			h, a = 1, 1
		}
	}

	// High xG teams should not be capped at a single goal.
	if homeXG >= 1.65 && h < 2 {
		h = 2
	}
	if awayXG >= 1.65 && a < 2 {
		a = 2
	}
	if homeXG >= 2.4 && h < 3 {
		h = 3
	}
	if awayXG >= 2.4 && a < 3 {
		a = 3
	}

	// Align with the most likely outcome when rounding sits on a draw.
	if h == a && h > 0 {
		switch {
		case homeWin > awayWin+0.08:
			h++
		case awayWin > homeWin+0.08:
			a++
		}
	}

	h = clampInt(h, 0, 5)
	a = clampInt(a, 0, 5)
	return h, a
}

func topScorelines(homeXG, awayXG float64, n int) []ScoreLine {
	type pair struct {
		ScoreLine
	}
	var all []ScoreLine
	for h := 0; h <= 5; h++ {
		for a := 0; a <= 5; a++ {
			all = append(all, ScoreLine{
				Home: h, Away: a,
				Prob: poissonPMF(h, homeXG) * poissonPMF(a, awayXG),
			})
		}
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Prob > all[j].Prob
	})
	if n > len(all) {
		n = len(all)
	}
	return all[:n]
}

func scoreProbability(homeXG, awayXG float64, h, a int) float64 {
	return poissonPMF(h, homeXG) * poissonPMF(a, awayXG)
}

func outcomeProbs(homeExp, awayExp float64) (homeWin, draw, awayWin float64) {
	for h := 0; h <= 5; h++ {
		for a := 0; a <= 5; a++ {
			p := poissonPMF(h, homeExp) * poissonPMF(a, awayExp)
			switch {
			case h > a:
				homeWin += p
			case h == a:
				draw += p
			default:
				awayWin += p
			}
		}
	}
	total := homeWin + draw + awayWin
	if total > 0 {
		homeWin, draw, awayWin = homeWin/total, draw/total, awayWin/total
	}
	return homeWin, draw, awayWin
}

func poissonPMF(k int, lambda float64) float64 {
	if lambda <= 0 {
		if k == 0 {
			return 1
		}
		return 0
	}
	return math.Exp(-lambda) * math.Pow(lambda, float64(k)) / float64(factorial(k))
}

func factorial(n int) int {
	if n <= 1 {
		return 1
	}
	return n * factorial(n-1)
}

func tournamentAvgFirstGoalMinute(matches []Match) int {
	var total, count int
	for _, m := range matches {
		best := 999
		for _, s := range append(m.HomeScorers, m.AwayScorers...) {
			if s.Minute > 0 && s.Minute < best {
				best = s.Minute
			}
		}
		if best < 999 {
			total += best
			count++
		}
	}
	if count == 0 {
		return 28
	}
	return total / count
}

func pickFirstScorer(m Match, homeF, awayF *TeamForm, topScorers []ScorerRow, avgMin int, goallessHT bool) (team, player string, minute int) {
	homeP := homeF.FirstGoalRate*0.55 + homeF.Strength*0.45
	awayP := awayF.FirstGoalRate*0.55 + awayF.Strength*0.45
	if homeP >= awayP {
		team = m.HomeTeam.Name
		player = topScorerForTeam(topScorers, m.HomeTeam.Name)
	} else {
		team = m.AwayTeam.Name
		player = topScorerForTeam(topScorers, m.AwayTeam.Name)
	}
	minute = avgMin
	if goallessHT {
		minute = clampInt(minute+12, 18, 44)
	} else {
		minute = clampInt(minute-6, 6, 38)
	}
	return team, player, minute
}

func topScorerForTeam(scorers []ScorerRow, team string) string {
	for _, s := range scorers {
		if s.Team == team {
			return s.Name
		}
	}
	return ""
}

func recentResults(team Team, matches []Match, n int) []string {
	type res struct {
		kick time.Time
		line string
	}
	var rows []res
	for _, m := range matches {
		if m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		var line string
		switch {
		case m.HomeTeam.Name == team.Name:
			switch {
			case *m.HomeScore > *m.AwayScore:
				line = fmt.Sprintf("W %d–%d vs %s", *m.HomeScore, *m.AwayScore, m.AwayTeam.Name)
			case *m.HomeScore < *m.AwayScore:
				line = fmt.Sprintf("L %d–%d vs %s", *m.HomeScore, *m.AwayScore, m.AwayTeam.Name)
			default:
				line = fmt.Sprintf("D %d–%d vs %s", *m.HomeScore, *m.AwayScore, m.AwayTeam.Name)
			}
		case m.AwayTeam.Name == team.Name:
			switch {
			case *m.AwayScore > *m.HomeScore:
				line = fmt.Sprintf("W %d–%d @ %s", *m.AwayScore, *m.HomeScore, m.HomeTeam.Name)
			case *m.AwayScore < *m.HomeScore:
				line = fmt.Sprintf("L %d–%d @ %s", *m.AwayScore, *m.HomeScore, m.HomeTeam.Name)
			default:
				line = fmt.Sprintf("D %d–%d @ %s", *m.AwayScore, *m.HomeScore, m.HomeTeam.Name)
			}
		default:
			continue
		}
		rows = append(rows, res{kick: m.KickoffAt, line: line})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].kick.Before(rows[j].kick)
	})
	var out []string
	for i := len(rows) - 1; i >= 0 && len(out) < n; i-- {
		out = append(out, rows[i].line)
	}
	return out
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func htShare(f *TeamForm) float64 {
	if f.HTGoalsPerGame > 0 && f.GoalsPerGame > 0 {
		r := f.HTGoalsPerGame / f.GoalsPerGame
		return clampF(r, 0.30, 0.55)
	}
	return 0.42
}

func tournamentAvgGoals(matches []Match) float64 {
	if len(matches) == 0 {
		return 2.6
	}
	total := 0
	for _, m := range matches {
		if m.HomeScore != nil {
			total += *m.HomeScore + *m.AwayScore
		}
	}
	return float64(total) / float64(len(matches))
}

func predictPens(m Match, homeF, awayF *TeamForm) (home, away int, winner string) {
	diff := homeF.Strength - awayF.Strength
	if diff >= 0.08 {
		home, away = 5, 4
		winner = m.HomeTeam.Name
	} else if diff <= -0.08 {
		home, away = 4, 5
		winner = m.AwayTeam.Name
	} else {
		if homeF.Pts >= awayF.Pts {
			home, away = 5, 4
			winner = m.HomeTeam.Name
		} else {
			home, away = 4, 5
			winner = m.AwayTeam.Name
		}
	}
	return home, away, winner
}

func buildFacts(m Match, homeF, awayF *TeamForm, knockout bool, scored []Match) []string {
	var facts []string
	if homeF.Played > 0 {
		facts = append(facts, fmt.Sprintf("%s: %.1f gpg, %.1f conceded, %d pts (%dW-%dD-%dL)",
			m.HomeTeam.Name, homeF.GoalsPerGame, homeF.ConcededPerGame, homeF.Pts, homeF.W, homeF.D, homeF.L))
	}
	if awayF.Played > 0 {
		facts = append(facts, fmt.Sprintf("%s: %.1f gpg, %.1f conceded, %d pts (%dW-%dD-%dL)",
			m.AwayTeam.Name, awayF.GoalsPerGame, awayF.ConcededPerGame, awayF.Pts, awayF.W, awayF.D, awayF.L))
	}
	if h2h := headToHeadFact(m, scored); h2h != "" {
		facts = append(facts, h2h)
	}
	if knockout {
		facts = append(facts, "Knockout — level at 90' goes to extra time")
	}
	if homeF.FirstGoalRate > 0.55 {
		facts = append(facts, m.HomeTeam.Name+" scored first in "+pct(homeF.FirstGoalRate)+" of WC matches")
	}
	if awayF.FirstGoalRate > 0.55 {
		facts = append(facts, m.AwayTeam.Name+" scored first in "+pct(awayF.FirstGoalRate)+" of WC matches")
	}
	return facts
}

func buildSummary(m Match, homeF, awayF *TeamForm, ftH, ftA int, homeXG, awayXG, homeWin, draw, awayWin float64, knockout bool) string {
	stage := StageLabel(m.Stage)
	favorite := m.HomeTeam.Name
	favProb := homeWin
	if awayWin > homeWin {
		favorite = m.AwayTeam.Name
		favProb = awayWin
	} else if draw > homeWin && draw > awayWin {
		return fmt.Sprintf(
			"A %d–%d draw is the call for this %s: both sides average %.1f and %.1f goals per game in the tournament, "+
				"and the model rates a stalemate at %.0f%% (xG %.1f–%.1f).",
			ftH, ftA, stage, homeF.GoalsPerGame, awayF.GoalsPerGame, draw*100, homeXG, awayXG)
	}

	return fmt.Sprintf(
		"Predicting %d–%d in this %s: %s favoured (%.0f%%). %s average %.1f goals/game; %s concede %.1f. "+
			"Model xG %.1f–%.1f, draw chance %.0f%%.",
		ftH, ftA, stage, favorite, favProb*100,
		m.HomeTeam.Name, homeF.GoalsPerGame, m.AwayTeam.Name, awayF.ConcededPerGame,
		homeXG, awayXG, draw*100)
}

func buildReasons(m Match, homeF, awayF *TeamForm, homeXG, awayXG float64, ftH, ftA int, scoreProb, homeWin, draw, awayWin float64, knockout bool, tops []ScoreLine) []string {
	var r []string
	r = append(r, fmt.Sprintf("xG %.1f–%.1f built from WC attack (%.1f / %.1f gpg) vs opponent defence (%.1f / %.1f conceded)",
		homeXG, awayXG, homeF.GoalsPerGame, awayF.GoalsPerGame, awayF.ConcededPerGame, homeF.ConcededPerGame))
	r = append(r, fmt.Sprintf("Rounded forecast %d–%d (%.0f%% Poisson mass on that exact score)",
		ftH, ftA, scoreProb*100))
	r = append(r, fmt.Sprintf("Outcome split: %s %.0f%% · draw %.0f%% · %s %.0f%%",
		m.HomeTeam.Name, homeWin*100, draw*100, m.AwayTeam.Name, awayWin*100))
	if len(tops) >= 4 {
		r = append(r, fmt.Sprintf("Also plausible: %d–%d (%.0f%%), %d–%d (%.0f%%), %d–%d (%.0f%%)",
			tops[1].Home, tops[1].Away, tops[1].Prob*100,
			tops[2].Home, tops[2].Away, tops[2].Prob*100,
			tops[3].Home, tops[3].Away, tops[3].Prob*100))
	}
	if homeF.Strength > awayF.Strength+0.1 {
		r = append(r, fmt.Sprintf("%s rank higher on composite WC form (attack + defence + points)", m.HomeTeam.Name))
	} else if awayF.Strength > homeF.Strength+0.1 {
		r = append(r, fmt.Sprintf("%s rank higher on composite WC form (attack + defence + points)", m.AwayTeam.Name))
	}
	if knockout {
		r = append(r, "Knockout rules: level at 90' → extra time → penalties")
	}
	return r
}

func headToHeadFact(m Match, scored []Match) string {
	for _, f := range scored {
		if (f.HomeTeam.ID == m.HomeTeam.ID && f.AwayTeam.ID == m.AwayTeam.ID) ||
			(f.HomeTeam.ID == m.AwayTeam.ID && f.AwayTeam.ID == m.HomeTeam.ID) ||
			(f.HomeTeam.Name == m.HomeTeam.Name && f.AwayTeam.Name == m.AwayTeam.Name) {
			if f.HomeScore != nil && f.AwayScore != nil {
				return fmt.Sprintf("Met in tournament: %s %d–%d %s",
					f.HomeTeam.Name, *f.HomeScore, *f.AwayScore, f.AwayTeam.Name)
			}
		}
	}
	return ""
}

func pct(v float64) string {
	return fmt.Sprintf("%.0f%%", v*100)
}

func formatH2H(home, away Team, scored []Match) []string {
	var lines []string
	for _, f := range scored {
		if (f.HomeTeam.Name == home.Name && f.AwayTeam.Name == away.Name) ||
			(f.HomeTeam.Name == away.Name && f.AwayTeam.Name == home.Name) {
			if f.HomeScore != nil && f.AwayScore != nil {
				lines = append(lines, fmt.Sprintf("%s %d–%d %s (%s)",
					f.HomeTeam.Flag, *f.HomeScore, *f.AwayScore, f.AwayTeam.Flag, StageLabel(f.Stage)))
			}
		}
	}
	return lines
}

// EvaluateAccuracy compares a prediction to a finished result.
func EvaluateAccuracy(pred MatchPrediction, actual Match) string {
	if actual.Status != StatusFinished || actual.HomeScore == nil || actual.AwayScore == nil {
		return ""
	}
	ah, aa := *actual.HomeScore, *actual.AwayScore
	if pred.FTHome == ah && pred.FTAway == aa {
		return "✓ exact"
	}
	diff := abs(pred.FTHome-ah) + abs(pred.FTAway-aa)
	if diff <= 1 {
		return "≈ close"
	}
	return "✗ off"
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func confidenceLevel(homeF, awayF *TeamForm, knockout bool) string {
	if homeF.Played < 2 || awayF.Played < 2 {
		return "low"
	}
	gap := math.Abs(homeF.Strength - awayF.Strength)
	if gap > 0.18 && !knockout {
		return "high"
	}
	if gap > 0.12 {
		return "medium"
	}
	return "low"
}

// StageLabel returns a human-readable stage name.
func StageLabel(stage string) string {
	switch stage {
	case "group", "":
		return "Group stage"
	case "r32":
		return "Round of 32"
	case "r16":
		return "Round of 16"
	case "qf":
		return "Quarter-final"
	case "sf":
		return "Semi-final"
	case "third":
		return "Third-place play-off"
	case "final":
		return "Final"
	default:
		return strings.ReplaceAll(stage, "_", " ")
	}
}