package data

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
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
