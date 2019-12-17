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

//var refereeAddress = flag.String("refereeAddress", "224.5.23.1:10003", "The multicast address of ssl-game-controller")
var visionAddress = flag.String("visionAddress", "224.5.23.2:10006", "The multicast address of ssl-vision")

var timeWindow = flag.Duration("timeWindow", time.Second*5, "The time window for taking timing statistics")

func main() {

	multicastSources := network.NewMulticastSourceWatcher(*visionAddress)
	go multicastSources.Watch()

	visionWatcher := vision.NewWatcher(*visionAddress, *timeWindow)
	go visionWatcher.Watch()

	var clockWatchers []clock.Watcher
	activeSources := map[string]struct{}{}

	for {
		for _, source := range multicastSources.Sources {
			if _, ok := activeSources[source]; !ok {
				clockWatcher := clock.NewWatcher(source, *timeWindow)
				clockWatchers = append(clockWatchers, clockWatcher)
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
		for camId := range sortedCamIds(visionWatcher.CamStats) {
			fmt.Println("Camera ", camId)
			fmt.Println(visionWatcher.CamStats[camId])
			fmt.Println()
		}

		fmt.Println()
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
