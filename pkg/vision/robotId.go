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
