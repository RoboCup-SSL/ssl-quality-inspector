package timing

import (
	"sync"
	"time"
)

type Fps struct {
	TimeWindow time.Duration
	durations  map[time.Time]struct{}
	mutex      sync.Mutex
}

func NewFps(timeWindow time.Duration) (t Fps) {
	t.TimeWindow = timeWindow
	t.durations = map[time.Time]struct{}{}
	return t
}

func (f *Fps) Inc() {
	f.mutex.Lock()
	f.durations[time.Now()] = struct{}{}
	f.mutex.Unlock()
}

func (f *Fps) Prune() {
	f.mutex.Lock()
	lastValidMeasureTime := time.Now().Add(-f.TimeWindow)
	for measuredTime := range f.durations {
		if measuredTime.Before(lastValidMeasureTime) {
			delete(f.durations, measuredTime)
		}
	}
	f.mutex.Unlock()
}

func (f *Fps) Clear() {
	f.mutex.Lock()
	f.durations = map[time.Time]struct{}{}
	f.mutex.Unlock()
}

func (f *Fps) Float32() float32 {
	f.mutex.Lock()
	numDurations := len(f.durations)
	f.mutex.Unlock()
	return float32(numDurations) / float32(f.TimeWindow.Seconds())
}
