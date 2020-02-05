package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"sync"
	"time"
)

type Stats struct {
	timeWindow  time.Duration
	maxBotId    int
	CamStats    map[int]*CamStats
	tSentLatest time.Time
	tPruned     time.Time
	LogList     []string
	Mutex       sync.Mutex
}

func NewStats(timeWindow time.Duration, maxBotId int) (w Stats) {
	w.timeWindow = timeWindow
	w.maxBotId = maxBotId
	w.CamStats = map[int]*CamStats{}
	return w
}

func (s *Stats) Log(tSent time.Time, str string) {
	timeFormatted := tSent.Format("2006-01-02T15:04:05.000")
	s.LogList = append(s.LogList, timeFormatted+": "+str)
}

func (s *Stats) Process(wrapper *sslproto.SSL_WrapperPacket) {
	s.Mutex.Lock()
	if wrapper == nil {
		for _, camStats := range s.CamStats {
			camStats.Prune(s.tSentLatest)
		}
	} else if wrapper.Detection != nil {
		camId := int(*wrapper.Detection.CameraId)
		if _, ok := s.CamStats[camId]; !ok {
			s.CamStats[camId] = new(CamStats)
			*s.CamStats[camId] = NewCamStats(s.timeWindow, s.maxBotId)
		}
		s.processCam(wrapper.Detection, s.CamStats[camId])
	}
	s.Mutex.Unlock()
}

func (s *Stats) processCam(frame *sslproto.SSL_DetectionFrame, camStats *CamStats) {

	frameId := *frame.FrameNumber
	processingTime := time.Duration(int64((*frame.TSent - *frame.TCapture) * 1e9))

	sentSec := int64(*frame.TSent)
	sentNs := int64((*frame.TSent - float64(sentSec)) * 1e9)
	tSent := time.Unix(sentSec, sentNs)
	receivingTime := time.Now().Sub(tSent)

	camStats.TimingProcessing.Add(processingTime)
	camStats.TimingReceiving.Add(receivingTime)

	camStats.FrameStats.Add(frameId, tSent)
	camStats.FrameStats.Fps.Inc()
	camStats.Prune(tSent)

	for _, robot := range frame.RobotsBlue {
		robotId := NewRobotId(int(*robot.RobotId), TeamBlue)
		camStats.Robots[robotId].Add(tSent, frameId)
		s.updateRobotVisibility(camStats.Robots, robotId, tSent)
	}
	for _, robot := range frame.RobotsYellow {
		robotId := NewRobotId(int(*robot.RobotId), TeamYellow)
		camStats.Robots[robotId].Add(tSent, frameId)
		s.updateRobotVisibility(camStats.Robots, robotId, tSent)
	}

	camStats.Update()

	for _, newBall := range frame.Balls {
		newPos := Position2d{X: *newBall.X / 1000.0, Y: *newBall.Y / 1000.0}
		ball := s.ballStats(tSent, newPos, camStats)
		ball.Add(tSent, frameId, newPos)
	}

	s.pruneBalls(camStats, tSent)

	s.tSentLatest = tSent
}

func (s *Stats) pruneBalls(camStats *CamStats, tSent time.Time) {
	var newBalls []*BallStats
	for _, ball := range camStats.Balls {
		if tSent.Sub(ball.LastDetection.Time).Seconds() < 1 {
			newBalls = append(newBalls, ball)
			s.updateBallVisibility(ball, tSent)
		}
	}
	camStats.Balls = newBalls
}

func (s *Stats) ballStats(tSent time.Time, newBallPos Position2d, camStats *CamStats) *BallStats {
	for _, ball := range camStats.Balls {
		if ball.Matches(tSent, newBallPos) {
			return ball
		}
	}
	stats := NewBallStats(s.timeWindow, Detection{Pos: newBallPos, Time: tSent})
	camStats.Balls = append(camStats.Balls, &stats)
	return &stats
}

func (s *Stats) updateBallVisibility(ball *BallStats, tSent time.Time) {
	numFrames := ball.FrameStats.NumFrames()
	quality := ball.FrameStats.Quality()
	if ball.Visible && numFrames == 0 {
		dt := tSent.Sub(ball.LastDetection.Time).Seconds()
		s.Log(tSent, fmt.Sprintf("Ball vanished at %v after %.1fs", ball.LastDetection.Pos, dt))
		ball.Visible = false
	} else if !ball.Visible && quality > 0.9 {
		s.Log(tSent, fmt.Sprintf("Ball appeared at %v", ball.LastDetection.Pos))
		ball.Visible = true
	}
}

func (s *Stats) updateRobot(robots map[RobotId]*RobotStats, robotId RobotId, frameId uint32, tSent time.Time) {
	robots[robotId].FrameStats.Add(frameId, tSent)
	robots[robotId].FrameStats.Fps.Inc()
}

func (s *Stats) updateRobotVisibility(robots map[RobotId]*RobotStats, robotId RobotId, tSent time.Time) {
	numFrames := robots[robotId].FrameStats.NumFrames()
	quality := robots[robotId].FrameStats.Quality()
	if robots[robotId].Visible && numFrames == 0 {
		s.Log(tSent, fmt.Sprintf("Robot %v vanished", robotId))
		robots[robotId].Visible = false
	} else if !robots[robotId].Visible && quality > 0.9 {
		s.Log(tSent, fmt.Sprintf("Robot %v appeared", robotId))
		robots[robotId].Visible = true
	}
}
