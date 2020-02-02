package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"time"
)

type RobotStats struct {
	FrameStats *timing.FrameStats
	timeWindow time.Duration
	Visible    bool
}

func NewRobotStats(timeWindow time.Duration) (s RobotStats) {
	s.FrameStats = timing.NewFrameStats(timeWindow)
	s.timeWindow = timeWindow

	return s
}

func (s RobotStats) String() string {
	return fmt.Sprintf("%v", s.FrameStats)
}

func (s *RobotStats) Add(tSent time.Time, frameId uint32) {
	s.FrameStats.Add(frameId, tSent)
	s.FrameStats.Fps.Inc()
}

func (s *RobotStats) Clear() {
	s.FrameStats.Clear()
}

func (s *RobotStats) Prune(tSent time.Time) {
	s.FrameStats.Prune(tSent.Add(-s.timeWindow))
}
