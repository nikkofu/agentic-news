package run

import "time"

func ShouldFallback(now time.Time) bool {
	deadlineGuard := time.Date(now.Year(), now.Month(), now.Day(), 6, 50, 0, 0, now.Location())
	return !now.Before(deadlineGuard)
}
