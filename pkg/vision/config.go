package vision

import "time"

type StatsConfig struct {
	TimeWindowVisibility   time.Duration
	TimeWindowQualityCam   time.Duration
	TimeWindowQualityBall  time.Duration
	TimeWindowQualityRobot time.Duration
}
