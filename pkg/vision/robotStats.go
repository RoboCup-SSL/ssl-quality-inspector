package vision

import (
	"time"
)

type RobotStats struct {
	Id RobotId
	ObjectStats
}

func NewRobotStats(robotId RobotId, detection Detection, timeWindow time.Duration) (s RobotStats) {
	s.Id = robotId
	s.ObjectStats = NewObjectStats(detection, timeWindow)

	return s
}
