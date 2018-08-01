package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"math"
	"sort"
	"time"
)

type WorldMap struct {
	Lakes         []*Lake
	Vertices      map[int]*MapVertex
	perlin        [][]float64
	perlinTexture *sdl.Texture
	minima        []Point2D
	seed          int64
	param         int
}

func GenerateWorldMap(r *sdl.Renderer) *WorldMap {
	seed := time.Now().UnixNano()
	// seed = 1528868059045470378
	// seed = 1528907672650396933
	seed = 1528917396999568729
	fmt.Println(seed)
	m := WorldMap{seed: seed, param: 0}
	m.Vertices = make(map[int]*MapVertex)
	m.generatePerlin()
	m.perlinTexture = CreatePerlinTexture(r, m.perlin)
	m.findMinima()
	fmt.Println("finished finding minima")
	m.makeLakes()
	return &m
}

func (wm *WorldMap) makeLakes() {
	rawLakes := make([]*Lake, 0)
	for id, min := range wm.minima {
		l := wm.MakeLake(id, min)
		l.buildVXVY()
		if l != nil {
			rawLakes = append(rawLakes, l)
		}
	}
	wm.Lakes = rawLakes
	// wm.Lakes = wm.mergedLakes(rawLakes)
}

func (wm *WorldMap) MSmergedLakes(rawLakes []*Lake) []*Lake {
	pointGroups := make(map[int][]Point2D)
	merged := make(map[*Lake]bool)
	for _, l1 := range rawLakes {
		if merged[l1] {
			continue
		}
		pointGroups[l1.id] = l1.Vertices
		for _, l2 := range wm.Lakes {
			if l1 == l2 || merged[l2] {
				continue
			}
			for _, v := range l2.Vertices {
				if l1.containsPoint2D(v) {
					merged[l2] = true
					pointGroups[l1.id] = append(pointGroups[l1.id], l2.Vertices...)
					break
				}
			}
		}
	}
	lakes := make([]*Lake, 0)
	for i, pg := range pointGroups {
		l := Lake{id: i}
		floatyBois := make([][2]float64, len(pg))
		for i, v := range pg {
			floatyBois[i] = [2]float64{float64(v.X), float64(v.Y)}
		}
		hull := concave(floatyBois, 3)
		for _, p := range hull {
			l.Vertices = append(l.Vertices,
				Point2D{
					int(math.Min(WORLD_WIDTH-1, math.Max(0, p[0]))),
					int(math.Min(WORLD_HEIGHT-1, math.Max(0, p[1])))})
		}
		l.buildVXVY()
		lakes = append(lakes, &l)
	}
	return lakes
}

func (wm *WorldMap) mergedLakes(rawLakes []*Lake) []*Lake {
	pointGroups := make(map[int][]Point2D)
	merged := make(map[*Lake]bool)
	for _, l1 := range rawLakes {
		if merged[l1] {
			continue
		}
		pointGroups[l1.id] = l1.Vertices
		for _, l2 := range wm.Lakes {
			if l1 == l2 || merged[l2] {
				continue
			}
			for _, v := range l2.Vertices {
				if l1.containsPoint2D(v) {
					merged[l2] = true
					pointGroups[l1.id] = append(pointGroups[l1.id], l2.Vertices...)
					break
				}
			}
		}
	}
	lakes := make([]*Lake, 0)
	for i, pg := range pointGroups {
		l := Lake{id: i}
		flatPoints := make([]float64, 2*len(pg))
		for i, p := range pg {
			flatPoints[2*i] = float64(p.X)
			flatPoints[2*i+1] = float64(p.Y)
		}
		hull := Compute(flatPoints)
		for i := 0; i < len(hull); i += 2 {
			l.Vertices = append(l.Vertices,
				Point2D{
					int(math.Min(WORLD_WIDTH-1, math.Max(0, hull[i]))),
					int(math.Min(WORLD_HEIGHT-1, math.Max(0, hull[i+1])))})
			l.interpolated = append(l.interpolated, false)
		}
		l.buildVXVY()
		lakes = append(lakes, &l)
	}
	return lakes
}

func (wm *WorldMap) Regen(param int) {
	wm.param = param
	highlighted := -1
	for _, l := range wm.Lakes {
		if l.highlighted {
			highlighted = l.id
		}
	}
	wm.Lakes = wm.Lakes[:0]
	wm.makeLakes()
	for _, l := range wm.Lakes {
		if l.id == highlighted {
			l.highlighted = true
		}
	}
}

func (wm *WorldMap) generatePerlin() {
	terrain := PerlinNoiseInt2D(
		PW, PH, 16.0,
		2.0, 2.0, 3,
		wm.seed)
	water := PerlinNoiseInt2D(
		PW, PH, 32,
		4.0, 2.0, 3,
		wm.seed)
	combined := OpPerlins(terrain, water, func(a float64, b float64) float64 {
		x := (a + (a + 0.3) - b) / 2
		if x < 0 {
			return 0
		} else if x > 1 {
			return 1
		} else {
			return x
		}
	})
	wm.perlin = SmoothPerlin(combined)
}

func (wm *WorldMap) findMinima() {
	// start at numerous random positions around the map, and iterate until we
	// reach a local minimum. The local minimum which appears most often is the
	// one we are going to start our lake at
	minimums := make(map[Point2D]int)
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			min := wm.findLocalMinimum(
				Point2D{
					x * (WORLD_WIDTH / 8), y * (WORLD_HEIGHT / 8)}, 64)
			minimums[min]++
		}
	}
	type PointFreq struct {
		p    Point2D
		freq int
	}
	pfs := make([]PointFreq, 0)
	for p, freq := range minimums {
		pfs = append(pfs, PointFreq{p, freq})
	}
	sort.Slice(pfs, func(i, j int) bool {
		return pfs[i].freq < pfs[j].freq
	})
	for i := 0; i < len(pfs); i++ {
		pf := pfs[len(pfs)-1-i]
		wm.minima = append(wm.minima, pf.p)
	}
}
func (wm *WorldMap) findLocalMinimum(p Point2D, scan int) Point2D {
	for {
		var descendPoint Point2D
		lowest := 1e9
		step := 1
		for iy := -scan * step; iy <= scan*step; iy += step {
			py := int(float64(p.Y+iy) / PSCALE)
			if !(py > 0 && py < PH-1) {
				continue
			}
			for ix := -scan * step; ix <= scan*step; ix += step {
				px := int(float64(p.X+ix) / PSCALE)
				if !(px > 0 && px < PW-1) {
					continue
				}
				pval := wm.perlin[py][px]
				if pval < lowest {
					lowest = pval
					descendPoint = Point2D{p.X + ix, p.Y + iy}
				}
			}
		}
		dx, dy, d := Distance(p, descendPoint)
		if d >= 4 {
			dx = 4 * dx / d
			dy = 4 * dy / d
			p.X = int(math.Min(WORLD_WIDTH-1, math.Max(float64(p.X)+dx, 0)))
			p.Y = int(math.Min(WORLD_HEIGHT-1, math.Max(float64(p.Y)+dy, 0)))
		} else {
			return descendPoint
		}
	}
}

func (wm *WorldMap) elevationAt(p Point2D) float64 {
	py := int(float64(p.Y) / PSCALE)
	if py < 0 {
		py = 0
	} else if py > PH-1 {
		py = PH - 1
	}
	px := int(float64(p.X) / PSCALE)
	if px < 0 {
		px = 0
	} else if px > PW-1 {
		px = PW - 1
	}
	return wm.perlin[py][px]
}

func (wm *WorldMap) MakeLake(id int, p Point2D) *Lake {

	l := Lake{id: id, source: p}

	// make the seed for a lake
	nVertices := 4
	radius := 16.0
	vertices := make([]Point2D, 0)
	interpolated := make([]bool, 0)

	var createInitialVertices = func() {
		for i := 0; i < nVertices; i++ {
			theta := float64(i) * (2 * math.Pi / float64(nVertices))
			vp := Point2D{
				p.X + int(radius*math.Sin(theta)),
				p.Y + int(radius*math.Cos(theta))}
			vertices = append(vertices, vp)
			interpolated = append(interpolated, false)
		}
	}
	createInitialVertices()

	var expandVerticesToShore = func() {
		for i := 0; i < len(vertices); i++ {
			v := &vertices[i]
			if math.Abs(wm.elevationAt(*v)-WATER_CUTOFF) < 0.04 {
				continue
			}
			var before Point2D
			j := i - 1
			for {
				if !interpolated[(j+len(vertices))%len(vertices)] {
					before = vertices[(j+len(vertices))%len(vertices)]
					break
				}
				j--
			}
			var after Point2D
			j = i + 1
			for {
				if !interpolated[j%len(vertices)] {
					after = vertices[j%len(vertices)]
					break
				}
				j++
			}
			// after/before are reversed because world-space to screen-space
			// flips the right-hand rule (wtf)
			out := VecFromPoints(after, before).PerpendicularUnit()
			for wm.moveToPointIfValidWater(v,
				Point2D{v.X + int(4*out.X),
					v.Y + int(4*out.Y)}) {
			}
		}
		for i := 0; i < len(vertices); i++ {
			interpolated[i] = false
		}
	}
	var interpolateVertices = func(segL float64) func() {
		type interpolatedVertex struct {
			v            Point2D
			interpolated bool
		}
		return func() {
			exp := make([]interpolatedVertex, 0)
			for i, vertex := range vertices {
				exp = append(exp, interpolatedVertex{vertex, false})
				next := vertices[(i+1)%len(vertices)]
				dnx, dny, d := Distance(vertex, next)
				if d < 8 {
					continue
				} else {
					div := d / segL
					for segment := 0; segment < int(div); segment++ {
						incNext := PointDelta(vertex,
							segment*int(dnx/div)+int(dnx/div),
							segment*int(dny/div)+int(dny/div))
						midpoint := Vec2D{
							float64(incNext.X),
							float64(incNext.Y)}
						// if our new midpoint's elevation is on land,
						// retract it inward

						exp = append(exp,
							interpolatedVertex{midpoint.ToPoint(), true})
					}
				}
			}
			vertices = vertices[:0]
			interpolated = interpolated[:0]
			for _, iv := range exp {
				vertices = append(vertices, iv.v)
				interpolated = append(interpolated, iv.interpolated)
			}
		}
	}
	var moveInterpolatedVertices = func() {
		for i, _ := range vertices {
			var before Point2D
			j := i - 1
			for {
				if !interpolated[(j+len(vertices))%len(vertices)] {
					before = vertices[(j+len(vertices))%len(vertices)]
					break
				}
				j--
			}
			var after Point2D
			j = i + 1
			for {
				if !interpolated[j%len(vertices)] {
					after = vertices[j%len(vertices)]
					break
				}
				j++
			}
			vx := float64(vertices[i].X)
			vy := float64(vertices[i].Y)
			for wm.elevationAt(Point2D{int(vx), int(vy)}) > WATER_CUTOFF &&
				vx > 0 && vy > 0 &&
				vx < WORLD_WIDTH-1 &&
				vy < WORLD_HEIGHT-1 {

				in := VecFromPoints(before, after).PerpendicularUnit()
				vx += in.X
				vy += in.Y
			}
			vertices[i].X = int(math.Min(WORLD_WIDTH-1, math.Max(0, vx)))
			vertices[i].Y = int(math.Min(WORLD_HEIGHT-1, math.Max(0, vy)))
		}
	}
	var simplify = func() {
		flatPoints := make([]float64, 2*len(vertices))
		for i, v := range vertices {
			flatPoints[2*i] = float64(v.X)
			flatPoints[2*i+1] = float64(v.Y)
		}
		hull := Compute(flatPoints)
		vertices = vertices[:0]
		interpolated = interpolated[:0]
		for i := 0; i < len(hull); i += 2 {
			vertices = append(vertices,
				Point2D{
					int(math.Min(WORLD_WIDTH-1, math.Max(0, hull[i]))),
					int(math.Min(WORLD_HEIGHT-1, math.Max(0, hull[i+1])))})
			interpolated = append(interpolated, false)
		}
	}
	var MSsimplify = func() {
		floatyBois := make([][2]float64, len(vertices))
		for i, v := range vertices {
			floatyBois[i] = [2]float64{float64(v.X), float64(v.Y)}
		}
		hull := concave(floatyBois, 3)
		// fmt.Println(hull)
		vertices = vertices[:0]
		interpolated = interpolated[:0]
		for _, p := range hull {
			vertices = append(vertices,
				Point2D{
					int(math.Min(WORLD_WIDTH-1, math.Max(0, p[0]))),
					int(math.Min(WORLD_HEIGHT-1, math.Max(0, p[1])))})
			interpolated = append(interpolated, false)
		}
	}
	fmt.Println(simplify)
	fmt.Println(MSsimplify)
	seq := []func(){
		expandVerticesToShore,
		interpolateVertices(16),
		moveInterpolatedVertices,

		expandVerticesToShore,
		interpolateVertices(16),
		moveInterpolatedVertices,

		expandVerticesToShore,
		interpolateVertices(16),
		moveInterpolatedVertices,

		expandVerticesToShore,
		interpolateVertices(16),
		moveInterpolatedVertices,
	}
	for i := 0; i < wm.param && i < len(seq); i++ {
		seq[i]()
	}

	l.Vertices = vertices
	l.interpolated = interpolated
	if len(vertices) > 0 {
		return &l
	}
	return nil
}

func (wm *WorldMap) moveToPointIfValidWater(p *Point2D, next Point2D) bool {
	next.X = int(math.Min(WORLD_WIDTH-1, math.Max(float64(next.X), 0)))
	next.Y = int(math.Min(WORLD_HEIGHT-1, math.Max(float64(next.Y), 0)))
	if wm.elevationAt(next) <= WATER_CUTOFF {
		*p = next
		if p.X == 0 || p.Y == 0 ||
			p.X == WORLD_WIDTH-1 || p.Y == WORLD_HEIGHT-1 {
			return false
		}
		return true
	}
	return false
}

func (wm *WorldMap) BuildLineOfSightNetwork() {
	vertices := make([]MapVertex, 0)
	// build list of all lake vertices not inside other lakes
	for i, lake := range wm.Lakes {
		for _, vertex := range lake.Vertices {
			pointInOtherLakes := false
			for j := 0; j < len(wm.Lakes); j++ {
				if j == i {
					continue
				}
				if wm.Lakes[j].containsPoint2D(vertex) {
					pointInOtherLakes = true
					break
				}
			}
			if !pointInOtherLakes {
				vertices = append(vertices, MapVertex{pos: vertex})
			}
		}
	}
}
