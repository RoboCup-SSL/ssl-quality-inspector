package network

import (
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

type MulticastSourceWatcher struct {
	Address string
	Sources []string
}

func NewMulticastSourceWatcher(address string) (w MulticastSourceWatcher) {
	w.Address = address
	return w
}

func (w *MulticastSourceWatcher) Watch() {
	w.watchAddress(w.Address)
}

func (w *MulticastSourceWatcher) watchAddress(address string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("Could not set read buffer to %v.", maxDatagramSize)
	}
	log.Println("Receiving from", address)
	for {
		_, udpAddr, err := conn.ReadFromUDP([]byte{0})
		if err != nil {
			log.Print("Could not read", err)
			time.Sleep(1 * time.Second)
			continue
		}
		w.addSource(address, udpAddr.IP.String())
	}
}

func (w *MulticastSourceWatcher) addSource(address string, remote string) {
	for _, a := range w.Sources {
		if a == remote {
			return
		}
	}
	w.Sources = append(w.Sources, remote)
	log.Printf("remote ip on %v: %v\n", address, remote)
}
