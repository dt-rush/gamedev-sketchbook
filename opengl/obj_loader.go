package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	gl "github.com/chsc/gogl/gl43"
)

// from http://www.opengl-tutorial.org/beginners-tutorials/tutorial-7-model-loading/
// TODO: texture coords
// TODO: normals
// TODO: material?
func LoadObjVerts(filename string) (verts []gl.Float, indices []gl.Uint) {

	// open file
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Sprintf("Couldn't open %s", filename))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)

	// set up output slice
	verts = make([]gl.Float, 0)
	indices = make([]gl.Uint, 0)

	// read line by line
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Split(line, " ")
		if split[0] == "v" {
			// v specifies a vert
			x, _ := strconv.ParseFloat(split[1], 32)
			y, _ := strconv.ParseFloat(split[2], 32)
			z, _ := strconv.ParseFloat(split[3], 32)
			verts = append(verts, gl.Float(x), gl.Float(y), gl.Float(z))
		} else if split[0] == "f" {
			// f specifies a face
			vertIndices := [3]int{}
			vertIndices[0], _ = strconv.Atoi(strings.Split(split[1], "/")[0])
			vertIndices[1], _ = strconv.Atoi(strings.Split(split[2], "/")[0])
			vertIndices[2], _ = strconv.Atoi(strings.Split(split[3], "/")[0])
			// fmt.Println(vertIndices[0], vertIndices[1], vertIndices[2])
			indices = append(indices, gl.Uint(vertIndices[0]-1), gl.Uint(vertIndices[1]-1), gl.Uint(vertIndices[2]-1))
		}
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	return verts, indices
}
