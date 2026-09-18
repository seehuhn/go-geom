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

package path

import (
	"math"
	"testing"

	"seehuhn.de/go/geom/vec"
)

// TestEvalCubicEndpoints checks that the ends of the curve are reproduced
// exactly, since a caller which joins segments end to end relies on it.
func TestEvalCubicEndpoints(t *testing.T) {
	p0 := vec.Vec2{X: 1.5, Y: -2.25}
	c1 := vec.Vec2{X: 17, Y: 3}
	c2 := vec.Vec2{X: -4, Y: 11}
	p3 := vec.Vec2{X: 6.125, Y: 0.5}

	if got := EvalCubic(p0, c1, c2, p3, 0); got != p0 {
		t.Errorf("at t=0: got %v, want %v", got, p0)
	}
	if got := EvalCubic(p0, c1, c2, p3, 1); got != p3 {
		t.Errorf("at t=1: got %v, want %v", got, p3)
	}
}

// TestEvalCubicBernstein checks the interpolation against the Bernstein form
// of the same curve.
func TestEvalCubicBernstein(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	c1 := vec.Vec2{X: 1, Y: 3}
	c2 := vec.Vec2{X: 4, Y: -1}
	p3 := vec.Vec2{X: 5, Y: 2}

	for i := range 21 {
		u := float64(i) / 20
		s := 1 - u
		want := vec.Vec2{
			X: s*s*s*p0.X + 3*s*s*u*c1.X + 3*s*u*u*c2.X + u*u*u*p3.X,
			Y: s*s*s*p0.Y + 3*s*s*u*c1.Y + 3*s*u*u*c2.Y + u*u*u*p3.Y,
		}
		got := EvalCubic(p0, c1, c2, p3, u)
		if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
			t.Errorf("at t=%g: got %v, want %v", u, got, want)
		}
	}
}

// TestEvalCubicLine checks that a curve whose control points are spaced
// evenly along a straight line is traversed at a constant rate.
func TestEvalCubicLine(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	c1 := vec.Vec2{X: 1, Y: 2}
	c2 := vec.Vec2{X: 2, Y: 4}
	p3 := vec.Vec2{X: 3, Y: 6}

	for i := range 11 {
		u := float64(i) / 10
		got := EvalCubic(p0, c1, c2, p3, u)
		want := vec.Vec2{X: 3 * u, Y: 6 * u}
		if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
			t.Errorf("at t=%g: got %v, want %v", u, got, want)
		}
	}
}

// TestEvalCubicConvexHull checks that the curve stays inside the bounding box
// of its control points.
func TestEvalCubicConvexHull(t *testing.T) {
	p0 := vec.Vec2{X: -1, Y: 0}
	c1 := vec.Vec2{X: 0, Y: 100}
	c2 := vec.Vec2{X: 2, Y: -50}
	p3 := vec.Vec2{X: 3, Y: 1}

	bbox := DataFromPath(func(yield func(Command, []vec.Vec2) bool) {
		yield(CmdMoveTo, []vec.Vec2{p0})
		yield(CmdCubeTo, []vec.Vec2{c1, c2, p3})
	}).Iter().BBox()

	for i := range 101 {
		u := float64(i) / 100
		p := EvalCubic(p0, c1, c2, p3, u)
		if p.X < bbox.LLx || p.X > bbox.URx || p.Y < bbox.LLy || p.Y > bbox.URy {
			t.Errorf("at t=%g: point %v outside %v", u, p, bbox)
		}
	}
}
