package network

import (
	"log"
	"net"
	"sync"
	"time"
)

const maxDatagramSize = 8192

type MulticastSourceWatcher struct {
	sources []string
	mutex   sync.Mutex
}

func NewMulticastSourceWatcher() (w *MulticastSourceWatcher) {
	w = new(MulticastSourceWatcher)
	return w
}

func (w *MulticastSourceWatcher) GetSources() []string {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	cpy := make([]string, len(w.sources))
	copy(cpy, w.sources)
	return cpy
}

func (w *MulticastSourceWatcher) Watch(address string) {
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
		w.mutex.Lock()
		w.addSource(address, udpAddr.IP.String())
		w.mutex.Unlock()
	}
}

func (w *MulticastSourceWatcher) addSource(address string, remote string) {
	for _, a := range w.sources {
		if a == remote {
			return
		}
	}
	w.sources = append(w.sources, remote)
	log.Printf("remote ip on %v: %v\n", address, remote)
}
