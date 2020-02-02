package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"sort"
	"time"
)

type CamStats struct {
	FrameStats       *timing.FrameStats
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
	Robots           map[RobotId]*RobotStats
	NumRobots        map[TeamColor]*int
}

func NewCamStats(timeWindow time.Duration) (s CamStats) {
	s.TimingProcessing = new(timing.Timing)
	s.TimingReceiving = new(timing.Timing)

	s.FrameStats = timing.NewFrameStats(timeWindow)
	*s.TimingProcessing = timing.NewTiming(timeWindow)
	*s.TimingReceiving = timing.NewTiming(timeWindow)
	s.Robots = map[RobotId]*RobotStats{}
	s.NumRobots = map[TeamColor]*int{}
	s.NumRobots[TeamYellow] = new(int)
	s.NumRobots[TeamBlue] = new(int)

	return s
}

func (s CamStats) String() string {
	str := fmt.Sprint(s.FrameStats)
	str += fmt.Sprintf(" | %v blue | %v yellow | %v balls\n",
		colorizeByTeam(*s.NumRobots[TeamBlue], TeamBlue),
		colorizeByTeam(*s.NumRobots[TeamYellow], TeamYellow),
		0)
	str += fmt.Sprintf("Processing Time: %v\n Receiving Time: %v", s.TimingProcessing, s.TimingReceiving)
	sortedIds := sortedRobotIds(s.Robots)
	nHalf := len(sortedIds) / 2
	for i := 0; i < nHalf; i++ {
		robotLeft := sortedIds[i]
		robotRight := sortedIds[i+nHalf]
		str += fmt.Sprintf("\n%v %v     %v %v", robotLeft.String(), s.Robots[robotLeft], robotRight.String(), s.Robots[robotRight])
	}
	return str
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
