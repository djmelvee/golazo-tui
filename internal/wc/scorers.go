package wc

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Scorer is a single goal from API scorer strings.
type Scorer struct {
	Name        string `json:"name"`
	Minute      int    `json:"minute"`
	InjuryTime  int    `json:"injury_time,omitempty"`
	Team        string `json:"team,omitempty"`
	Penalty     bool   `json:"penalty,omitempty"`
	OwnGoal     bool   `json:"own_goal,omitempty"`
}

// ScorerRow is an aggregated top-scorer entry.
type ScorerRow struct {
	Name  string
	Team  string
	Flag  string
	Goals int
}

var scorerMinuteRE = regexp.MustCompile(`(?i)^(.+?)\s+(\d+)(?:\+(\d+))?'?(?:\s*\(p\))?$`)

// ParseScorers parses API home_scorers/away_scorers blobs into Scorer slices.
func ParseScorers(raw, teamName string) []Scorer {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") {
		return nil
	}

	var entries []string
	if strings.HasPrefix(raw, "[") {
		_ = json.Unmarshal([]byte(raw), &entries)
	}
	if len(entries) == 0 {
		entries = splitScorerBlob(raw)
	}

	var out []Scorer
	for _, e := range entries {
		e = strings.TrimSpace(e)
		e = strings.Trim(e, `"'`)
		if e == "" || strings.EqualFold(e, "null") {
			continue
		}
		ownGoal := strings.Contains(e, "(OG)")
		if ownGoal {
			e = strings.ReplaceAll(e, "(OG)", "")
			e = strings.TrimSpace(e)
		}
		if m := scorerMinuteRE.FindStringSubmatch(e); len(m) >= 3 {
			min, _ := strconv.Atoi(strings.TrimSpace(m[2]))
			injury := 0
			if len(m) >= 4 && m[3] != "" {
				injury, _ = strconv.Atoi(m[3])
			}
			penalty := strings.Contains(strings.ToLower(e), "(p)")
			out = append(out, Scorer{
				Name:       strings.TrimSpace(m[1]),
				Minute:     min,
				InjuryTime: injury,
				Team:       teamName,
				Penalty:    penalty,
				OwnGoal:    ownGoal,
			})
		}
	}
	return out
}

func splitScorerBlob(s string) []string {
	s = strings.Trim(s, "{}")
	if s == "" {
		return nil
	}
	var parts []string
	var cur strings.Builder
	inQuote := false
	for _, r := range s {
		switch r {
		case '"', '\u201c', '\u201d':
			inQuote = !inQuote
		case ',':
			if !inQuote {
				parts = append(parts, cur.String())
				cur.Reset()
				continue
			}
		}
		cur.WriteRune(r)
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// BuildTopScorers aggregates goals from finished matches' scorer lists.
func BuildTopScorers(matches []Match) []ScorerRow {
	type key struct{ name, team string }
	counts := make(map[key]int)
	flags := make(map[key]string)

	for _, m := range matches {
		if m.HomeScore == nil && m.AwayScore == nil {
			continue
		}
		for _, s := range m.HomeScorers {
			k := key{name: s.Name, team: m.HomeTeam.Name}
			counts[k]++
			flags[k] = m.HomeTeam.Flag
		}
		for _, s := range m.AwayScorers {
			k := key{name: s.Name, team: m.AwayTeam.Name}
			counts[k]++
			flags[k] = m.AwayTeam.Flag
		}
	}

	rows := make([]ScorerRow, 0, len(counts))
	for k, g := range counts {
		rows = append(rows, ScorerRow{Name: k.name, Team: k.team, Flag: flags[k], Goals: g})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Goals != rows[j].Goals {
			return rows[i].Goals > rows[j].Goals
		}
		if rows[i].Name != rows[j].Name {
			return rows[i].Name < rows[j].Name
		}
		return rows[i].Team < rows[j].Team
	})
	return rows
}

// ScorerNameForGoal returns the nth scorer name (1-based) for a side, if known.
func ScorerNameForGoal(m Match, side string, goalNum int) string {
	var list []Scorer
	if side == "home" {
		list = m.HomeScorers
	} else {
		list = m.AwayScorers
	}
	if goalNum <= 0 || goalNum > len(list) {
		return ""
	}
	return list[goalNum-1].Name
}