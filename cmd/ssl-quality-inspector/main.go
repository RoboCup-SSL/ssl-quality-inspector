package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/clock"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/network"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/vision"
	"sort"
	"strings"
	"sync"
	"time"
)

var visionAddress = flag.String("visionAddress", "224.5.23.2:10006", "The multicast address of ssl-vision")

var timeWindowClock = flag.Duration("timeWindowClock", time.Millisecond*500, "The time window for watching clock timing")
var timeWindowVisibility = flag.Duration("timeWindowVisibility", time.Second*5, "The time window for taking timing statistics")
var timeWindowQualityCam = flag.Duration("timeWindowQualityCam", time.Millisecond*500, "The time window for measuring the camera quality")
var timeWindowQualityBall = flag.Duration("timeWindowQualityBall", time.Millisecond*200, "The time window for measuring the ball quality")
var timeWindowQualityRobot = flag.Duration("timeWindowQualityRobot", time.Millisecond*500, "The time window for measuring the robot quality")

func main() {

	multicastSources := network.NewMulticastSourceWatcher(*visionAddress)
	go multicastSources.Watch()

	var statsConfig vision.StatsConfig
	statsConfig.TimeWindowVisibility = *timeWindowVisibility
	statsConfig.TimeWindowQualityCam = *timeWindowQualityCam
	statsConfig.TimeWindowQualityBall = *timeWindowQualityBall
	statsConfig.TimeWindowQualityRobot = *timeWindowQualityRobot
	stats := vision.NewStats(statsConfig)
	udpWatcher := vision.NewUdpWatcher(*visionAddress, stats.Process)
	go udpWatcher.Watch()

	var clockWatchers []*clock.Watcher
	activeSources := map[string]struct{}{}
	clockWatchersMutex := new(sync.Mutex)

	for {
		multicastSources.Mutex.Lock()
		stats.Mutex.Lock()
		clockWatchersMutex.Lock()

		for _, source := range multicastSources.Sources {
			if _, ok := activeSources[source]; !ok {
				clockWatcher := clock.NewWatcher(source, *timeWindowClock)
				clockWatcher.Mutex = clockWatchersMutex
				clockWatchers = append(clockWatchers, &clockWatcher)
				go clockWatcher.Watch()
				activeSources[source] = struct{}{}
			}
		}

		// clear screen, move cursor to upper left corner
		fmt.Print("\033[H\033[2J")

		fmt.Println("Vision Multicast sources:")
		fmt.Println(strings.Join(multicastSources.Sources, " "))

		fmt.Println()
		fmt.Println("Reference clocks:")
		for _, clockWatcher := range clockWatchers {
			fmt.Println(clockWatcher.Host, " ClockOffset: ", clockWatcher.ClockOffset)
			fmt.Println(clockWatcher.Host, "         RTT: ", clockWatcher.RTT)
		}

		fmt.Println()
		fmt.Println("Vision:")
		for camId := range sortedCamIds(stats.CamStats) {
			fmt.Print("Camera ", camId)
			fmt.Println(stats.CamStats[camId])
			fmt.Println()
		}

		numLogs := len(stats.LogList)
		nEntries := 20
		oldest := numLogs - 1 - nEntries
		if oldest < 0 {
			oldest = 0
		}
		for i := oldest; i < numLogs; i++ {
			fmt.Println(stats.LogList[i])
		}

		fmt.Println()

		clockWatchersMutex.Unlock()
		stats.Mutex.Unlock()
		multicastSources.Mutex.Unlock()

		time.Sleep(time.Second)
	}
}

func sortedCamIds(camStats map[int]*vision.CamStats) []int {
	keys := make([]int, 0, len(camStats))
	for k := range camStats {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
