package clock

import (
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/timing"
	"github.com/beevik/ntp"
	"time"
)

type Watcher struct {
	Host        string
	online      bool
	ClockOffset *timing.Timing
	RTT         *timing.Timing
}

func NewWatcher(host string, timeWindow time.Duration) (w Watcher) {
	w.Host = host
	w.online = false
	w.ClockOffset = new(timing.Timing)
	w.RTT = new(timing.Timing)

	*w.ClockOffset = timing.NewTiming(timeWindow)
	*w.RTT = timing.NewTiming(timeWindow)

	return w
}

func (w *Watcher) Watch() {
	for {
		response, err := ntp.Query(w.Host)
		if err != nil {
			w.online = false
			w.ClockOffset.Clear()
			w.RTT.Clear()
			time.Sleep(time.Second)
		} else {
			w.online = true
			w.ClockOffset.Add(response.ClockOffset)
			w.RTT.Add(response.RTT)
		}
		time.Sleep(time.Millisecond * 100)
	}
}
