package vision

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

type Watcher struct {
	Address    string
	timeWindow time.Duration
	maxBotId   int
	CamStats   map[int]*CamStats
}

func NewWatcher(address string, timeWindow time.Duration, maxBotId int) (w Watcher) {
	w.Address = address
	w.timeWindow = timeWindow
	w.maxBotId = maxBotId
	w.CamStats = map[int]*CamStats{}
	return w
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
			w.processDetectionMessage(wrapper.Detection)
		}
	}
}

func (w *Watcher) processDetectionMessage(frame *sslproto.SSL_DetectionFrame) {
	camId := int(*frame.CameraId)
	if _, ok := w.CamStats[camId]; !ok {
		w.CamStats[camId] = new(CamStats)
		*w.CamStats[camId] = NewCamStats(w.timeWindow)
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
	camStats.FrameStats.Prune(tSent.Add(-w.timeWindow))
	camStats.FrameStats.Fps.Inc()

	for botId := 0; botId < w.maxBotId; botId++ {
		w.pruneRobot(camStats.Robots, NewRobotId(botId, TeamBlue), tSent)
		w.pruneRobot(camStats.Robots, NewRobotId(botId, TeamYellow), tSent)
	}

	for _, robot := range frame.RobotsBlue {
		robotId := NewRobotId(int(*robot.RobotId), TeamBlue)
		w.updateRobot(camStats.Robots, robotId, frameId, tSent)
	}
	for _, robot := range frame.RobotsYellow {
		robotId := NewRobotId(int(*robot.RobotId), TeamYellow)
		w.updateRobot(camStats.Robots, robotId, frameId, tSent)
	}
}

func (w *Watcher) updateRobot(robots map[RobotId]*RobotStats, robotId RobotId, frameId uint32, tSent time.Time) {
	robots[robotId].FrameStats.Add(frameId, tSent)
	robots[robotId].FrameStats.Fps.Inc()
}

func (w *Watcher) pruneRobot(robots map[RobotId]*RobotStats, robotId RobotId, tSent time.Time) {
	if _, ok := robots[robotId]; !ok {
		robots[robotId] = new(RobotStats)
		*robots[robotId] = NewRobotStats(w.timeWindow)
	}
	robots[robotId].FrameStats.Prune(tSent.Add(-w.timeWindow))
	robots[robotId].FrameStats.Fps.Prune()
}
