package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"time"
)

const maxBallVel = 10

type BallStats struct {
	FrameStats    *timing.FrameStats
	Visible       bool
	LastDetection Detection
}

type Detection struct {
	Time time.Time
	Pos  Position2d
}

func NewBallStats(timeWindow time.Duration, detection Detection) (s BallStats) {
	s.FrameStats = timing.NewFrameStats(timeWindow)
	s.LastDetection = detection
	s.Visible = false

	return s
}

func (s BallStats) String() string {
	return fmt.Sprintf("%v", s.FrameStats)
}

func (s *BallStats) Matches(t time.Time, pos Position2d) bool {
	dt := t.Sub(s.LastDetection.Time).Seconds()
	ds := s.LastDetection.Pos.DistanceTo(pos)
	v := ds / dt
	return v < maxBallVel
}

func (s *BallStats) Add(tSent time.Time, pos Position2d, frameId uint32) {
	s.LastDetection.Time = tSent
	s.LastDetection.Pos = pos
	s.FrameStats.Add(frameId, tSent)
	s.FrameStats.Fps.Inc()
}
