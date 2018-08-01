package main

import (
	"fmt"
)

type Position struct {
	X int
	Y int
}

func (p Position) String() string {
	return fmt.Sprintf("[%d, %d]", p.X, p.Y)
}

type PositionPair struct {
	p1 Position
	p2 Position
}
