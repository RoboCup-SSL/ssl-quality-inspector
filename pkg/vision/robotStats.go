package vision

import (
	"fmt"
	"time"
)

type RobotStats struct {
	FrameStats *FrameStats
	Visible    bool
}

func NewRobotStats(timeWindow time.Duration) (s RobotStats) {
	s.FrameStats = NewFrameStats(timeWindow)

	return s
}

func (s RobotStats) String() string {
	return fmt.Sprintf("%v", s.FrameStats)
}
