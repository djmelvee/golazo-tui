package wc

import (
	"testing"
	"time"
)

func finishedMatch(id int, home, away Team, hs, as int, stage string) Match {
	return Match{
		ID: id, Stage: stage, Status: StatusFinished,
		HomeTeam: home, AwayTeam: away,
		HomeScore: &hs, AwayScore: &as,
		HomeScorers: []Scorer{{Name: "Striker", Minute: 23}},
	}
}

func upcomingMatch(id int, home, away Team, stage string, kick time.Time) Match {
	return Match{
		ID: id, Stage: stage, Status: StatusUpcoming,
		HomeTeam: home, AwayTeam: away,
		KickoffAt: kick,
	}
}

func TestBuildTeamForms(t *testing.T) {
	home := Team{ID: 1, Name: "Brazil"}
	away := Team{ID: 2, Name: "Serbia"}
	matches := []Match{
		finishedMatch(1, home, away, 2, 0, "group"),
		finishedMatch(2, home, Team{ID: 3, Name: "Switzerland"}, 1, 1, "group"),
	}

	forms := BuildTeamForms(matches)
	if forms[1] == nil || forms[1].Played != 2 {
		t.Fatalf("home form: %+v", forms[1])
	}
	if forms[1].GF != 3 || forms[1].GA != 1 {
		t.Fatalf("home GF/GA: %d/%d", forms[1].GF, forms[1].GA)
	}
}

func TestMatchIsPredictable(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	future := now.Add(2 * time.Hour)
	past := now.Add(-2 * time.Hour)

	if !MatchIsPredictable(upcomingMatch(1, Team{Name: "A"}, Team{Name: "B"}, "r32", future), now) {
		t.Fatal("expected predictable")
	}
	if MatchIsPredictable(upcomingMatch(2, Team{}, Team{Name: "B"}, "r16", future), now) {
		t.Fatal("expected TBD to be skipped")
	}
	if MatchIsPredictable(upcomingMatch(3, Team{Name: "A"}, Team{Name: "B"}, "r16", past), now) {
		t.Fatal("expected past match to be skipped")
	}
}

func TestUpcomingForPredictFiltersTBD(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	matches := []Match{
		upcomingMatch(1, Team{ID: 1, Name: "Spain"}, Team{ID: 2, Name: "Austria"}, "r32", now.Add(time.Hour)),
		upcomingMatch(2, Team{}, Team{}, "r16", now.Add(24*time.Hour)),
	}
	out := UpcomingForPredict(matches, now)
	if len(out) != 1 || out[0].ID != 1 {
		t.Fatalf("got %+v", out)
	}
}

func TestBuildPredictions(t *testing.T) {
	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	strong := Team{ID: 10, Name: "France"}
	weak := Team{ID: 11, Name: "Canada"}
	matches := []Match{
		finishedMatch(100, strong, weak, 3, 0, "group"),
		finishedMatch(101, strong, Team{ID: 12, Name: "Peru"}, 2, 1, "group"),
		upcomingMatch(200, strong, weak, "r16", now.Add(24*time.Hour)),
		upcomingMatch(201, Team{}, Team{}, "qf", now.Add(48*time.Hour)),
	}

	preds := BuildPredictions(matches, now)
	if len(preds) != 1 {
		t.Fatalf("expected 1 prediction, got %d", len(preds))
	}
	p := preds[0]
	if p.FTHome == 0 && p.FTAway == 0 {
		t.Fatalf("expected non-zero prediction, got 0-0 (xG %.1f-%.1f)", p.HomeXG, p.AwayXG)
	}
	if p.Summary == "" {
		t.Fatal("expected summary")
	}
	if len(p.TopScores) < 3 {
		t.Fatal("expected top scorelines")
	}
	if p.FTHome < p.FTAway {
		t.Fatalf("expected France to be favoured: %d-%d", p.FTHome, p.FTAway)
	}
	if len(p.Reasons) == 0 {
		t.Fatal("expected prediction reasons")
	}
	if p.FirstScorerMin <= 0 {
		t.Fatalf("first goal minute: %d", p.FirstScorerMin)
	}
	if len(p.Facts) == 0 {
		t.Fatal("expected form facts")
	}
}

func TestParseKickoffStadiumTZ(t *testing.T) {
	k := parseKickoffAt("07/02/2026 12:00", "Los Angeles")
	if k.IsZero() {
		t.Fatal("expected kickoff")
	}
	la, _ := time.LoadLocation("America/Los_Angeles")
	want := time.Date(2026, 7, 2, 12, 0, 0, 0, la).UTC()
	if !k.Equal(want) {
		t.Fatalf("got %v want %v", k, want)
	}
}