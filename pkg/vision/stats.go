package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"sort"
	"time"
)

type TeamColor string

const (
	TeamYellow TeamColor = "Y"
	TeamBlue             = "B"
)

type RobotId struct {
	Id    int
	Color TeamColor
}

func NewRobotId(id int, color TeamColor) RobotId {
	return RobotId{id, color}
}

func (s RobotId) String() string {
	return fmt.Sprintf("%v %v", s.Id, s.Color)
}

type FrameStats struct {
	FramesReceived uint64
	FramesDropped  uint32
	Fps            *timing.Fps
	lastFrameId    uint32
}

func NewFrameStats(timeWindow time.Duration) (s *FrameStats) {
	s = new(FrameStats)
	s.Fps = new(timing.Fps)
	*s.Fps = timing.NewFps(timeWindow)

	return s
}

func (s FrameStats) String() string {
	fps := s.Fps.Float32()
	return fmt.Sprintf("%v frames received, %v lost @ %4.1f fps", s.FramesReceived, s.FramesDropped, fps)
}

type RobotStats struct {
	FrameStats *FrameStats
}

func NewRobotStats(timeWindow time.Duration) (s RobotStats) {
	s.FrameStats = NewFrameStats(timeWindow)

	return s
}

func (s RobotStats) String() string {
	return fmt.Sprintf("%v", s.FrameStats)
}

type CamStats struct {
	FrameStats       *FrameStats
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
	Robots           map[RobotId]*RobotStats
}

func NewCamStats(timeWindow time.Duration) (s CamStats) {
	s.TimingProcessing = new(timing.Timing)
	s.TimingReceiving = new(timing.Timing)

	s.FrameStats = NewFrameStats(timeWindow)
	*s.TimingProcessing = timing.NewTiming(timeWindow)
	*s.TimingReceiving = timing.NewTiming(timeWindow)
	s.Robots = map[RobotId]*RobotStats{}

	return s
}

func (s CamStats) String() string {
	str := fmt.Sprintf("%v\nProcessing Time: %v\n Receiving Time: %v", s.FrameStats, s.TimingProcessing, s.TimingReceiving)
	sortedIds := sortedRobotIds(s.Robots)
	for _, robotId := range sortedIds {
		str += fmt.Sprintf("\n%v %v", robotId.String(), s.Robots[robotId])
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
