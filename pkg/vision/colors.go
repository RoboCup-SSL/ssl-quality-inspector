package vision

import (
	"fmt"
	"math"
)

func colorizePercent(value float64) string {
	var color int
	if value < 0.3 {
		// Red
		color = 31
	} else if value < 0.6 {
		// Yellow
		color = 33
	} else {
		// Green
		color = 32
	}
	return fmt.Sprintf("\u001b[%dm%4.0f%%\u001b[0m", color, math.Round(value*100))
}

func colorizeByTeam(str interface{}, team TeamColor) string {
	var color int
	switch team {
	case TeamBlue:
		color = 34
	case TeamYellow:
		color = 93
	default:
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("\u001b[%dm%v\u001b[0m", color, str)
}
