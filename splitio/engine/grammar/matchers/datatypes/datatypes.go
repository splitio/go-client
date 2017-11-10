package datatypes

import (
	"math"
	"time"
)

const (
	// Number data type
	Number = "NUMBER"
	// Datetime data type
	Datetime = "DATETIME"
)

// ZeroTimeTS Takes a timestamp in milliseconds as a parameter and
// returns another timestamp in seconds with the same date and zero time.
func ZeroTimeTS(ts int64) int64 {
	ns := int64(math.Mod(float64(ts), float64(1000))) * 1000000
	t := time.Unix(ts/1000, ns).UTC() // Timestamp is converted from milliseconds to seconds
	rounded := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	return rounded.Unix()
}

// ZeroSecondsTS Takes a timestamp in milliseconds as a parameter and
// returns another timestamp in seconds with the same date & time but zero seconds.
func ZeroSecondsTS(ts int64) int64 {
	ns := int64(math.Mod(float64(ts), float64(1000))) * 1000000
	t := time.Unix(ts/1000, ns).UTC() // Timestamp is converted from milliseconds to seconds
	rounded := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, time.UTC)
	return rounded.Unix()
}
