package main

const GRID_WORLD_DIMENSION = 1024
const GRID_CELL_DIMENSION = GRID_WORLD_DIMENSION / (ENTITYSZ * 2)

const WINDOW_WIDTH = 640
const WINDOW_HEIGHT = 640
const FONTSZ = 16

const GRIDCELL_PX_W = WINDOW_WIDTH / GRID_CELL_DIMENSION
const GRIDCELL_WORLD_W = GRID_WORLD_DIMENSION / GRID_CELL_DIMENSION
const GRIDCELL_PX_H = WINDOW_HEIGHT / GRID_CELL_DIMENSION
const GRIDCELL_WORLD_H = GRID_WORLD_DIMENSION / GRID_CELL_DIMENSION

const FPS = 60
const SCALE = 2
const POINTSZ = 8 * SCALE
const MOVESPEED = 3
const VECLENGTH = 64

const OBSTACLESZ = 32 * SCALE
const ENTITYSZ = 12 * SCALE
