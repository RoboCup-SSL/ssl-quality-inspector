package vision

import (
	"fmt"
	"math"
)

type Position2d struct {
	X float32
	Y float32
}

func (p *Position2d) DistanceTo(pos Position2d) float64 {
	dx := p.X - pos.X
	dy := p.Y - pos.Y
	return math.Sqrt(float64(dx*dx + dy*dy))
}

func (p Position2d) String() string {
	return fmt.Sprintf("x:%6.3f|y:%6.3f", p.X, p.Y)
}
