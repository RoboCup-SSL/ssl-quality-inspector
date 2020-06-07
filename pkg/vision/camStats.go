package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"sort"
	"time"
)

const maxBallVel = 10.0
const maxBotVel = 6.0

type CamStats struct {
	FrameStats       *timing.FrameStats
	Robots           map[TeamColor][]*RobotStats
	Balls            []*ObjectStats
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
	statsConfig      StatsConfig
}

func NewCamStats(statsConfig StatsConfig) (s *CamStats) {
	s = new(CamStats)
	s.FrameStats = timing.NewFrameStats(statsConfig.TimeWindowQualityCam)
	s.Robots = map[TeamColor][]*RobotStats{}
	s.statsConfig = statsConfig
	s.TimingProcessing = timing.NewTiming(statsConfig.TimeWindowQualityCam)
	s.TimingReceiving = timing.NewTiming(statsConfig.TimeWindowQualityCam)

	return s
}

func (s CamStats) String() string {
	str := fmt.Sprint(s.FrameStats)
	str += fmt.Sprintf(" | %v blue | %v yellow | %v balls\n",
		colorizeByTeam(s.NumVisibleRobots(TeamBlue), TeamBlue),
		colorizeByTeam(s.NumVisibleRobots(TeamYellow), TeamYellow),
		len(s.Balls))
	str += fmt.Sprintf("Processing Time: %v\n Receiving Time: %v\n", s.TimingProcessing, s.TimingReceiving)

	str += "Balls: \n"
	for _, ball := range s.Balls {
		str += fmt.Sprintf("%v\n", ball)
	}
	str += "Robots: \n"

	for _, robot := range s.sortedRobotStats() {
		str += fmt.Sprintf("%v %v\n", robot.Id, robot)
	}
	return str
}

func (s *CamStats) Clear() {
	s.FrameStats.Clear()
	for teamColor := range s.Robots {
		for _, robot := range s.Robots[teamColor] {
			robot.Clear()
		}
	}
	for _, robot := range s.Balls {
		robot.Clear()
	}
}

func (s *CamStats) Merge() {
	s.Balls = mergeObjects(s.Balls)
	s.Robots[TeamYellow] = mergeRobots(s.Robots[TeamYellow])
	s.Robots[TeamBlue] = mergeRobots(s.Robots[TeamBlue])
}

func mergeRobots(objects []*RobotStats) (mergedRobots []*RobotStats) {
	for _, robot := range objects {
		if matchedRobot := matchingRobot(mergedRobots, robot); matchedRobot != nil {
			if matchedRobot.LastDetection.Time.Before(robot.LastDetection.Time) {
				*matchedRobot = *robot
			}
		} else {
			mergedRobots = append(mergedRobots, robot)
		}
	}
	return
}

func matchingRobot(mergedRobots []*RobotStats, robot *RobotStats) *RobotStats {
	for _, mergedRobot := range mergedRobots {
		if mergedRobot.LastDetection.Time == robot.LastDetection.Time {
			// Robots can not match if they were both updated at the same time
			continue
		}
		velocity := mergedRobot.Velocity(robot.LastDetection.Time, robot.LastDetection.Pos)
		if velocity < maxBotVel {
			return mergedRobot
		}
	}
	return nil
}

func mergeObjects(objects []*ObjectStats) (mergedBalls []*ObjectStats) {
	for _, ball := range objects {
		if matchedBall := matchingObject(mergedBalls, ball); matchedBall != nil {
			if matchedBall.LastDetection.Time.Before(ball.LastDetection.Time) {
				*matchedBall = *ball
			}
		} else {
			mergedBalls = append(mergedBalls, ball)
		}
	}
	return
}

func matchingObject(mergedBalls []*ObjectStats, ball *ObjectStats) *ObjectStats {
	for _, mergedBall := range mergedBalls {
		if mergedBall.LastDetection.Time == ball.LastDetection.Time {
			// Balls can not match if they were both updated at the same time
			continue
		}
		velocity := mergedBall.Velocity(ball.LastDetection.Time, ball.LastDetection.Pos)
		if velocity < maxBallVel {
			return mergedBall
		}
	}
	return nil
}

func (s *CamStats) Prune(tSent time.Time) {
	s.FrameStats.Prune(tSent.Add(-s.statsConfig.TimeWindowQualityCam))
	for teamColor := range s.Robots {
		var newRobots []*RobotStats
		for _, robot := range s.Robots[teamColor] {
			if tSent.Sub(robot.LastDetection.Time) < s.statsConfig.TimeWindowVisibility {
				newRobots = append(newRobots, robot)
				robot.Prune(tSent)
			}
		}
		s.Robots[teamColor] = newRobots
	}
	var newBalls []*ObjectStats
	for _, ball := range s.Balls {
		if tSent.Sub(ball.LastDetection.Time) < s.statsConfig.TimeWindowVisibility {
			newBalls = append(newBalls, ball)
			ball.Prune(tSent)
		}
	}
	s.Balls = newBalls
}

func (s *CamStats) NumVisibleRobots(teamColor TeamColor) int {
	numRobots := 0
	for _, robot := range s.Robots[teamColor] {
		if robot.FrameStats.Quality() > 0.5 {
			numRobots++
		}
	}
	return numRobots
}

func (s *CamStats) GetBallStats(tSent time.Time, newPos Position2d) (ballStats *ObjectStats) {
	minV := maxBallVel
	for _, ball := range s.Balls {
		if ball.LastDetection.Time == tSent {
			// already got a sample
			continue
		}
		v := ball.Velocity(tSent, newPos)
		if v < minV {
			ballStats = ball
			minV = v
		}
	}
	if ballStats == nil {
		ballStats = NewObjectStats(Detection{Pos: newPos, Time: tSent}, s.statsConfig.TimeWindowQualityBall)
		s.Balls = append(s.Balls, ballStats)
	}
	return
}

func (s *CamStats) GetRobotStats(robotId RobotId, tSent time.Time, robotPos Position2d) (robotStats *RobotStats) {
	minV := maxBotVel
	for _, robot := range s.Robots[robotId.Color] {
		if robot.LastDetection.Time == tSent {
			// already got a sample
			continue
		}
		v := robot.Velocity(tSent, robotPos)
		if v < minV {
			robotStats = robot
			minV = v
		}
	}
	if robotStats == nil {
		robotStats = new(RobotStats)
		*robotStats = NewRobotStats(robotId, Detection{Pos: robotPos, Time: tSent}, s.statsConfig.TimeWindowQualityRobot)
		s.Robots[robotId.Color] = append(s.Robots[robotId.Color], robotStats)
	}
	return
}

func (s *CamStats) sortedRobotStats() []*RobotStats {
	var robots []*RobotStats

	for teamColor := range s.Robots {
		for _, robot := range s.Robots[teamColor] {
			robots = append(robots, robot)
		}
	}
	sort.Slice(robots, func(i, j int) bool {
		if robots[i].Age() != robots[j].Age() {
			return robots[i].Age() > robots[j].Age()
		}
		if robots[i].Id.Color != robots[j].Id.Color {
			return robots[i].Id.Color < robots[j].Id.Color
		}
		return robots[i].Id.Id < robots[j].Id.Id
	})
	return robots
}
