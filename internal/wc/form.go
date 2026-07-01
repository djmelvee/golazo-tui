package wc

import (
	"math"
	"sort"
	"strings"
	"time"
)

// TeamForm holds tournament stats derived from finished matches.
type TeamForm struct {
	Team            Team
	Played          int
	W, D, L         int
	GF, GA          int
	Pts             int
	GoalsPerGame    float64
	ConcededPerGame float64
	FirstGoalRate   float64 // share of matches where team scored first
	HTGoalsPerGame  float64
	CleanSheetRate  float64
	Strength        float64 // 0–1 composite rating
}

// FinishedOnly returns matches with StatusFinished and a recorded score.
func FinishedOnly(matches []Match) []Match {
	var out []Match
	for _, m := range matches {
		if m.Status == StatusFinished && m.HomeScore != nil && m.AwayScore != nil {
			out = append(out, m)
		}
	}
	return out
}

// BuildTeamForms computes per-team tournament form from finished matches only.
func BuildTeamForms(matches []Match) map[int]*TeamForm {
	forms := make(map[int]*TeamForm)
	var totalGF int

	for _, m := range FinishedOnly(matches) {
		totalGF += *m.HomeScore + *m.AwayScore
		home := ensureForm(forms, m.HomeTeam)
		away := ensureForm(forms, m.AwayTeam)

		home.Played++
		away.Played++
		home.GF += *m.HomeScore
		home.GA += *m.AwayScore
		away.GF += *m.AwayScore
		away.GA += *m.HomeScore

		switch {
		case *m.HomeScore > *m.AwayScore:
			home.W++
			away.L++
		case *m.HomeScore < *m.AwayScore:
			home.L++
			away.W++
		default:
			home.D++
			away.D++
		}
		home.Pts = home.W*3 + home.D
		away.Pts = away.W*3 + away.D

		if *m.AwayScore == 0 {
			home.CleanSheetRate++
		}
		if *m.HomeScore == 0 {
			away.CleanSheetRate++
		}

		homeHT, awayHT := htGoals(m)
		home.HTGoalsPerGame += float64(homeHT)
		away.HTGoalsPerGame += float64(awayHT)

		first := firstScorerSide(m)
		if first == "home" {
			home.FirstGoalRate++
		} else if first == "away" {
			away.FirstGoalRate++
		}
	}

	avgGF := 2.6
	if n := len(matches); n > 0 && totalGF > 0 {
		finished := 0
		for _, m := range matches {
			if m.Status == StatusFinished {
				finished++
			}
		}
		if finished > 0 {
			avgGF = float64(totalGF) / float64(finished)
		}
	}

	for _, f := range forms {
		if f.Played == 0 {
			continue
		}
		p := float64(f.Played)
		f.GoalsPerGame = float64(f.GF) / p
		f.ConcededPerGame = float64(f.GA) / p
		f.FirstGoalRate /= p
		f.HTGoalsPerGame /= p
		f.CleanSheetRate /= p

		attack := f.GoalsPerGame / avgGF
		defense := 1.0 - (f.ConcededPerGame / avgGF)
		if defense < 0.1 {
			defense = 0.1
		}
		formPts := float64(f.Pts) / (p * 3.0)
		f.Strength = clampF(0.35*attack+0.30*defense+0.35*formPts, 0.15, 0.95)
	}
	return forms
}

func ensureForm(forms map[int]*TeamForm, t Team) *TeamForm {
	if forms[t.ID] == nil {
		forms[t.ID] = &TeamForm{Team: t}
	}
	return forms[t.ID]
}

func htGoals(m Match) (home, away int) {
	for _, s := range m.HomeScorers {
		if s.Minute <= 45 {
			home++
		}
	}
	for _, s := range m.AwayScorers {
		if s.Minute <= 45 {
			away++
		}
	}
	if len(m.HomeScorers)+len(m.AwayScorers) == 0 && m.HomeScore != nil {
		// Estimate: ~42% of goals in first half when no scorer data.
		home = int(math.Round(float64(*m.HomeScore) * 0.42))
		away = int(math.Round(float64(*m.AwayScore) * 0.42))
	}
	return home, away
}

func firstScorerSide(m Match) string {
	bestMin := 999
	side := ""
	check := func(scorers []Scorer, s string) {
		for _, g := range scorers {
			if g.Minute > 0 && g.Minute < bestMin {
				bestMin = g.Minute
				side = s
			}
		}
	}
	check(m.HomeScorers, "home")
	check(m.AwayScorers, "away")
	if side != "" {
		return side
	}
	if m.HomeScore != nil && m.AwayScore != nil {
		if *m.HomeScore > 0 && *m.AwayScore == 0 {
			return "home"
		}
		if *m.AwayScore > 0 && *m.HomeScore == 0 {
			return "away"
		}
	}
	return ""
}

// TournamentPhase returns "knockout" when no group-stage matches remain open.
func TournamentPhase(matches []Match) string {
	for _, m := range matches {
		st := m.Stage
		if st == "" || st == "group" {
			if m.Status == StatusUpcoming || m.Status == StatusLive {
				return "group"
			}
		}
	}
	return "knockout"
}

// MatchIsPredictable reports whether a fixture has known teams and a future kickoff.
func MatchIsPredictable(m Match, now time.Time) bool {
	if strings.TrimSpace(m.HomeTeam.Name) == "" || strings.TrimSpace(m.AwayTeam.Name) == "" {
		return false
	}
	if m.KickoffAt.IsZero() {
		return false
	}
	return m.KickoffAt.After(now.Add(-2 * time.Minute))
}

// UpcomingForPredict returns future fixtures with known teams, sorted by kickoff.
func UpcomingForPredict(matches []Match, now time.Time) []Match {
	var out []Match
	for _, m := range matches {
		if m.Status != StatusUpcoming {
			continue
		}
		if !MatchIsPredictable(m, now) {
			continue
		}
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].KickoffAt.Before(out[j].KickoffAt)
	})
	return out
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}