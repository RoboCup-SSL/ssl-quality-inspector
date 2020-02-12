package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"time"
)

type ObjectStats struct {
	FrameStats     *timing.FrameStats
	FirstDetection Detection
	LastDetection  Detection
	timeWindow     time.Duration
}

type Detection struct {
	Time time.Time
	Pos  Position2d
}

func NewObjectStats(detection Detection, timeWindow time.Duration) (s ObjectStats) {
	s.FrameStats = timing.NewFrameStats(timeWindow)
	s.FirstDetection = detection
	s.LastDetection = detection

	s.timeWindow = timeWindow

	return s
}

func (s ObjectStats) String() string {
	age := s.Age()
	age = time.Duration(age.Milliseconds() * 1_000_000)
	return fmt.Sprintf("%v | %v", s.FrameStats, age)
}

func (s *ObjectStats) Matches(t time.Time, pos Position2d, maxVel float64) bool {
	dt := t.Sub(s.LastDetection.Time).Seconds()
	ds := s.LastDetection.Pos.DistanceTo(pos)
	v := ds / dt
	return v < maxVel
}

func (s *ObjectStats) Add(tSent time.Time, frameId uint32, pos Position2d) {
	s.LastDetection.Time = tSent
	s.LastDetection.Pos = pos
	s.FrameStats.Add(frameId, tSent)
}

func (s *ObjectStats) Prune(tSent time.Time) {
	s.FrameStats.Prune(tSent.Add(-s.timeWindow))
}

func (s *ObjectStats) Clear() {
	s.FrameStats.Clear()
}

func (s *ObjectStats) Age() time.Duration {
	return s.LastDetection.Time.Sub(s.FirstDetection.Time)
}

func (s *ObjectStats) TimeSinceLastDetection(tLatest time.Time) time.Duration {
	return tLatest.Sub(s.LastDetection.Time)
}
