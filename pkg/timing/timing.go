package timing

import (
	"fmt"
	"sort"
	"time"
)

type Timing struct {
	TimeWindow time.Duration
	Min        time.Duration
	Max        time.Duration
	Avg        time.Duration
	Median     time.Duration
	durations  map[time.Time]time.Duration
}

func NewTiming(timeWindow time.Duration) (t Timing) {
	t.TimeWindow = timeWindow
	t.durations = map[time.Time]time.Duration{}
	return t
}

func (t *Timing) Clear() {
	t.durations = map[time.Time]time.Duration{}
}

func (t *Timing) Add(duration time.Duration) {
	now := time.Now()
	t.durations[now] = duration

	lastValidMeasureTime := now.Add(-t.TimeWindow)
	var toBeDeleted []time.Time
	for measuredTime := range t.durations {
		if measuredTime.Before(lastValidMeasureTime) {
			toBeDeleted = append(toBeDeleted, measuredTime)
		}
	}
	for _, measuredTime := range toBeDeleted {
		delete(t.durations, measuredTime)
	}

	sortedDurations := t.sortedDurations()
	t.Min = sortedDurations[0]
	t.Max = sortedDurations[len(sortedDurations)-1]
	t.Avg = t.calcAvg()
	t.Median = sortedDurations[len(sortedDurations)/2]
}

func (t *Timing) calcAvg() time.Duration {
	sum := time.Duration(0)
	for _, d := range t.durations {
		sum += d
	}
	return time.Duration(sum.Nanoseconds() / int64(len(t.durations)))
}

func (t *Timing) sortedDurations() []time.Duration {
	durations := make([]time.Duration, len(t.durations))
	i := 0
	for _, d := range t.durations {
		durations[i] = d
		i++
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	return durations
}

func (t Timing) String() string {
	return fmt.Sprintf("Min: %10v Max: %10v Avg: %10v Median: %10v (%v measures in %v)", t.Min, t.Max, t.Avg, t.Median, len(t.durations), t.TimeWindow)
}
