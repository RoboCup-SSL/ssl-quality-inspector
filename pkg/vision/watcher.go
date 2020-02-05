package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"sync"
	"time"
)

const maxDatagramSize = 8192

type Watcher struct {
	Address     string
	timeWindow  time.Duration
	maxBotId    int
	CamStats    map[int]*CamStats
	tSentLatest time.Time
	tPruned     time.Time
	LogList     []string
	Mutex       sync.Mutex
}

func NewWatcher(address string, timeWindow time.Duration, maxBotId int) (w Watcher) {
	w.Address = address
	w.timeWindow = timeWindow
	w.maxBotId = maxBotId
	w.CamStats = map[int]*CamStats{}
	return w
}

func (w *Watcher) Log(tSent time.Time, str string) {
	timeFormatted := tSent.Format("2006-01-02T15:04:05.000")
	w.LogList = append(w.LogList, timeFormatted+": "+str)
}

func (w *Watcher) Watch() {
	addr, err := net.ResolveUDPAddr("udp", w.Address)
	if err != nil {
		log.Print(err)
		return
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Print(err)
		return
	}

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("Could not set read buffer to %v.", maxDatagramSize)
	}

	go func() {
		for {
			w.Mutex.Lock()
			w.prune()
			w.Mutex.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

	w.receive(conn)
}

func (w *Watcher) receive(conn *net.UDPConn) {
	b := make([]byte, maxDatagramSize)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Print("Could not read", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if n >= maxDatagramSize {
			log.Fatal("Buffer size too small")
		}
		wrapper := sslproto.SSL_WrapperPacket{}
		if err := proto.Unmarshal(b[0:n], &wrapper); err != nil {
			log.Println("Could not unmarshal message")
			continue
		}

		if wrapper.Detection != nil {
			w.Mutex.Lock()
			w.processDetectionMessage(wrapper.Detection)
			w.Mutex.Unlock()
		}
	}
}

func (w *Watcher) prune() {
	if w.tPruned == w.tSentLatest {
		for _, camStats := range w.CamStats {
			camStats.Clear()
		}
	} else {
		for _, camStats := range w.CamStats {
			camStats.Prune(w.tSentLatest)
		}
		w.tPruned = w.tSentLatest
	}
}

func (w *Watcher) processDetectionMessage(frame *sslproto.SSL_DetectionFrame) {
	camId := int(*frame.CameraId)
	if _, ok := w.CamStats[camId]; !ok {
		w.CamStats[camId] = new(CamStats)
		*w.CamStats[camId] = NewCamStats(w.timeWindow, w.maxBotId)
	}
	w.processCam(frame, w.CamStats[camId])
}

func (w *Watcher) processCam(frame *sslproto.SSL_DetectionFrame, camStats *CamStats) {

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
		w.updateRobotVisibility(camStats.Robots, robotId, tSent)
	}
	for _, robot := range frame.RobotsYellow {
		robotId := NewRobotId(int(*robot.RobotId), TeamYellow)
		camStats.Robots[robotId].Add(tSent, frameId)
		w.updateRobotVisibility(camStats.Robots, robotId, tSent)
	}

	camStats.Update()

	for _, newBall := range frame.Balls {
		newPos := Position2d{X: *newBall.X / 1000.0, Y: *newBall.Y / 1000.0}
		ball := w.ballStats(tSent, newPos, camStats)
		ball.Add(tSent, frameId, newPos)
	}

	w.pruneBalls(camStats, tSent)

	w.tSentLatest = tSent
}

func (w *Watcher) pruneBalls(camStats *CamStats, tSent time.Time) {
	var newBalls []*BallStats
	for _, ball := range camStats.Balls {
		if tSent.Sub(ball.LastDetection.Time).Seconds() < 1 {
			newBalls = append(newBalls, ball)
			w.updateBallVisibility(ball, tSent)
		}
	}
	camStats.Balls = newBalls
}

func (w *Watcher) ballStats(tSent time.Time, newBallPos Position2d, camStats *CamStats) *BallStats {
	for _, ball := range camStats.Balls {
		if ball.Matches(tSent, newBallPos) {
			return ball
		}
	}
	stats := NewBallStats(w.timeWindow, Detection{Pos: newBallPos, Time: tSent})
	camStats.Balls = append(camStats.Balls, &stats)
	return &stats
}

func (w *Watcher) updateBallVisibility(ball *BallStats, tSent time.Time) {
	numFrames := ball.FrameStats.NumFrames()
	quality := ball.FrameStats.Quality()
	if ball.Visible && numFrames == 0 {
		dt := tSent.Sub(ball.LastDetection.Time).Seconds()
		w.Log(tSent, fmt.Sprintf("Ball vanished at %v after %.1fs", ball.LastDetection.Pos, dt))
		ball.Visible = false
	} else if !ball.Visible && quality > 0.9 {
		w.Log(tSent, fmt.Sprintf("Ball appeared at %v", ball.LastDetection.Pos))
		ball.Visible = true
	}
}

func (w *Watcher) updateRobot(robots map[RobotId]*RobotStats, robotId RobotId, frameId uint32, tSent time.Time) {
	robots[robotId].FrameStats.Add(frameId, tSent)
	robots[robotId].FrameStats.Fps.Inc()
}

func (w *Watcher) updateRobotVisibility(robots map[RobotId]*RobotStats, robotId RobotId, tSent time.Time) {
	numFrames := robots[robotId].FrameStats.NumFrames()
	quality := robots[robotId].FrameStats.Quality()
	if robots[robotId].Visible && numFrames == 0 {
		w.Log(tSent, fmt.Sprintf("Robot %v vanished", robotId))
		robots[robotId].Visible = false
	} else if !robots[robotId].Visible && quality > 0.9 {
		w.Log(tSent, fmt.Sprintf("Robot %v appeared", robotId))
		robots[robotId].Visible = true
	}
}
