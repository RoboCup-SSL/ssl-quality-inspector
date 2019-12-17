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
	CamStats   map[int]*CamStats
}

func NewWatcher(address string, timeWindow time.Duration) (w Watcher) {
	w.Address = address
	w.timeWindow = timeWindow
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
	} else if *frame.FrameNumber > w.CamStats[camId].FrameStats.lastFrameId {
		w.CamStats[camId].FrameStats.FramesDropped += *frame.FrameNumber - w.CamStats[camId].FrameStats.lastFrameId - 1
	}
	w.processCam(frame, w.CamStats[camId])
}

func (w *Watcher) processCam(frame *sslproto.SSL_DetectionFrame, camStats *CamStats) {

	processingTime := time.Duration(int64((*frame.TSent - *frame.TCapture) * 1e9))

	sentSec := int64(*frame.TSent)
	sentNs := int64((*frame.TSent - float64(sentSec)) * 1e9)
	receivingTime := time.Now().Sub(time.Unix(sentSec, sentNs))

	camStats.TimingProcessing.Add(processingTime)
	camStats.TimingReceiving.Add(receivingTime)

	camStats.FrameStats.FramesReceived++
	camStats.FrameStats.Fps.Inc()
	camStats.FrameStats.lastFrameId = *frame.FrameNumber

	for _, robot := range frame.RobotsBlue {
		robotId := NewRobotId(int(*robot.RobotId), TeamBlue)
		w.processRobot(frame, camStats, robotId)
	}
	for _, robot := range frame.RobotsYellow {
		robotId := NewRobotId(int(*robot.RobotId), TeamYellow)
		w.processRobot(frame, camStats, robotId)
	}
}

func (w *Watcher) processRobot(frame *sslproto.SSL_DetectionFrame, camStats *CamStats, robotId RobotId) {
	if _, ok := camStats.Robots[robotId]; !ok {
		camStats.Robots[robotId] = new(RobotStats)
		*camStats.Robots[robotId] = NewRobotStats(w.timeWindow)
	} else if *frame.FrameNumber > camStats.Robots[robotId].FrameStats.lastFrameId {
		camStats.Robots[robotId].FrameStats.FramesDropped += *frame.FrameNumber - camStats.Robots[robotId].FrameStats.lastFrameId - 1
	}
	camStats.Robots[robotId].FrameStats.FramesReceived++
	camStats.Robots[robotId].FrameStats.Fps.Inc()
	camStats.Robots[robotId].FrameStats.lastFrameId = *frame.FrameNumber
}
