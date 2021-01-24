package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/clock"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/network"
	"github.com/RoboCup-SSL/ssl-quality-inspector/pkg/vision"
	"sort"
	"strings"
	"time"
)

var visionAddress = flag.String("visionAddress", "224.5.23.2:10006", "The multicast address of ssl-vision")

var timeWindowClock = flag.Duration("timeWindowClock", time.Millisecond*500, "The time window for watching clock timing")
var timeWindowVisibility = flag.Duration("timeWindowVisibility", time.Second*5, "The time window for taking timing statistics")
var timeWindowQualityCam = flag.Duration("timeWindowQualityCam", time.Millisecond*500, "The time window for measuring the camera quality")
var timeWindowQualityBall = flag.Duration("timeWindowQualityBall", time.Millisecond*200, "The time window for measuring the ball quality")
var timeWindowQualityRobot = flag.Duration("timeWindowQualityRobot", time.Millisecond*500, "The time window for measuring the robot quality")

func main() {

	flag.Parse()

	multicastSources := network.NewMulticastSourceWatcher()
	go multicastSources.Watch(*visionAddress)

	var statsConfig vision.StatsConfig
	statsConfig.TimeWindowVisibility = *timeWindowVisibility
	statsConfig.TimeWindowQualityCam = *timeWindowQualityCam
	statsConfig.TimeWindowQualityBall = *timeWindowQualityBall
	statsConfig.TimeWindowQualityRobot = *timeWindowQualityRobot
	stats := vision.NewStats(statsConfig)
	udpWatcher := vision.NewUdpWatcher(stats.Process)
	go udpWatcher.Watch(*visionAddress)

	clockWatchers := map[string]*clock.Watcher{}
	activeSources := map[string]bool{}

	for {
		stats.Mutex.Lock()

		multicastSources := multicastSources.GetSources()
		for _, source := range multicastSources {
			if _, ok := activeSources[source]; !ok {
				clockWatcher := clock.NewWatcher(*timeWindowClock)
				clockWatchers[source] = clockWatcher
				go clockWatcher.Watch(source)
				activeSources[source] = true
			}
		}

		watcherDataMap := map[string]clock.Data{}
		for source, clockWatcher := range clockWatchers {
			watcherDataMap[source] = *clockWatcher.GetData()
		}

		// clear screen, move cursor to upper left corner
		fmt.Print("\033[H\033[2J")

		fmt.Println("Vision Multicast sources:")
		fmt.Println(strings.Join(multicastSources, " "))

		fmt.Println()
		fmt.Println("Reference clocks:")
		for source, watcherData := range watcherDataMap {
			fmt.Println(source, " ClockOffset: ", watcherData.ClockOffset)
			fmt.Println(source, "         RTT: ", watcherData.RTT)
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

		stats.Mutex.Unlock()

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
