package vision

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"sync"
	"time"
)

type Stats struct {
	timeWindow time.Duration
	maxBotId   int
	CamStats   map[int]*CamStats
	tPruned    time.Time
	LogList    []string
	Mutex      sync.Mutex
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
			camStats.Clear()
		}
	} else if wrapper.Detection != nil {
		camId := int(*wrapper.Detection.CameraId)
		if _, ok := s.CamStats[camId]; !ok {
			s.CamStats[camId] = new(CamStats)
			*s.CamStats[camId] = NewCamStats(s.timeWindow)
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

	for _, robot := range frame.RobotsBlue {
		robotId := NewRobotId(int(*robot.RobotId), TeamBlue)
		robotPos := Position2d{X: *robot.X / 1000.0, Y: *robot.Y / 1000.0}
		robotStats := camStats.GetRobotStats(robotId, tSent, robotPos)
		robotStats.Add(tSent, frameId, robotPos)
	}
	for _, robot := range frame.RobotsYellow {
		robotId := NewRobotId(int(*robot.RobotId), TeamYellow)
		robotPos := Position2d{X: *robot.X / 1000.0, Y: *robot.Y / 1000.0}
		robotStats := camStats.GetRobotStats(robotId, tSent, robotPos)
		robotStats.Add(tSent, frameId, robotPos)
	}
	for _, ball := range frame.Balls {
		ballPos := Position2d{X: *ball.X / 1000.0, Y: *ball.Y / 1000.0}
		ball := camStats.GetBallStats(tSent, ballPos)
		ball.Add(tSent, frameId, ballPos)
	}

	camStats.Prune(tSent)
}
