package vision

import "fmt"

type TeamColor string

const (
	TeamYellow TeamColor = "Y"
	TeamBlue             = "B"
)

type RobotId struct {
	Id    int
	Color TeamColor
}

func NewRobotId(id int, color TeamColor) RobotId {
	return RobotId{id, color}
}

func (s RobotId) String() string {
	return fmt.Sprintf("%v %v",
		colorizeByTeam(fmt.Sprintf("%2d", s.Id), s.Color),
		colorizeByTeam(s.Color, s.Color))
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
