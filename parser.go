// Package cronkit provides a cron expression parser and job scheduler.
package cronkit

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule represents a parsed cron expression.
type Schedule struct {
	Minutes    []int
	Hours      []int
	DaysOfMonth []int
	Months     []int
	DaysOfWeek []int
}

// Parse parses a standard 5-field cron expression.
// Fields: minute hour day-of-month month day-of-week
func Parse(expr string) (*Schedule, error) {
	fields := strings.Fields(expr)
	if len(fields) != 5 {
		return nil, fmt.Errorf("cronkit: expected 5 fields, got %d", len(fields))
	}

	minutes, err := parseField(fields[0], 0, 59)
	if err != nil {
		return nil, fmt.Errorf("cronkit: minute field: %w", err)
	}
	hours, err := parseField(fields[1], 0, 23)
	if err != nil {
		return nil, fmt.Errorf("cronkit: hour field: %w", err)
	}
	dom, err := parseField(fields[2], 1, 31)
	if err != nil {
		return nil, fmt.Errorf("cronkit: day-of-month field: %w", err)
	}
	months, err := parseField(fields[3], 1, 12)
	if err != nil {
		return nil, fmt.Errorf("cronkit: month field: %w", err)
	}
	dow, err := parseField(fields[4], 0, 6)
	if err != nil {
		return nil, fmt.Errorf("cronkit: day-of-week field: %w", err)
	}

	return &Schedule{
		Minutes:     minutes,
		Hours:       hours,
		DaysOfMonth: dom,
		Months:      months,
		DaysOfWeek:  dow,
	}, nil
}

// Next returns the next time after t that matches the schedule.
func (s *Schedule) Next(after time.Time) time.Time {
	t := after.Add(time.Minute).Truncate(time.Minute)

	for i := 0; i < 525960; i++ { // max ~1 year of minutes
		if s.matches(t) {
			return t
		}
		t = t.Add(time.Minute)
	}

	return time.Time{}
}

func (s *Schedule) matches(t time.Time) bool {
	return contains(s.Minutes, t.Minute()) &&
		contains(s.Hours, t.Hour()) &&
		contains(s.DaysOfMonth, t.Day()) &&
		contains(s.Months, int(t.Month())) &&
		contains(s.DaysOfWeek, int(t.Weekday()))
}

func contains(vals []int, v int) bool {
	for _, val := range vals {
		if val == v {
			return true
		}
	}
	return false
}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		return makeRange(min, max, 1), nil
	}

	// Handle */n
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(field[2:])
		if err != nil || step <= 0 {
			return nil, fmt.Errorf("invalid step: %s", field)
		}
		return makeRange(min, max, step), nil
	}

	var result []int
	for _, part := range strings.Split(field, ",") {
		part = strings.TrimSpace(part)

		// Handle range (e.g., 1-5)
		if strings.Contains(part, "-") {
			rangeParts := strings.SplitN(part, "-", 2)
			lo, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", rangeParts[0])
			}
			hi, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", rangeParts[1])
			}
			if lo < min || hi > max || lo > hi {
				return nil, fmt.Errorf("range %d-%d out of bounds [%d-%d]", lo, hi, min, max)
			}
			result = append(result, makeRange(lo, hi, 1)...)
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid value: %s", part)
			}
			if n < min || n > max {
				return nil, fmt.Errorf("value %d out of bounds [%d-%d]", n, min, max)
			}
			result = append(result, n)
		}
	}

	return result, nil
}

func makeRange(min, max, step int) []int {
	var result []int
	for i := min; i <= max; i += step {
		result = append(result, i)
	}
	return result
}
