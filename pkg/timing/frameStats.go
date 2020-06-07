package timing

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type FrameStats struct {
	Fps    *Fps
	frames map[uint32]time.Time
	mutex  sync.Mutex
}

func NewFrameStats(timeWindow time.Duration) (s *FrameStats) {
	s = new(FrameStats)
	s.Fps = NewFps(timeWindow)
	s.frames = map[uint32]time.Time{}
	return s
}

func (s *FrameStats) Add(frameId uint32, t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.frames[frameId] = t
	s.Fps.Inc()
}

func (s *FrameStats) Prune(to time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for frameId, t := range s.frames {
		if t.Before(to) {
			delete(s.frames, frameId)
		}
	}
}

func (s *FrameStats) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.frames = map[uint32]time.Time{}
	s.Fps.Clear()
}

func (s *FrameStats) Quality() float64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	min := uint32(math.MaxUint32)
	max := uint32(0)
	for frameId := range s.frames {
		if frameId < min {
			min = frameId
		}
		if frameId > max {
			max = frameId
		}
	}
	sum := float64(len(s.frames))
	return sum / float64(max-min+1)
}

func (s *FrameStats) NumFrames() int {
	return len(s.frames)
}

func (s *FrameStats) String() string {
	fps := s.Fps.Float32()
	quality := s.Quality()
	return fmt.Sprintf("%v @ %3.0f fps", colorizePercent(quality), fps)
}
