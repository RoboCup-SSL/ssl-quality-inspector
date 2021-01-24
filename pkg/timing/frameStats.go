package timing

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type FrameStats struct {
	Fps        *Fps
	frames     map[uint32]time.Time
	mutex      sync.Mutex
	lastTime   *time.Time
	deltaTimes map[uint32]time.Duration
}

func NewFrameStats(timeWindow time.Duration) (s *FrameStats) {
	s = new(FrameStats)
	s.Fps = NewFps(timeWindow)
	s.frames = map[uint32]time.Time{}
	s.deltaTimes = map[uint32]time.Duration{}
	return s
}

func (s *FrameStats) Add(frameId uint32, t time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.frames[frameId] = t
	s.Fps.Inc()
	if s.lastTime != nil {
		s.deltaTimes[frameId] = t.Sub(*s.lastTime)
	}
	s.lastTime = &t
}

func (s *FrameStats) Prune(to time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for frameId, t := range s.frames {
		if t.Before(to) {
			delete(s.frames, frameId)
			delete(s.deltaTimes, frameId)
		}
	}
}

func (s *FrameStats) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.frames = map[uint32]time.Time{}
	s.deltaTimes = map[uint32]time.Duration{}
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

func (s *FrameStats) DeltaTime() (mu float64, stdDev float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var sum time.Duration
	for _, dt := range s.deltaTimes {
		sum += dt
	}
	mu = sum.Seconds() / float64(len(s.deltaTimes))

	var sqSum float64
	for _, dt := range s.deltaTimes {
		diff := mu - dt.Seconds()
		sqSum += diff * diff
	}
	stdDev = math.Sqrt(sqSum / float64(len(s.deltaTimes)))
	return
}

func (s *FrameStats) NumFrames() int {
	return len(s.frames)
}

func (s *FrameStats) String() string {
	fps := s.Fps.Float32()
	quality := s.Quality()
	dt, dtStdDev := s.DeltaTime()
	return fmt.Sprintf("%v @ %3.0f fps | Δ %.1fms σ %.3f", colorizePercent(quality), fps, dt*1000, dtStdDev*1000)
}
