package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	gl "github.com/chsc/gogl/gl43"
)

type StaticModelDef struct {
	verts       []gl.Float
	vertIndices []gl.Uint
	norms       []gl.Float
	normIndices []gl.Uint
}

// from http://www.opengl-tutorial.org/beginners-tutorials/tutorial-7-model-loading/
// TODO: texture coords
// TODO: material?
func LoadObj(filename string) (vertNormArr []gl.Float) {

	// open file
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Couldn't open %s", filename))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	// set up output slice
	m := StaticModelDef{
		verts:       make([]gl.Float, 0),
		vertIndices: make([]gl.Uint, 0),
		norms:       make([]gl.Float, 0),
		normIndices: make([]gl.Uint, 0),
	}

	// read line by line
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, " ")
		if split[0] == "v" {
			// v specifies a vert
			x, _ := strconv.ParseFloat(split[1], 32)
			y, _ := strconv.ParseFloat(split[2], 32)
			z, _ := strconv.ParseFloat(split[3], 32)
			m.verts = append(m.verts, gl.Float(x), gl.Float(y), gl.Float(z))
		} else if split[0] == "vn" {
			x, _ := strconv.ParseFloat(split[1], 32)
			y, _ := strconv.ParseFloat(split[2], 32)
			z, _ := strconv.ParseFloat(split[3], 32)
			m.norms = append(m.norms, gl.Float(x), gl.Float(y), gl.Float(z))
		} else if split[0] == "f" {
			// f specifies a face
			split0 := strings.Split(split[1], "/")
			split1 := strings.Split(split[2], "/")
			split2 := strings.Split(split[3], "/")

			vertIxA, _ := strconv.Atoi(split0[0])
			vertIxB, _ := strconv.Atoi(split1[0])
			vertIxC, _ := strconv.Atoi(split2[0])
			m.vertIndices = append(m.vertIndices,
				gl.Uint(vertIxA-1),
				gl.Uint(vertIxB-1),
				gl.Uint(vertIxC-1))

			// TODO: texture uv

			normIxA, _ := strconv.Atoi(split0[2])
			normIxB, _ := strconv.Atoi(split1[2])
			normIxC, _ := strconv.Atoi(split2[2])
			m.normIndices = append(m.normIndices,
				gl.Uint(normIxA-1),
				gl.Uint(normIxB-1),
				gl.Uint(normIxC-1))
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// populate the output float arr
	vertNormArr = make([]gl.Float, 0)
	/*
		8/11/7 vertex 1 of triangle: vertex 8, normal 7
		7/12/7 vertex 2 of triangle: vertex 7, normal 7
		6/10/7 vertex 3 of triangle: vertex 6, normal 7
	*/
	for i := 0; i < len(m.vertIndices); i++ {
		vIx := m.vertIndices[i]
		nIx := m.normIndices[i]
		vertNormArr = append(vertNormArr,
			m.verts[3*vIx],
			m.verts[3*vIx+1],
			m.verts[3*vIx+2],
			m.norms[3*nIx],
			m.norms[3*nIx+1],
			m.norms[3*nIx+2])
	}
	return vertNormArr
}
