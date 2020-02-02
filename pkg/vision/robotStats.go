package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"time"
)

type RobotStats struct {
	FrameStats *timing.FrameStats
	Visible    bool
}

func NewRobotStats(timeWindow time.Duration) (s RobotStats) {
	s.FrameStats = timing.NewFrameStats(timeWindow)

	return s
}

func (s RobotStats) String() string {
	return fmt.Sprintf("%v", s.FrameStats)
}
