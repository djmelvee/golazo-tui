package wc

import (
	"strings"
	"time"
)

// stadiumZone maps WC2026 host-city names to IANA timezones.
var stadiumZone = map[string]string{
	"Atlanta":          "America/New_York",
	"Boston":           "America/New_York",
	"Miami":            "America/New_York",
	"New York":         "America/New_York",
	"New York/New Jersey": "America/New_York",
	"Philadelphia":     "America/New_York",
	"Charlotte":        "America/New_York",
	"Washington":       "America/New_York",
	"Washington D.C.":  "America/New_York",
	"Dallas":           "America/Chicago",
	"Houston":          "America/Chicago",
	"Kansas City":      "America/Chicago",
	"Chicago":          "America/Chicago",
	"Minneapolis":      "America/Chicago",
	"Seattle":          "America/Los_Angeles",
	"Los Angeles":      "America/Los_Angeles",
	"San Francisco":    "America/Los_Angeles",
	"San Francisco Bay Area": "America/Los_Angeles",
	"Vancouver":        "America/Vancouver",
	"Toronto":          "America/Toronto",
	"Guadalajara":      "America/Mexico_City",
	"Mexico City":      "America/Mexico_City",
	"Monterrey":        "America/Monterrey",
}

func zoneForStadiumCity(city string) *time.Location {
	city = strings.TrimSpace(city)
	if city == "" {
		return kickoffDefaultLoc
	}
	if tz, ok := stadiumZone[city]; ok {
		if loc, err := time.LoadLocation(tz); err == nil {
			return loc
		}
	}
	for key, tz := range stadiumZone {
		if strings.Contains(city, key) || strings.Contains(key, city) {
			if loc, err := time.LoadLocation(tz); err == nil {
				return loc
			}
		}
	}
	return kickoffDefaultLoc
}

var kickoffDefaultLoc = func() *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// parseKickoffAt parses API local_date in the stadium timezone when possible.
func parseKickoffAt(localDate, stadiumCity string) time.Time {
	s := strings.TrimSpace(localDate)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t.UTC()
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.UTC()
	}

	loc := zoneForStadiumCity(stadiumCity)
	if t, err := time.ParseInLocation("01/02/2006 15:04", s, loc); err == nil {
		return t.UTC()
	}
	if t, err := time.ParseInLocation("01/02/2006 3:04 PM", s, loc); err == nil {
		return t.UTC()
	}
	return time.Time{}
}