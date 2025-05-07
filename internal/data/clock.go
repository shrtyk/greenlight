package data

import "time"

var MockTimeStamp = time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC)

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (r RealClock) Now() time.Time { return time.Now() }

type MockClock struct{}

func (r MockClock) Now() time.Time { return MockTimeStamp }
