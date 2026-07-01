package wc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const gamesFixture = `{"games":[
{"id":"1","home_team_id":"1","away_team_id":"2","home_score":"2","away_score":"0",
 "home_team_name_en":"Mexico","away_team_name_en":"South Africa","group":"A","matchday":"1",
 "local_date":"06/11/2026 13:00","stadium_id":"1","finished":"TRUE","time_elapsed":"finished","type":"group"},
{"id":"81","home_team_id":"25","away_team_id":"34","home_score":"0","away_score":"2",
 "home_team_name_en":"Belgium","away_team_name_en":"Senegal","group":"","matchday":"0",
 "local_date":"07/01/2026 13:00","stadium_id":"5","finished":"FALSE","time_elapsed":"live","type":"round_of_32"},
{"id":"82","home_team_id":"13","away_team_id":"6","home_score":"null","away_score":"null",
 "home_team_name_en":"United States","away_team_name_en":"Bosnia and Herzegovina","group":"","matchday":"0",
 "local_date":"12/31/2026 17:00","stadium_id":"16","finished":"FALSE","time_elapsed":"notstarted","type":"round_of_32"}
]}`

func TestDecodeGames(t *testing.T) {
	games, err := decodeGames([]byte(gamesFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 3 {
		t.Fatalf("got %d games, want 3", len(games))
	}

	if !games[0].Finished || games[0].HomeScore != 2 || games[0].AwayScore != 0 {
		t.Fatalf("finished game: %+v", games[0])
	}
	if deriveStatus(games[0]) != StatusFinished {
		t.Fatalf("game 0 status: %s", deriveStatus(games[0]))
	}

	if !games[1].IsLive || games[1].Finished || games[1].HomeScore != 0 || games[1].AwayScore != 2 {
		t.Fatalf("live game: %+v", games[1])
	}
	if deriveStatus(games[1]) != StatusLive {
		t.Fatalf("game 1 status: %s", deriveStatus(games[1]))
	}

	if games[2].IsLive || games[2].Finished {
		t.Fatalf("upcoming game should not be live/finished: %+v", games[2])
	}
	if deriveStatus(games[2]) != StatusUpcoming {
		t.Fatalf("game 2 status: %s", deriveStatus(games[2]))
	}
}

func TestParseKickoffUSFormat(t *testing.T) {
	k := parseKickoffAt("07/01/2026 13:00", "Miami")
	if k.IsZero() {
		t.Fatal("expected parsed kickoff")
	}
	loc, _ := time.LoadLocation("America/New_York")
	want := time.Date(2026, 7, 1, 13, 0, 0, 0, loc).UTC()
	if !k.Equal(want) {
		t.Fatalf("got %v want %v", k, want)
	}
}

func TestFetchMatchesFromMockAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/get/games":
			w.Write([]byte(gamesFixture))
		case "/get/stadiums":
			w.Write([]byte(`{"stadiums":[{"id":"5","name_en":"Stadium","city_en":"City"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := New(srv.URL, "")
	live, err := client.FetchMatches(context.Background(), string(StatusLive))
	if err != nil {
		t.Fatal(err)
	}
	if len(live) != 1 {
		t.Fatalf("live matches: got %d want 1", len(live))
	}
	if live[0].HomeTeam.Name != "Belgium" || live[0].AwayScore == nil || *live[0].AwayScore != 2 {
		t.Fatalf("live match: %+v", live[0])
	}

	finished, err := client.FetchMatches(context.Background(), string(StatusFinished))
	if err != nil {
		t.Fatal(err)
	}
	if len(finished) != 1 {
		t.Fatalf("finished matches: got %d want 1", len(finished))
	}

	upcoming, err := client.FetchMatches(context.Background(), string(StatusUpcoming))
	if err != nil {
		t.Fatal(err)
	}
	if len(upcoming) != 1 {
		t.Fatalf("upcoming matches: got %d want 1", len(upcoming))
	}
}

func TestDecodeLiveAPIFixture(t *testing.T) {
	raw, err := os.ReadFile("../../tmp_games.json")
	if err != nil {
		t.Skip("tmp_games.json not present; run curl against live API first")
	}
	games, err := decodeGames(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 104 {
		t.Fatalf("got %d games, want 104", len(games))
	}
}

func TestTeamFlagAliases(t *testing.T) {
	if teamFlag("Turkey") == "" {
		t.Fatal("expected Turkey alias flag")
	}
	if teamFlag("Ivory Coast") == "" {
		t.Fatal("expected Ivory Coast alias flag")
	}
}