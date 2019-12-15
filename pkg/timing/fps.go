package timing

import (
	"time"
)

type Fps struct {
	TimeWindow time.Duration
	durations  map[time.Time]struct{}
}

func NewFps(timeWindow time.Duration) (t Fps) {
	t.TimeWindow = timeWindow
	t.durations = map[time.Time]struct{}{}
	return t
}

func (f *Fps) Inc() {
	now := time.Now()
	f.durations[now] = struct{}{}

	lastValidMeasureTime := now.Add(-f.TimeWindow)
	var toBeDeleted []time.Time
	for measuredTime := range f.durations {
		if measuredTime.Before(lastValidMeasureTime) {
			toBeDeleted = append(toBeDeleted, measuredTime)
		}
	}
	for _, measuredTime := range toBeDeleted {
		delete(f.durations, measuredTime)
	}
}

func (f *Fps) Float32() float32 {
	return float32(len(f.durations)) / float32(f.TimeWindow.Seconds())
}
