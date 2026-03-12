package cronkit

import (
	"testing"
	"time"
)

func TestParseValidExpressions(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"* * * * *"},
		{"0 0 * * *"},
		{"*/15 * * * *"},
		{"0 9 1 * *"},
		{"30 4 1-7 * 1"},
		{"0,15,30,45 * * * *"},
		{"0 0 1,15 * *"},
	}
	for _, tc := range tests {
		s, err := Parse(tc.expr)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", tc.expr, err)
		}
		if s == nil {
			t.Errorf("Parse(%q) returned nil schedule", tc.expr)
		}
	}
}

func TestParseInvalidExpressions(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{""},
		{"* *"},
		{"* * * * * *"},
		{"60 * * * *"},
		{"* 24 * * *"},
		{"* * 0 * *"},
		{"* * * 13 *"},
		{"* * * * 7"},
		{"abc * * * *"},
		{"5-2 * * * *"},
	}
	for _, tc := range tests {
		_, err := Parse(tc.expr)
		if err == nil {
			t.Errorf("Parse(%q) expected error, got nil", tc.expr)
		}
	}
}

func TestParseFieldTypes(t *testing.T) {
	// Wildcard
	s, _ := Parse("* * * * *")
	if len(s.Minutes) != 60 {
		t.Errorf("expected 60 minutes for *, got %d", len(s.Minutes))
	}

	// Step
	s, _ = Parse("*/15 * * * *")
	expected := []int{0, 15, 30, 45}
	if len(s.Minutes) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, s.Minutes)
	}
	for i, v := range expected {
		if s.Minutes[i] != v {
			t.Errorf("Minutes[%d]: expected %d, got %d", i, v, s.Minutes[i])
		}
	}

	// Range
	s, _ = Parse("* * * * 1-5")
	if len(s.DaysOfWeek) != 5 {
		t.Errorf("expected 5 days for 1-5, got %d", len(s.DaysOfWeek))
	}

	// List
	s, _ = Parse("0,30 * * * *")
	if len(s.Minutes) != 2 || s.Minutes[0] != 0 || s.Minutes[1] != 30 {
		t.Errorf("expected [0,30], got %v", s.Minutes)
	}
}

func TestNextBasic(t *testing.T) {
	// Every hour at minute 0
	s, _ := Parse("0 * * * *")
	after := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	next := s.Next(after)

	if next.Hour() != 11 || next.Minute() != 0 {
		t.Errorf("expected 11:00, got %02d:%02d", next.Hour(), next.Minute())
	}
}

func TestNextEveryMinute(t *testing.T) {
	s, _ := Parse("* * * * *")
	after := time.Date(2026, 1, 1, 10, 30, 0, 0, time.UTC)
	next := s.Next(after)

	if next.Minute() != 31 {
		t.Errorf("expected minute 31, got %d", next.Minute())
	}
}

func TestNextSpecificTime(t *testing.T) {
	// 9:30 AM every day
	s, _ := Parse("30 9 * * *")
	after := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	next := s.Next(after)

	if next.Day() != 2 || next.Hour() != 9 || next.Minute() != 30 {
		t.Errorf("expected Jan 2 09:30, got %v", next)
	}
}

// Test POSIX cron OR semantics for DOM/DOW
func TestDomDowOrSemantics(t *testing.T) {
	// "0 0 15 * 1" = run at midnight on the 15th OR on Mondays
	s, err := Parse("0 0 15 * 1")
	if err != nil {
		t.Fatal(err)
	}

	// Monday that is NOT the 15th should match
	// 2026-01-05 is a Monday
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	if !s.matches(monday) {
		t.Error("expected Monday (non-15th) to match with OR semantics")
	}

	// The 15th that is NOT a Monday should also match
	// 2026-01-15 is a Thursday
	fifteenth := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !s.matches(fifteenth) {
		t.Error("expected 15th (non-Monday) to match with OR semantics")
	}

	// A day that is neither Monday nor the 15th should NOT match
	tuesday := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	if s.matches(tuesday) {
		t.Error("expected non-Monday, non-15th to NOT match")
	}
}

func TestDomWildcardDowRestricted(t *testing.T) {
	// "0 0 * * 1" = every Monday (DOM is wildcard, DOW restricted)
	s, _ := Parse("0 0 * * 1")

	// Monday should match
	monday := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	if !s.matches(monday) {
		t.Error("expected Monday to match")
	}

	// Tuesday should NOT match
	tuesday := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	if s.matches(tuesday) {
		t.Error("expected Tuesday to NOT match")
	}
}

func TestDomRestrictedDowWildcard(t *testing.T) {
	// "0 0 15 * *" = 15th of every month (DOM restricted, DOW wildcard)
	s, _ := Parse("0 0 15 * *")

	fifteenth := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !s.matches(fifteenth) {
		t.Error("expected 15th to match")
	}

	fourteenth := time.Date(2026, 1, 14, 0, 0, 0, 0, time.UTC)
	if s.matches(fourteenth) {
		t.Error("expected 14th to NOT match")
	}
}

func TestNextNoMatch(t *testing.T) {
	// Feb 31 never exists — should return zero time
	s, _ := Parse("0 0 31 2 *")
	after := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	next := s.Next(after)

	if !next.IsZero() {
		t.Errorf("expected zero time for impossible schedule, got %v", next)
	}
}
