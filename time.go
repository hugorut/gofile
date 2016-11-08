package gofile

import "time"

type Time interface {
	Now() time.Time
}

// OSTime struct is a wrapper around the time function for the core os package
type OSTime struct{}

// Now returns the default time.Now
func (OSTime) Now() time.Time { return time.Now() }
