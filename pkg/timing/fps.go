package timing

import (
	"sync"
	"time"
)

type Fps struct {
	timeWindow time.Duration
	durations  map[time.Time]struct{}
	mutex      sync.Mutex
}

func NewFps(timeWindow time.Duration) (t *Fps) {
	t = new(Fps)
	t.timeWindow = timeWindow
	t.durations = map[time.Time]struct{}{}
	return t
}

func (f *Fps) Inc() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	now := time.Now()
	f.durations[now] = struct{}{}
	f.prune(now)
}

func (f *Fps) prune(now time.Time) {
	lastValidMeasureTime := now.Add(-f.timeWindow)
	for measuredTime := range f.durations {
		if measuredTime.Before(lastValidMeasureTime) {
			delete(f.durations, measuredTime)
		}
	}
}

func (f *Fps) Clear() {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.durations = map[time.Time]struct{}{}
}

func (f *Fps) Float32() float32 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	numDurations := len(f.durations)
	return float32(numDurations) / float32(f.timeWindow.Seconds())
}
