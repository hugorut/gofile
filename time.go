package filesystem

import "time"

type Time interface {
	Now() time.Time
}

// os time is a wrapper around the time function for the
// core os package
type OSTime struct{}

func (OSTime) Now() time.Time { return time.Now() }
