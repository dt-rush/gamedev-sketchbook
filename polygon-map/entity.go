package main

type Entity struct {
	pos        Point2D
	moveTarget *Point2D
	path       []Point2D
}
