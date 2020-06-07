package clock

import (
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"github.com/beevik/ntp"
	"sync"
	"time"
)

type Watcher struct {
	online bool
	data   *Data
	mutex  sync.Mutex
}

type Data struct {
	ClockOffset *timing.Timing
	RTT         *timing.Timing
}

func NewWatcher(timeWindow time.Duration) (w *Watcher) {
	w = new(Watcher)
	w.online = false
	w.data = new(Data)
	w.data.ClockOffset = timing.NewTiming(timeWindow)
	w.data.RTT = timing.NewTiming(timeWindow)

	return w
}

func (w *Watcher) GetData() *Data {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.data
}

func (w *Watcher) Watch(host string) {
	for {
		response, err := ntp.Query(host)
		w.mutex.Lock()
		if err != nil {
			w.online = false
			w.data.ClockOffset.Clear()
			w.data.RTT.Clear()
			time.Sleep(time.Second)
		} else {
			w.online = true
			w.data.ClockOffset.Add(response.ClockOffset)
			w.data.RTT.Add(response.RTT)
		}
		w.mutex.Unlock()
		time.Sleep(time.Millisecond * 100)
	}
}
