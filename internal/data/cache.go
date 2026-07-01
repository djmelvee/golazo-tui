package data

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/djmelvee/golazo-tui/internal/wc"
	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS kv (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TEXT NOT NULL
);`

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func OpenRO(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Set(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO kv (key, value, updated_at) VALUES (?, ?, ?)
         ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		key, string(b), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (s *Store) Get(key string, dest any) (time.Time, error) {
	var val, updatedAt string
	err := s.db.QueryRow(`SELECT value, updated_at FROM kv WHERE key = ?`, key).
		Scan(&val, &updatedAt)
	if err != nil {
		return time.Time{}, err
	}
	t, _ := time.Parse(time.RFC3339, updatedAt)
	return t, json.Unmarshal([]byte(val), dest)
}

func (s *Store) LastUpdated(key string) time.Time {
	var updatedAt string
	err := s.db.QueryRow(`SELECT updated_at FROM kv WHERE key = ?`, key).Scan(&updatedAt)
	if err != nil {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, updatedAt)
	return t
}

func (s *Store) LiveMatches() []wc.Match {
	var out []wc.Match
	s.Get("matches:live", &out) //nolint:errcheck
	return out
}

func (s *Store) UpcomingMatches() []wc.Match {
	var out []wc.Match
	s.Get("matches:upcoming", &out) //nolint:errcheck
	return out
}

// FutureMatches returns upcoming fixtures whose kickoff is still in the future
// and both teams are known (filters TBD bracket placeholders).
func (s *Store) FutureMatches() []wc.Match {
	now := time.Now()
	var out []wc.Match
	for _, m := range s.UpcomingMatches() {
		if !wc.MatchIsPredictable(m, now) {
			continue
		}
		out = append(out, m)
	}
	return out
}

func (s *Store) FinishedMatches() []wc.Match {
	var out []wc.Match
	s.Get("matches:finished", &out) //nolint:errcheck
	return out
}

func (s *Store) Standings() map[string][]wc.GroupRow {
	out := make(map[string][]wc.GroupRow)
	s.Get("standings", &out) //nolint:errcheck
	return out
}

func (s *Store) GetEvents(matchID int) []wc.GoalEvent {
	var out []wc.GoalEvent
	s.Get(fmt.Sprintf("events:%d", matchID), &out) //nolint:errcheck
	return out
}

func (s *Store) SetEvents(matchID int, events []wc.GoalEvent) error {
	return s.Set(fmt.Sprintf("events:%d", matchID), events)
}

type lastScore struct {
	H int `json:"h"`
	A int `json:"a"`
}

// SetLastScore caches the most-recently-seen score for a match.
// Called by the fetcher whenever a live match has a non-zero score, so that
// the TUI can show it even after the match transitions to FT in time-promotion.
func (s *Store) SetLastScore(matchID, home, away int) error {
	return s.Set(fmt.Sprintf("lastscore:%d", matchID), lastScore{H: home, A: away})
}

// GetLastScore returns the last cached score for a match, if any.
func (s *Store) GetLastScore(matchID int) (home, away int, ok bool) {
	var ls lastScore
	if _, err := s.Get(fmt.Sprintf("lastscore:%d", matchID), &ls); err != nil {
		return 0, 0, false
	}
	return ls.H, ls.A, true
}

func (s *Store) GetToken() string {
	var token string
	s.Get("auth:token", &token) //nolint:errcheck
	return token
}

func (s *Store) SetToken(token string) error {
	return s.Set("auth:token", token)
}

// AllMatches returns every cached match across live, upcoming, and finished.
func (s *Store) AllMatches() []wc.Match {
	out := make([]wc.Match, 0,
		len(s.LiveMatches())+len(s.UpcomingMatches())+len(s.FinishedMatches()))
	out = append(out, s.LiveMatches()...)
	out = append(out, s.UpcomingMatches()...)
	out = append(out, s.FinishedMatches()...)
	return out
}

// FindMatch returns a match by ID from any bucket.
func (s *Store) FindMatch(id int) *wc.Match {
	for _, m := range s.AllMatches() {
		if m.ID == id {
			cp := m
			return &cp
		}
	}
	return nil
}

func (s *Store) GetPrefBool(key string, defaultVal bool) bool {
	var v bool
	if _, err := s.Get("prefs:"+key, &v); err != nil {
		return defaultVal
	}
	return v
}

func (s *Store) SetPrefBool(key string, val bool) error {
	return s.Set("prefs:"+key, val)
}

func (s *Store) GetPrefString(key, defaultVal string) string {
	var v string
	if _, err := s.Get("prefs:"+key, &v); err != nil || v == "" {
		return defaultVal
	}
	return v
}

func (s *Store) SetPrefString(key, val string) error {
	return s.Set("prefs:"+key, val)
}

// GetOK reads a key and reports whether it existed.
func (s *Store) GetOK(key string, dest any) (time.Time, bool) {
	t, err := s.Get(key, dest)
	return t, err == nil
}

// IsFresh reports whether key was updated within maxAge.
func (s *Store) IsFresh(key string, maxAge time.Duration) bool {
	t := s.LastUpdated(key)
	if t.IsZero() {
		return false
	}
	return time.Since(t) <= maxAge
}

// BatchWriter applies multiple Set operations in one transaction.
type BatchWriter struct {
	store *Store
	tx    *sql.Tx
}

func (s *Store) BeginBatch() (*BatchWriter, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	return &BatchWriter{store: s, tx: tx}, nil
}

func (b *BatchWriter) Set(key string, v any) error {
	bts, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = b.tx.Exec(
		`INSERT INTO kv (key, value, updated_at) VALUES (?, ?, ?)
         ON CONFLICT(key) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`,
		key, string(bts), time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func (b *BatchWriter) Commit() error {
	return b.tx.Commit()
}

func (b *BatchWriter) Rollback() error {
	return b.tx.Rollback()
}

type predictionsCache struct {
	UpdatedAt time.Time            `json:"updated_at"`
	Preds     []wc.MatchPrediction `json:"preds"`
}

// GetPredictions returns cached predictions when still valid for the match buckets.
func (s *Store) GetPredictions() ([]wc.MatchPrediction, bool) {
	var pc predictionsCache
	if _, err := s.Get("predictions:cache", &pc); err != nil {
		return nil, false
	}
	liveT := s.LastUpdated("matches:live")
	upT := s.LastUpdated("matches:upcoming")
	ftT := s.LastUpdated("matches:finished")
	if pc.UpdatedAt.Before(liveT) || pc.UpdatedAt.Before(upT) || pc.UpdatedAt.Before(ftT) {
		return nil, false
	}
	return pc.Preds, true
}

func (s *Store) SetPredictions(preds []wc.MatchPrediction) error {
	pc := predictionsCache{
		UpdatedAt: time.Now().UTC(),
		Preds:     preds,
	}
	return s.Set("predictions:cache", pc)
}

// RecentGoal is a goal event with match context for the goal history panel.
type RecentGoal struct {
	Goal      wc.GoalEvent `json:"goal"`
	HomeTeam  string       `json:"home_team"`
	AwayTeam  string       `json:"away_team"`
	HomeFlag  string       `json:"home_flag"`
	AwayFlag  string       `json:"away_flag"`
}

// RecentGoals returns the most recent goal events across all matches.
func (s *Store) RecentGoals(limit int) []RecentGoal {
	if limit <= 0 {
		limit = 10
	}
	byID := make(map[int]wc.Match)
	for _, m := range s.AllMatches() {
		byID[m.ID] = m
	}
	var all []RecentGoal
	for id, m := range byID {
		for _, ev := range s.GetEvents(id) {
			all = append(all, RecentGoal{
				Goal: ev, HomeTeam: m.HomeTeam.Name, AwayTeam: m.AwayTeam.Name,
				HomeFlag: m.HomeTeam.Flag, AwayFlag: m.AwayTeam.Flag,
			})
		}
	}
	sortRecentGoals(all)
	if len(all) > limit {
		all = all[:limit]
	}
	return all
}

func sortRecentGoals(goals []RecentGoal) {
	sort.Slice(goals, func(i, j int) bool {
		return goals[i].Goal.DetectedAt.After(goals[j].Goal.DetectedAt)
	})
}
