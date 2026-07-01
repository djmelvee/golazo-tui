package tz

import (
	_ "time/tzdata"
	"time"
)

// Amsterdam is Europe/Amsterdam (CET/CEST). All user-facing match times use this zone.
var Amsterdam = func() *time.Location {
	loc, err := time.LoadLocation("Europe/Amsterdam")
	if err != nil {
		return time.UTC
	}
	return loc
}()

// Now returns the current time in Amsterdam.
func Now() time.Time {
	return time.Now().In(Amsterdam)
}

// In converts t to Amsterdam local time.
func In(t time.Time) time.Time {
	return t.In(Amsterdam)
}

// ZoneAbbr returns CET or CEST for t when shown in Amsterdam.
func ZoneAbbr(t time.Time) string {
	return t.In(Amsterdam).Format("MST")
}

// FormatClock renders HH:MM in the active display timezone.
func FormatClock(t time.Time) string {
	return DisplayIn(t).Format("15:04 MST")
}

// FormatKickoff renders a full kickoff label in the active display timezone.
func FormatKickoff(t time.Time) string {
	return DisplayIn(t).Format("Mon 02 Jan 15:04 MST")
}

// FormatKickoffShort is the compact live-dashboard kickoff format.
func FormatKickoffShort(t time.Time) string {
	return DisplayIn(t).Format("Mon 02 Jan  15:04 MST")
}

// FormatDateHeader renders a fixtures date group header.
func FormatDateHeader(t time.Time) string {
	return DisplayIn(t).Format("Mon 02 Jan 2006")
}

// FormatDigestDate renders the digest newspaper date line.
func FormatDigestDate(t time.Time) string {
	return DisplayIn(t).Format("Monday 2 January 2006")
}

// StartOfDay returns midnight on t's calendar day in Amsterdam.
func StartOfDay(t time.Time) time.Time {
	at := t.In(Amsterdam)
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, Amsterdam)
}