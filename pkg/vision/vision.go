package vision

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
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

type CamStats struct {
	FramesReceived   uint64
	FramesDropped    uint32
	Fps              *timing.Fps
	TimingProcessing *timing.Timing
	TimingReceiving  *timing.Timing
	lastFrameId      uint32
}

func (s CamStats) String() string {
	fps := s.Fps.Float32()
	return fmt.Sprintf("%v frames received, %v lost @ %4.1f fps\nProcessing Time: %v\n Receiving Time: %v", s.FramesReceived, s.FramesDropped, fps, s.TimingProcessing, s.TimingReceiving)
}

func NewCamStats(timeWindow time.Duration) (s CamStats) {
	s.Fps = new(timing.Fps)
	s.TimingProcessing = new(timing.Timing)
	s.TimingReceiving = new(timing.Timing)

	*s.Fps = timing.NewFps(timeWindow)
	*s.TimingProcessing = timing.NewTiming(timeWindow)
	*s.TimingReceiving = timing.NewTiming(timeWindow)

	return s
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
	} else if *frame.FrameNumber > w.CamStats[camId].lastFrameId {
		w.CamStats[camId].FramesDropped += *frame.FrameNumber - w.CamStats[camId].lastFrameId - 1
	}

	processingTime := time.Duration(int64((*frame.TSent - *frame.TCapture) * 1e9))

	sentSec := int64(*frame.TSent)
	sentNs := int64((*frame.TSent - float64(sentSec)) * 1e9)
	receivingTime := time.Now().Sub(time.Unix(sentSec, sentNs))

	w.CamStats[camId].FramesReceived++
	w.CamStats[camId].Fps.Inc()
	w.CamStats[camId].TimingProcessing.Add(processingTime)
	w.CamStats[camId].TimingReceiving.Add(receivingTime)

	w.CamStats[camId].lastFrameId = *frame.FrameNumber
}
