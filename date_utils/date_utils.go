package date_utils

import "time"

// GetCurrentTime returns the current time as a pointer to time.Time
func GetCurrentTime() *time.Time {
	now := time.Now()
	return &now
}

// GetTimeDiffInMillis returns the difference between two times in milliseconds
func GetTimeDiffInMillis(t time.Time) int64 {
	start := t.UnixNano() / int64(time.Millisecond)
	end := time.Now().UnixNano() / int64(time.Millisecond)
	diff := end - start
	return diff
}
