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

package polygon

import (
	"math"
	"testing"

	"seehuhn.de/go/geom/vec"
)

var unitSquare = []vec.Vec2{
	{X: 0, Y: 0},
	{X: 1, Y: 0},
	{X: 1, Y: 1},
	{X: 0, Y: 1},
}

func TestPerimeter(t *testing.T) {
	type testCase struct {
		name     string
		vertices []vec.Vec2
		want     float64
	}
	testCases := []testCase{
		{"nil", nil, 0},
		{"single vertex", []vec.Vec2{{X: 3, Y: 4}}, 0},
		{"two vertices", []vec.Vec2{{X: 0, Y: 0}, {X: 3, Y: 4}}, 10},
		{"unit square", unitSquare, 4},
		{"3-4-5 triangle", []vec.Vec2{{X: 0, Y: 0}, {X: 3, Y: 0}, {X: 0, Y: 4}}, 12},
		{"repeated vertex", []vec.Vec2{{X: 0, Y: 0}, {X: 0, Y: 0}, {X: 1, Y: 0}}, 2},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Perimeter(tc.vertices); math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("got %g, want %g", got, tc.want)
			}
		})
	}
}

func TestSignedArea(t *testing.T) {
	type testCase struct {
		name     string
		vertices []vec.Vec2
		want     float64
	}
	cw := make([]vec.Vec2, len(unitSquare))
	for i, v := range unitSquare {
		cw[len(unitSquare)-1-i] = v
	}
	testCases := []testCase{
		{"nil", nil, 0},
		{"two vertices", []vec.Vec2{{X: 0, Y: 0}, {X: 1, Y: 0}}, 0},
		{"unit square", unitSquare, 1},
		{"unit square reversed", cw, -1},
		{"collinear", []vec.Vec2{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 2, Y: 0}}, 0},
		{
			"triangle", []vec.Vec2{{X: 0, Y: 0}, {X: 4, Y: 0}, {X: 0, Y: 3}}, 6,
		},
		{
			// A figure of eight: the two loops wind in opposite directions,
			// so their areas cancel.
			"figure of eight",
			[]vec.Vec2{
				{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1},
			},
			0,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := SignedArea(tc.vertices); math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("got %g, want %g", got, tc.want)
			}
		})
	}
}

// TestSignedAreaTranslation checks that moving the polygon far from the origin
// does not change the area it reports.  Formed from the raw coordinates, the
// products would be much larger than their difference and would cancel away
// its significant digits.
func TestSignedAreaTranslation(t *testing.T) {
	for _, offset := range []float64{0, 1, 1e6, 1e10} {
		moved := make([]vec.Vec2, len(unitSquare))
		for i, v := range unitSquare {
			moved[i] = vec.Vec2{X: v.X + offset, Y: v.Y + offset}
		}
		if got := SignedArea(moved); math.Abs(got-1) > 1e-9 {
			t.Errorf("offset %g: got %g, want 1", offset, got)
		}
	}
}

func TestResampleDegenerate(t *testing.T) {
	type testCase struct {
		name     string
		vertices []vec.Vec2
		n        int
	}
	testCases := []testCase{
		{"nil", nil, 4},
		{"single vertex", []vec.Vec2{{X: 1, Y: 2}}, 4},
		{"zero count", unitSquare, 0},
		{"negative count", unitSquare, -1},
		{"zero perimeter", []vec.Vec2{{}, {}, {}}, 4},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Resample(tc.vertices, tc.n); got != nil {
				t.Errorf("got %v, want nil", got)
			}
		})
	}
}

// TestResampleTwoVertices checks the degenerate polygon which runs out along
// a single edge and back again.
func TestResampleTwoVertices(t *testing.T) {
	vertices := []vec.Vec2{{X: 0, Y: 0}, {X: 1, Y: 0}}
	want := []vec.Vec2{
		{X: 0, Y: 0}, {X: 0.5, Y: 0}, {X: 1, Y: 0}, {X: 0.5, Y: 0},
	}
	got := Resample(vertices, len(want))
	if len(got) != len(want) {
		t.Fatalf("got %d points, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i].X-want[i].X) > 1e-12 || math.Abs(got[i].Y-want[i].Y) > 1e-12 {
			t.Errorf("point %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestResampleSquare(t *testing.T) {
	// Eight points round a unit square land on the corners and edge midpoints.
	want := []vec.Vec2{
		{X: 0, Y: 0}, {X: 0.5, Y: 0},
		{X: 1, Y: 0}, {X: 1, Y: 0.5},
		{X: 1, Y: 1}, {X: 0.5, Y: 1},
		{X: 0, Y: 1}, {X: 0, Y: 0.5},
	}
	got := Resample(unitSquare, len(want))
	if len(got) != len(want) {
		t.Fatalf("got %d points, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i].X-want[i].X) > 1e-12 || math.Abs(got[i].Y-want[i].Y) > 1e-12 {
			t.Errorf("point %d: got %v, want %v", i, got[i], want[i])
		}
	}
}

// TestResampleSpacing checks that consecutive points, including the pair which
// straddles the closing edge, are one step apart along the boundary.
func TestResampleSpacing(t *testing.T) {
	vertices := []vec.Vec2{
		{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 13, Y: 4}, {X: 4, Y: 7}, {X: -2, Y: 3},
	}
	for _, n := range []int{1, 2, 3, 5, 17, 100} {
		points := Resample(vertices, n)
		if len(points) != n {
			t.Fatalf("n=%d: got %d points", n, len(points))
		}
		if points[0] != vertices[0] {
			t.Errorf("n=%d: first point is %v, want %v", n, points[0], vertices[0])
		}

		// The straight-line distance between neighbours is at most the step,
		// with equality where the two lie on the same edge.
		step := Perimeter(vertices) / float64(n)
		for i := range n {
			d := points[(i+1)%n].Sub(points[i]).Length()
			if d > step+1e-9 {
				t.Errorf("n=%d: points %d and %d are %g apart, step is %g",
					n, i, (i+1)%n, d, step)
			}
		}
	}
}

// TestResampleOnBoundary checks that every point returned lies on one of the
// polygon's edges.
func TestResampleOnBoundary(t *testing.T) {
	vertices := []vec.Vec2{
		{X: 0, Y: 0}, {X: 6, Y: 1}, {X: 2, Y: 9}, {X: -3, Y: 2},
	}
	points := Resample(vertices, 37)
	for i, p := range points {
		best := math.Inf(1)
		for e := range vertices {
			a := vertices[e]
			b := vertices[(e+1)%len(vertices)]
			ab := b.Sub(a)
			t := max(0, min(1, p.Sub(a).Dot(ab)/ab.Dot(ab)))
			best = min(best, p.Sub(a.Add(ab.Mul(t))).Length())
		}
		if best > 1e-12 {
			t.Errorf("point %d (%v) is %g away from the boundary", i, p, best)
		}
	}
}

// TestResampleRepeatedVertex checks that a zero-length edge does not divide by
// zero or swallow a sample.
func TestResampleRepeatedVertex(t *testing.T) {
	vertices := []vec.Vec2{
		{X: 0, Y: 0}, {X: 0, Y: 0}, {X: 4, Y: 0}, {X: 4, Y: 3},
	}
	points := Resample(vertices, 6)
	if len(points) != 6 {
		t.Fatalf("got %d points, want 6", len(points))
	}
	for i, p := range points {
		if math.IsNaN(p.X) || math.IsNaN(p.Y) {
			t.Errorf("point %d is %v", i, p)
		}
	}
}
