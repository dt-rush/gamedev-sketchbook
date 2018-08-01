package main

func screenSpaceToWorldSpace(p Point2D) Point2D {
	return Point2D{
		int(WORLD_WIDTH * (float64(p.X) / float64(WINDOW_WIDTH))),
		int(WORLD_HEIGHT * (1 - float64(p.Y)/float64(WINDOW_HEIGHT))),
	}
}

func worldSpaceToScreenSpace(p Point2D) Point2D {
	return Point2D{
		int(WINDOW_WIDTH * (float64(p.X) / float64(WORLD_WIDTH))),
		int(WINDOW_HEIGHT * (1 - float64(p.Y)/float64(WORLD_HEIGHT))),
	}
}
