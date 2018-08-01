package main

type Entity struct {
	pos        Position
	moveTarget *Position
	path       []Position
}
