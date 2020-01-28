package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"math"
	"sort"
	"sync"
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
	return fmt.Sprintf("%2d %v", s.Id, s.Color)
}

type FrameStats struct {
	Fps    *timing.Fps
	frames map[uint32]time.Time
	mutex  *sync.Mutex
}

func NewFrameStats(timeWindow time.Duration) (s *FrameStats) {
	s = new(FrameStats)
	s.Fps = new(timing.Fps)
	*s.Fps = timing.NewFps(timeWindow)
	s.frames = map[uint32]time.Time{}
	s.mutex = new(sync.Mutex)

	return s
}

func (s *FrameStats) Add(frameId uint32, t time.Time) {
	s.mutex.Lock()
	s.frames[frameId] = t
	s.mutex.Unlock()
}

func (s *FrameStats) Prune(to time.Time) {
	s.mutex.Lock()
	for frameId, t := range s.frames {
		if t.Before(to) {
			delete(s.frames, frameId)
		}
	}
	s.mutex.Unlock()
}

func (s *FrameStats) Quality() float64 {
	min := uint32(math.MaxUint32)
	max := uint32(0)
	s.mutex.Lock()
	for frameId := range s.frames {
		if frameId < min {
			min = frameId
		}
		if frameId > max {
			max = frameId
		}
	}
	sum := float64(len(s.frames))
	s.mutex.Unlock()
	return sum / float64(max-min+1)
}

func (s FrameStats) String() string {
	fps := s.Fps.Float32()
	return fmt.Sprintf("%4.0f%% @ %3.0f fps", s.Quality()*100, fps)
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
	str := fmt.Sprintln(s.FrameStats)
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
