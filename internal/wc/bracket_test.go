package wc

import "testing"

func TestBuildBracket(t *testing.T) {
	hs, as := 2, 1
	matches := []Match{{
		ID: 81, Stage: "r32", Status: StatusFinished,
		HomeTeam: Team{Name: "England"}, AwayTeam: Team{Name: "DR Congo"},
		HomeScore: &hs, AwayScore: &as,
	}}
	rounds := BuildBracket(matches)
	if len(rounds) != 1 || len(rounds[0].Slots) != 1 {
		t.Fatalf("rounds: %+v", rounds)
	}
	if rounds[0].Slots[0].Winner != "England" {
		t.Fatalf("winner: %s", rounds[0].Slots[0].Winner)
	}
}

func TestNormalizeStage(t *testing.T) {
	if NormalizeStage("round_of_32") != "r32" {
		t.Fatal("expected r32")
	}
}