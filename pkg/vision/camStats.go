package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"sort"
	"time"
)

type CamStats struct {
	FrameStats       *timing.FrameStats
	Robots           map[RobotId]*RobotStats
	Balls            []*BallStats
	timeWindow       time.Duration
	maxBotId         int
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
	NumRobots        map[TeamColor]int
}

func NewCamStats(timeWindow time.Duration, maxBotId int) (s CamStats) {

	s.FrameStats = timing.NewFrameStats(timeWindow)
	s.Robots = map[RobotId]*RobotStats{}
	s.timeWindow = timeWindow
	s.maxBotId = maxBotId
	s.TimingProcessing = new(timing.Timing)
	s.TimingReceiving = new(timing.Timing)

	*s.TimingProcessing = timing.NewTiming(timeWindow)
	*s.TimingReceiving = timing.NewTiming(timeWindow)
	s.NumRobots = map[TeamColor]int{}
	s.NumRobots[TeamYellow] = 0
	s.NumRobots[TeamBlue] = 0

	s.createRobotStats()

	return s
}

func (s *CamStats) createRobotStats() {
	for team := range s.NumRobots {
		for botId := 0; botId < s.maxBotId; botId++ {
			robotId := NewRobotId(botId, team)
			s.Robots[robotId] = new(RobotStats)
			*s.Robots[robotId] = NewRobotStats(s.timeWindow)
		}
	}
}

func (s CamStats) String() string {
	str := fmt.Sprint(s.FrameStats)
	str += fmt.Sprintf(" | %v blue | %v yellow | %v balls\n",
		colorizeByTeam(s.NumRobots[TeamBlue], TeamBlue),
		colorizeByTeam(s.NumRobots[TeamYellow], TeamYellow),
		len(s.Balls))
	str += fmt.Sprintf("Processing Time: %v\n Receiving Time: %v\n", s.TimingProcessing, s.TimingReceiving)

	str += "Balls: "
	for _, ball := range s.Balls {
		str += fmt.Sprintf("%v  ", ball)
	}
	str += "\n"

	sortedIds := sortedRobotIds(s.Robots)
	nHalf := len(sortedIds) / 2
	for i := 0; i < nHalf; i++ {
		robotLeft := sortedIds[i]
		robotRight := sortedIds[i+nHalf]
		str += fmt.Sprintf("%v %v     %v %v\n", robotLeft.String(), s.Robots[robotLeft], robotRight.String(), s.Robots[robotRight])
	}
	return str
}

func (s *CamStats) Clear() {
	s.FrameStats.Clear()
	for _, robot := range s.Robots {
		robot.Clear()
	}
}

func (s *CamStats) Prune(tSent time.Time) {
	s.FrameStats.Prune(tSent.Add(-s.timeWindow))
	for _, robot := range s.Robots {
		robot.Prune(tSent)
	}
	for _, ball := range s.Balls {
		ball.Prune(tSent)
	}
}

func (s *CamStats) NumVisibleRobots(teamColor TeamColor) int {
	numRobots := 0
	for botId := 0; botId < s.maxBotId; botId++ {
		if s.Robots[NewRobotId(botId, teamColor)].Visible {
			numRobots++
		}
	}
	return numRobots
}

func (s *CamStats) Update() {
	for team := range s.NumRobots {
		s.NumRobots[team] = s.NumVisibleRobots(team)
	}
}

func sortedRobotIds(robotStats map[RobotId]*RobotStats) []RobotId {
	ids := make([]RobotId, len(robotStats))
	i := 0
	for id := range robotStats {
		ids[i] = id
		i++
	}
	sort.Slice(ids, func(i, j int) bool {
		if ids[i].Color != ids[j].Color {
			return ids[i].Color < ids[j].Color
		}
		return ids[i].Id < ids[j].Id
	})
	return ids
}
