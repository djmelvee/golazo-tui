package wc

import "sort"

// BracketRound is one knockout stage column.
type BracketRound struct {
	Stage  string
	Label  string
	Slots  []BracketSlot
}

// BracketSlot is a single match position in the tree.
type BracketSlot struct {
	Match  Match
	Winner string // team name, empty if TBD
}

var knockoutStages = []struct {
	stage, label string
}{
	{"r32", "Round of 32"},
	{"r16", "Round of 16"},
	{"qf", "Quarter-finals"},
	{"sf", "Semi-finals"},
	{"final", "Final"},
}

// BuildBracket groups knockout matches by stage, sorted by kickoff.
func BuildBracket(matches []Match) []BracketRound {
	byStage := make(map[string][]Match)
	for _, m := range matches {
		if !isKnockoutStage(m.Stage) {
			continue
		}
		byStage[m.Stage] = append(byStage[m.Stage], m)
	}

	var rounds []BracketRound
	for _, st := range knockoutStages {
		ms := byStage[st.stage]
		if len(ms) == 0 {
			continue
		}
		sort.Slice(ms, func(i, j int) bool {
			return ms[i].KickoffAt.Before(ms[j].KickoffAt)
		})
		slots := make([]BracketSlot, len(ms))
		for i, m := range ms {
			slots[i] = BracketSlot{Match: m, Winner: bracketWinner(m)}
		}
		rounds = append(rounds, BracketRound{Stage: st.stage, Label: st.label, Slots: slots})
	}
	return rounds
}

func isKnockoutStage(stage string) bool {
	switch stage {
	case "r32", "r16", "qf", "sf", "final",
		"round_of_32", "round_of_16", "quarter_final", "semi_final":
		return true
	default:
		return false
	}
}

func bracketWinner(m Match) string {
	if m.HomeScore == nil || m.AwayScore == nil {
		return ""
	}
	if m.Status != StatusFinished && m.Status != StatusLive {
		return ""
	}
	if *m.HomeScore > *m.AwayScore {
		return m.HomeTeam.Name
	}
	if *m.AwayScore > *m.HomeScore {
		return m.AwayTeam.Name
	}
	return ""
}

// NormalizeStage maps API type strings to short stage codes.
func NormalizeStage(t string) string {
	switch t {
	case "", "group":
		return "group"
	case "round_of_32":
		return "r32"
	case "round_of_16":
		return "r16"
	case "quarter_final", "quarterfinal":
		return "qf"
	case "semi_final", "semifinal":
		return "sf"
	case "final":
		return "final"
	default:
		return t
	}
}