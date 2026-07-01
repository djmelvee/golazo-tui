package wc

import "testing"

func TestParseScorers(t *testing.T) {
	raw := `{"I.B. Hwang 67'","H.G. Oh 80'"}`
	got := ParseScorers(raw, "South Korea")
	if len(got) != 2 {
		t.Fatalf("got %d scorers, want 2", len(got))
	}
	if got[0].Name == "" || got[0].Minute != 67 {
		t.Fatalf("first scorer: %+v", got[0])
	}
}

func TestBuildTopScorers(t *testing.T) {
	hs, as := 2, 1
	matches := []Match{{
		Status: StatusFinished,
		HomeTeam: Team{Name: "Mexico", Flag: "🇲🇽"},
		AwayTeam: Team{Name: "South Africa", Flag: "🇿🇦"},
		HomeScorers: []Scorer{{Name: "A", Minute: 9}, {Name: "A", Minute: 67}},
		AwayScorers: []Scorer{{Name: "B", Minute: 40}},
		HomeScore: &hs, AwayScore: &as,
	}}
	rows := BuildTopScorers(matches)
	if len(rows) != 2 || rows[0].Goals != 2 || rows[0].Name != "A" {
		t.Fatalf("rows: %+v", rows)
	}
}