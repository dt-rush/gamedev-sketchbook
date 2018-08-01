package main

type Lake struct {
	id           int
	source       Point2D
	Vertices     []Point2D
	interpolated []bool
	vx           []int16
	vy           []int16
	highlighted  bool
}

func (l *Lake) containsPoint2D(p Point2D) bool {
	return Point2DInPolygon(p, l.Vertices)
}

func (l *Lake) buildVXVY() {
	l.vx = make([]int16, len(l.Vertices))
	l.vy = make([]int16, len(l.Vertices))
	for i, v := range l.Vertices {
		ssv := worldSpaceToScreenSpace(v)
		if ssv.X < 0 {
			ssv.X = 0
		} else if ssv.X > WINDOW_WIDTH-2 {
			ssv.X = WINDOW_WIDTH - 2
		}
		if ssv.Y < 0 {
			ssv.Y = 0
		} else if ssv.Y > WINDOW_HEIGHT-2 {
			ssv.Y = WINDOW_HEIGHT - 2
		}
		l.vx[i] = int16(ssv.X)
		l.vy[i] = int16(ssv.Y)
	}

}
