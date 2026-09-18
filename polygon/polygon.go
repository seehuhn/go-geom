// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2026  Jochen Voss <voss@seehuhn.de>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package polygon implements operations on closed polygons.
//
// A polygon is given as the slice of its vertices, with an edge from each
// vertex to the next and a closing edge from the last vertex back to the
// first.  The functions here make no assumption about the polygon being
// simple: edges may cross, and vertices may repeat.
package polygon

import "seehuhn.de/go/geom/vec"

// Perimeter returns the total length of the polygon's edges.
// The result is zero for a polygon with fewer than two vertices.
func Perimeter(vertices []vec.Vec2) float64 {
	n := len(vertices)
	if n < 2 {
		return 0
	}
	var total float64
	for i := range n {
		total += vertices[(i+1)%n].Sub(vertices[i]).Length()
	}
	return total
}

// SignedArea returns the area enclosed by the polygon, counted positive where
// the vertices run counter-clockwise and negative where they run clockwise.
// The result is zero for a polygon with fewer than three vertices, and for a
// self-crossing polygon it is the sum over the enclosed regions, each weighted
// by its winding number.
func SignedArea(vertices []vec.Vec2) float64 {
	n := len(vertices)
	if n < 3 {
		return 0
	}

	// The polygon is split into triangles which all share its first vertex,
	// and each triangle is measured relative to that vertex.  The shift does
	// not change the area, but for a polygon which sits far from the origin
	// it keeps the products small: formed from the raw coordinates, they
	// would be much larger than their difference and would cancel away most
	// of its significant digits.
	origin := vertices[0]
	var twiceArea float64
	for i := 1; i < n-1; i++ {
		twiceArea += vertices[i].Sub(origin).Cross(vertices[i+1].Sub(origin))
	}
	return twiceArea / 2
}

// Resample returns n points spaced equally along the polygon's edges.  The
// first point returned is the first vertex, and the points follow the edges in
// the order the vertices are given in.  The spacing is the perimeter divided
// by n, so that the step from the last point back to the first matches the
// others.
//
// The result is nil if n is not positive, or if the polygon has fewer than two
// vertices or a perimeter which is not positive.  Two vertices are enough: the
// boundary then runs out along the single edge and back again, as it does for
// [Perimeter].
func Resample(vertices []vec.Vec2, n int) []vec.Vec2 {
	m := len(vertices)
	if m < 2 || n < 1 {
		return nil
	}

	edgeLen := make([]float64, m)
	var perimeter float64
	for i := range m {
		edgeLen[i] = vertices[(i+1)%m].Sub(vertices[i]).Length()
		perimeter += edgeLen[i]
	}
	if !(perimeter > 0) {
		return nil
	}

	step := perimeter / float64(n)
	points := make([]vec.Vec2, n)

	// The sample distances increase, so the edges are walked once for all
	// samples together rather than searched from the start for each one.
	// The walk stops at the last edge, which therefore absorbs any distance
	// left over by the rounding error of the accumulated edge lengths.
	e := 0
	edgeStart := 0.0
	for i := range n {
		dist := float64(i) * step
		for e < m-1 && dist >= edgeStart+edgeLen[e] {
			edgeStart += edgeLen[e]
			e++
		}
		var t float64
		if edgeLen[e] > 0 {
			t = min((dist-edgeStart)/edgeLen[e], 1)
		}
		a := vertices[e]
		b := vertices[(e+1)%m]
		points[i] = a.Add(b.Sub(a).Mul(t))
	}
	return points
}
