package visualization

import (
	"google.golang.org/protobuf/proto"
	"log"
	"net"
)

const maxDatagramSize = 8192

type Publisher struct {
	address string
	conn    *net.UDPConn
}

func NewPublisher(address string) (publisher Publisher) {

	publisher.address = address

	publisher.connect()

	return
}

func (p *Publisher) connect() {
	p.conn = nil

	addr, err := net.ResolveUDPAddr("udp", p.address)
	if err != nil {
		log.Printf("Could not resolve address '%v': %v", p.address, err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Printf("Could not connect to '%v': %v", addr, err)
		return
	}

	if err := conn.SetWriteBuffer(maxDatagramSize); err != nil {
		log.Printf("Could not set write buffer to %v.", maxDatagramSize)
	}
	log.Println("Publishing to", p.address)

	p.conn = conn
	return
}

func (p *Publisher) Send(frame *VisualizationFrame) {
	bytes, err := proto.Marshal(frame)
	if err != nil {
		log.Printf("Could not marshal referee message: %v\nError: %v", frame, err)
		return
	}
	_, err = p.conn.Write(bytes)
	if err != nil {
		log.Printf("Could not write message: %v", err)
	}
}
