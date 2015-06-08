package main

import (
	"fmt"
)

type vertex struct {
	x float64
	y float64
	isEar bool
}

type Triangle struct {
	Vertices [3]vertex
}

type Polygon struct {
	Vertices []vertex
}

func makePolygon(points []float64) Polygon {
	vs := make([]vertex, len(points)/2)
/*
	for i, j := len(points) - 2, 0; i >= 0; i, j = i - 2, j + 1 {
		vs[j] = vertex{points[i], points[i + 1], false}
	}
*/
	for i := 0; i < len(points); i += 2 {
		vs[i/2] = vertex{points[i], points[i + 1], false}
	}

	return Polygon{vs}
}

func Area(v []vertex) float64 {
	var sum float64
	for i := 0; i < len(v) - 1; i++ {
		sum += (v[i].x + v[i + 1].x)*(v[i].y - v[i +  1].y)
	}

	return sum/2
}

func left(a, b, c vertex) bool {
	if Area([]vertex{a, b, c}) > 0 {
		return true
	}

	return false
}

func leftOn(a, b, c vertex) bool {
	if Area([]vertex{a, b, c}) >= 0 {
		return true
	}

	return false
}

func collinear(a, b, c vertex) bool {
	if Area([]vertex{a, b, c}) == 0 {
		return true
	}

	return false
}

func between(a, b, c vertex) bool {
	if !collinear(a, b, c) {
		return false
	}

	if a.x != b.x {
		return ((a.x <= c.x) && (c.x <= b.x)) || ((a.x >= c.x) && (c.x >= b.x))
	} else {
		return ((a.y <= c.y) && (c.y <= b.y)) || ((a.y >= c.y) && (c.y >= b.y))
	}
}

func intersectProp(a, b, c, d vertex) bool {
	if collinear(a, b, c) || collinear(a, b, d) || collinear(c, d, a) || collinear(c, d, b) {
		return false
	}

	return (left(a, b, c) != left(a, b, d)) || (left(c, d, a) != left(c, d, b))
}

func intersect(a, b, c, d vertex) bool {
	if intersectProp(a, b, c, d) {
		return true
	}

	if between(a, b ,c) || between(a, b, d) || between(c, d, a) || between(c, d, b) {
		return true
	}

	return false
}

func diagonalie(polygon *Polygon, aIdx, bIdx int) bool {
	fmt.Printf("diagonalie len: %d, aIdx: %d, bIdx: %d\n", len(polygon.Vertices), aIdx, bIdx)
	for c, _ := range polygon.Vertices {
		c1 := (c + 1) % len(polygon.Vertices)
		if c != aIdx && c1 != aIdx && c != bIdx && c1 != bIdx && intersect(polygon.Vertices[aIdx], polygon.Vertices[bIdx], polygon.Vertices[c], polygon.Vertices[c1]) {
			return false
		}
	}

	return true
}

func inCone(polygon *Polygon, aIdx, bIdx int) bool {
	a0 := aIdx - 1
	a1 := (aIdx + 1) % len(polygon.Vertices)

	if a0 < 0 {
		a0 = len(polygon.Vertices) - 1
	}

	fmt.Printf("inCone len: %d, a0: %d, a1: %d\n", len(polygon.Vertices), a0, a1)

	if leftOn(polygon.Vertices[aIdx], polygon.Vertices[a1], polygon.Vertices[a0]) {
		return left(polygon.Vertices[aIdx], polygon.Vertices[bIdx], polygon.Vertices[a0]) && left(polygon.Vertices[bIdx], polygon.Vertices[aIdx], polygon.Vertices[a1])
	}

	return !(leftOn(polygon.Vertices[aIdx], polygon.Vertices[bIdx], polygon.Vertices[a1]) && leftOn(polygon.Vertices[bIdx], polygon.Vertices[aIdx], polygon.Vertices[a0]))
}

func diagonal(polygon *Polygon, aIdx, bIdx int) bool {
	fmt.Printf("diagonal len: %d, aIdx: %d, bIdx: %d\n", len(polygon.Vertices), aIdx, bIdx)
	return inCone(polygon, aIdx, bIdx) && inCone(polygon, bIdx, aIdx) && diagonalie(polygon, aIdx, bIdx)
}

func initEar(polygon *Polygon) {
	for v1, _ := range polygon.Vertices {
		v0 := v1 - 1
		if v0 < 0 {
			v0 = len(polygon.Vertices) - 1
		}

		v2 := (v1 + 1) % len(polygon.Vertices)

		ear := diagonal(polygon, v0, v2)
		fmt.Printf("initEar v0: %d, v2: %d, ear: %t\n", v0, v2, ear)
		polygon.Vertices[v1].isEar = ear
	}
}

func Triangulate(polygon Polygon) []Triangle {
	res := []Triangle{}
	initEar(&polygon)
	vs := polygon.Vertices[:]

	fmt.Println("vs:", vs)
	for len(vs) > 3 {
//		fmt.Printf("triangulate loop vs len: %d\n", len(vs))
		for v2, _ := range vs {
			if vs[v2].isEar == false {
				continue
			}

			v3 := (v2 + 1) % len(vs)
			v4 := (v2 + 2) % len(vs)

			v1 := v2 - 1
			v0 := v2 - 2

			if v1 < 0 {
				v1 = len(vs) - 2
			}

			if v0 < 0 {
				v0 = len(vs) - 2
			}

			res = append(res, Triangle{[3]vertex{vs[v1], vs[v2], vs[v3]}})

			vs[v1].isEar = diagonal(&Polygon{vs}, v0, v3)
			vs[v3].isEar = diagonal(&Polygon{vs}, v1, v4)

			vs = append(vs[:v2], vs[v3:]...)
		}
	}

	return res
}
