package main

const WORLD_HEIGHT = 1024
const WORLD_WIDTH = 1024

const WINDOW_HEIGHT = 640
const WINDOW_WIDTH = 640

const FPS = 8

const N_LAKES = 1
const PSCALE = 8.0
const PW = int(float64(WORLD_WIDTH) / PSCALE)
const PH = int(float64(WORLD_HEIGHT) / PSCALE)
const WATER_CUTOFF = 0.6

const DRAW_LAKE_SOURCE = true
const DRAW_LAKE_VERTICES = true
