package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"math"
	"sync"
	"time"
)

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
	s.Fps.Prune()
	s.mutex.Unlock()
}

func (s *FrameStats) Clear() {
	s.mutex.Lock()
	s.frames = map[uint32]time.Time{}
	s.Fps.Clear()
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

func (s *FrameStats) NumFrames() int {
	return len(s.frames)
}

func (s FrameStats) String() string {
	fps := s.Fps.Float32()
	quality := s.Quality()
	return fmt.Sprintf("%v @ %3.0f fps", colorizePercent(quality), fps)
}
