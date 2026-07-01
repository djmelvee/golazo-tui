package screens

import (
	"strings"
	"testing"

	"github.com/djmelvee/golazo-tui/internal/wc"
)

func TestBuildEventLog_LiveWithGoals(t *testing.T) {
	hs, as := 2, 1
	min := 67
	d := &MatchDetail{
		match: wc.Match{
			ID: 1, Status: wc.StatusLive, Minute: &min,
			HomeTeam: wc.Team{Name: "Brazil", Flag: "🇧🇷"},
			AwayTeam: wc.Team{Name: "France", Flag: "🇫🇷"},
			HomeScorers: []wc.Scorer{
				{Name: "Vini Jr.", Minute: 12},
				{Name: "Richarlison", Minute: 58},
			},
			AwayScorers: []wc.Scorer{{Name: "Mbappé", Minute: 34, Penalty: true}},
			HomeScore: &hs, AwayScore: &as,
			Stage: "qf", Group: "",
		},
		allMatches: []wc.Match{},
	}
	log := d.buildEventLog()

	var labels []string
	for _, ev := range log {
		labels = append(labels, stripANSI(ev.label))
	}

	wantSubs := []string{
		"Kick-off",
		"Vini Jr.",
		"Mbappé",
		"Richarlison",
		"HT",
		"Second half underway",
	}
	for _, w := range wantSubs {
		found := false
		for _, l := range labels {
			if strings.Contains(l, w) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing %q in timeline: %v", w, labels)
		}
	}
}

func TestBuildEventLog_FinishedComeback(t *testing.T) {
	hs, as := 3, 2
	d := &MatchDetail{
		match: wc.Match{
			Status: wc.StatusFinished,
			HomeTeam: wc.Team{Name: "Spain", Flag: "🇪🇸"},
			AwayTeam: wc.Team{Name: "Germany", Flag: "🇩🇪"},
			HomeScorers: []wc.Scorer{
				{Name: "A", Minute: 70},
				{Name: "B", Minute: 82},
				{Name: "C", Minute: 88},
			},
			AwayScorers: []wc.Scorer{
				{Name: "X", Minute: 10},
				{Name: "Y", Minute: 40},
			},
			HomeScore: &hs, AwayScore: &as,
			Stage: "r16",
		},
	}
	log := d.buildEventLog()
	last := log[len(log)-1]
	if !strings.Contains(stripANSI(last.label), "Spain") || !strings.Contains(stripANSI(last.label), "win") {
		t.Fatalf("expected FT win note, got %q", last.label)
	}
}

func TestBuildEventLog_UpcomingPreview(t *testing.T) {
	hs, as := 1, 0
	d := &MatchDetail{
		match: wc.Match{
			Status: wc.StatusUpcoming,
			HomeTeam: wc.Team{Name: "Mexico", Flag: "🇲🇽", Group: "A"},
			AwayTeam: wc.Team{Name: "Poland", Flag: "🇵🇱", Group: "A"},
			Stage: "group", Group: "A", Matchday: 2,
		},
		allMatches: []wc.Match{{
			Status: wc.StatusFinished,
			HomeTeam: wc.Team{Name: "Mexico", Flag: "🇲🇽"},
			AwayTeam: wc.Team{Name: "Poland", Flag: "🇵🇱"},
			HomeScore: &hs, AwayScore: &as,
			HomeScorers: []wc.Scorer{{Name: "H", Minute: 20}},
		}},
	}
	log := d.buildEventLog()
	if len(log) == 0 {
		t.Fatal("expected preview events")
	}
	hasStakes := false
	for _, ev := range log {
		if strings.Contains(stripANSI(ev.label), "Group A") {
			hasStakes = true
		}
	}
	if !hasStakes {
		t.Fatal("expected group stakes in preview")
	}
	for _, ev := range log {
		if strings.Contains(ev.label, "Kick-off") && !strings.Contains(stripANSI(ev.label), "Kick-off 2") {
			// kick-off event for live match only
		}
	}
}

func TestGoalCommentary(t *testing.T) {
	g := timelineGoal{side: "home", homeScore: 1, awayScore: 0, minute: 8}
	if c := goalCommentary("A", "B", 0, 0, g); c != "Opens the scoring!" {
		t.Fatalf("got %q", c)
	}
	g2 := timelineGoal{side: "away", homeScore: 1, awayScore: 1, minute: 55}
	if c := goalCommentary("A", "B", 1, 0, g2); c != "Levels it up!" {
		t.Fatalf("got %q", c)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		if r == '\x1b' {
			skip = true
			continue
		}
		if skip {
			if r == 'm' {
				skip = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}