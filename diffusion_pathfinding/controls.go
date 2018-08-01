package main

type ControlMode int

const N_MODES = 3

const (
	MODE_PLACING_WAYPOINT = 0
	MODE_PLACING_OBSTACLE = iota
	MODE_PLACING_CHASER   = iota
)

var MODENAMES []string = []string{
	"MODE_PLACING_WAYPOINT",
	"MODE_PLACING_OBSTACLE",
	"MODE_PLACING_CHASER",
}

type Controls struct {
	mode ControlMode
}

func NewControls() *Controls {
	return &Controls{mode: MODE_PLACING_WAYPOINT}
}

func (c *Controls) ToggleMode() {
	c.mode = (c.mode + 1) % N_MODES
}
