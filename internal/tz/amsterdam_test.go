package tz

import (
	"strings"
	"testing"
	"time"
)

func TestFormatKickoffUsesAmsterdam(t *testing.T) {
	utc := time.Date(2026, 7, 2, 10, 0, 0, 0, time.UTC)
	got := FormatKickoff(utc)
	if !strings.Contains(got, "12:00") {
		t.Fatalf("expected 12:00 Amsterdam in July (CEST), got %q", got)
	}
	if !strings.Contains(got, "CEST") && !strings.Contains(got, "CET") {
		t.Fatalf("expected zone label, got %q", got)
	}
}

func TestStartOfDayAmsterdam(t *testing.T) {
	utc := time.Date(2026, 7, 2, 22, 30, 0, 0, time.UTC)
	start := StartOfDay(utc)
	at := start.In(Amsterdam)
	if at.Hour() != 0 || at.Minute() != 0 {
		t.Fatalf("expected midnight Amsterdam, got %v", at)
	}
}