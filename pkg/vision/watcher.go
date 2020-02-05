package vision

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

type UdpWatcher struct {
	Address  string
	Callback func(*sslproto.SSL_WrapperPacket)
}

func NewUdpWatcher(address string, callback func(*sslproto.SSL_WrapperPacket)) (w UdpWatcher) {
	w.Address = address
	w.Callback = callback
	return w
}

func (w *UdpWatcher) Watch() {
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

	for {
		c1 := make(chan *sslproto.SSL_WrapperPacket, 1)

		go w.receive(conn, c1)

		select {
		case wrapper := <-c1:
			w.Callback(wrapper)
		case <-time.After(1 * time.Second):
			w.Callback(nil)
		}
	}
}

func (w *UdpWatcher) receive(conn *net.UDPConn, c1 chan *sslproto.SSL_WrapperPacket) {
	b := make([]byte, maxDatagramSize)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Print("Could not read", err)
			continue
		} else if n >= maxDatagramSize {
			log.Print("Buffer size too small")
		}

		wrapper := new(sslproto.SSL_WrapperPacket)
		if err := proto.Unmarshal(b[0:n], wrapper); err != nil {
			log.Println("Could not unmarshal message")
			continue
		}
		c1 <- wrapper
	}
}
