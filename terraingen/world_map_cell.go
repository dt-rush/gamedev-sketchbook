package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const (
	CELL_WATER  = iota
	CELL_SAND   = iota
	CELL_GRASS  = iota
	CELL_FOREST = iota
)

var TERRAIN_COSTS = []int{
	100,
	1,
	1,
	40,
}

type WorldMapCell struct {
	m     *WorldMap
	rep   string
	kind  int
	pos   Position
	color sdl.Color
	data  interface{}
}

func NewWorldMapCell(m *WorldMap, rep string, kind int, color sdl.Color) WorldMapCell {
	return WorldMapCell{
		m:     m,
		rep:   rep,
		kind:  kind,
		color: color}
}

func (m *WorldMap) WaterCell(depth int) WorldMapCell {
	return NewWorldMapCell(m, "o", CELL_WATER,
		sdl.Color{R: 0, G: 0, B: uint8(48 + 16*depth)})
}

func (m *WorldMap) SandCell() WorldMapCell {
	return NewWorldMapCell(m, ".", CELL_SAND,
		sdl.Color{R: 182, G: 182, B: 0})
}

func (m *WorldMap) GrassCell() WorldMapCell {
	return NewWorldMapCell(m, ".", CELL_GRASS,
		sdl.Color{R: 0, G: 182, B: 0})
}

type ForestCellData struct {
	density float64
}

func (m *WorldMap) ForestCell(density float64) WorldMapCell {
	c := NewWorldMapCell(m, "#", CELL_FOREST,
		sdl.Color{R: 0, G: uint8(180 - (density/0.1)*16), B: 0})
	c.data = ForestCellData{density}
	return c
}
