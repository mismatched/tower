package libtower

import "time"

// Time type
type Time struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
}

// Timeout type
type Timeout time.Duration
