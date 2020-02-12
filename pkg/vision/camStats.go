package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"sort"
	"time"
)

const maxBallVel = 10
const maxBotVel = 6

type CamStats struct {
	FrameStats       *timing.FrameStats
	Robots           map[TeamColor][]*RobotStats
	Balls            []*ObjectStats
	timeWindow       time.Duration
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
}

func NewCamStats(timeWindow time.Duration) (s CamStats) {

	s.FrameStats = timing.NewFrameStats(timeWindow)
	s.Robots = map[TeamColor][]*RobotStats{}
	s.timeWindow = timeWindow
	s.TimingProcessing = new(timing.Timing)
	s.TimingReceiving = new(timing.Timing)

	*s.TimingProcessing = timing.NewTiming(timeWindow)
	*s.TimingReceiving = timing.NewTiming(timeWindow)

	return s
}

func (s CamStats) String() string {
	str := fmt.Sprint(s.FrameStats)
	str += fmt.Sprintf(" | %v blue | %v yellow | %v balls\n",
		colorizeByTeam(s.NumVisibleRobots(TeamBlue), TeamBlue),
		colorizeByTeam(s.NumVisibleRobots(TeamYellow), TeamYellow),
		len(s.Balls))
	str += fmt.Sprintf("Processing Time: %v\n Receiving Time: %v\n", s.TimingProcessing, s.TimingReceiving)

	str += "Balls: "
	for _, ball := range s.Balls {
		str += fmt.Sprintf("%v  ", ball)
	}
	str += "\n"

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

func (s *CamStats) Prune(tSent time.Time) {
	s.FrameStats.Prune(tSent.Add(-s.timeWindow))
	for teamColor := range s.Robots {
		var newRobots []*RobotStats
		for _, robot := range s.Robots[teamColor] {
			if tSent.Sub(robot.LastDetection.Time).Seconds() < 1 {
				newRobots = append(newRobots, robot)
			}
			robot.Prune(tSent)
		}
		s.Robots[teamColor] = newRobots
	}
	var newBalls []*ObjectStats
	for _, ball := range s.Balls {
		if tSent.Sub(ball.LastDetection.Time).Seconds() < 1 {
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

func (s *CamStats) GetBallStats(tSent time.Time, newPos Position2d) *ObjectStats {
	for _, ball := range s.Balls {
		if ball.Matches(tSent, newPos, maxBallVel) {
			return ball
		}
	}
	stats := NewObjectStats(Detection{Pos: newPos, Time: tSent}, s.timeWindow)
	s.Balls = append(s.Balls, &stats)
	return &stats
}

func (s *CamStats) GetRobotStats(id RobotId, tSent time.Time, newPos Position2d) *RobotStats {
	for _, robot := range s.Robots[id.Color] {
		if robot.Matches(tSent, newPos, maxBotVel) {
			return robot
		}
	}
	robot := NewRobotStats(id, Detection{Pos: newPos, Time: tSent}, s.timeWindow)
	s.Robots[id.Color] = append(s.Robots[id.Color], &robot)
	return &robot
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
