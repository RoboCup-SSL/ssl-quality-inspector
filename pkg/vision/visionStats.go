package vision

import (
	"sync"
	"time"
)

type Stats struct {
	StatsConfig
	CamStats         map[int]*CamStats
	tPruned          time.Time
	LogList          []string
	Mutex            sync.Mutex
	DataLossDetector *DataLossDetector
}

func NewStats(statsConfig StatsConfig) (s *Stats) {
	s = new(Stats)
	s.StatsConfig = statsConfig
	s.CamStats = map[int]*CamStats{}
	s.DataLossDetector = NewDataLossDetector()
	return s
}

func (s *Stats) Log(tSent time.Time, str string) {
	timeFormatted := tSent.Format("2006-01-02T15:04:05.000")
	s.LogList = append(s.LogList, timeFormatted+": "+str)
}

func (s *Stats) Process(wrapper *SSL_WrapperPacket) {
	s.Mutex.Lock()
	if wrapper == nil {
		for _, camStats := range s.CamStats {
			camStats.Clear()
		}
	} else if wrapper.Detection != nil {
		camId := int(*wrapper.Detection.CameraId)
		if _, ok := s.CamStats[camId]; !ok {
			s.CamStats[camId] = new(CamStats)
			*s.CamStats[camId] = NewCamStats(s.StatsConfig)
		}
		s.processCam(wrapper.Detection, s.CamStats[camId])
		s.DataLossDetector.ProcessDetection(wrapper.Detection)
	}
	s.Mutex.Unlock()
}

func (s *Stats) processCam(frame *SSL_DetectionFrame, camStats *CamStats) {

	frameId := *frame.FrameNumber
	processingTime := time.Duration(int64((*frame.TSent - *frame.TCapture) * 1e9))

	sentSec := int64(*frame.TSent)
	sentNs := int64((*frame.TSent - float64(sentSec)) * 1e9)
	tSent := time.Unix(sentSec, sentNs)
	receivingTime := time.Now().Sub(tSent)

	camStats.TimingProcessing.Add(processingTime)
	camStats.TimingReceiving.Add(receivingTime)

	camStats.FrameStats.Add(frameId, tSent)

	processRobots(frame.RobotsBlue, TeamBlue, camStats, tSent, frameId)
	processRobots(frame.RobotsYellow, TeamYellow, camStats, tSent, frameId)

	for _, ball := range frame.Balls {
		ballPos := Position2d{X: *ball.X / 1000.0, Y: *ball.Y / 1000.0}
		ball := camStats.GetBallStats(tSent, ballPos)
		ball.Add(tSent, frameId, ballPos)
	}

	camStats.Prune(tSent)
	camStats.Merge()
}

func processRobots(robots []*SSL_DetectionRobot, teamColor TeamColor, camStats *CamStats, tSent time.Time, frameId uint32) {
	for _, robot := range robots {
		robotId := NewRobotId(int(*robot.RobotId), teamColor)
		robotPos := Position2d{X: *robot.X / 1000.0, Y: *robot.Y / 1000.0}
		robotStats := camStats.GetRobotStats(robotId, tSent, robotPos)
		robotStats.Add(tSent, frameId, robotPos)
	}
}
