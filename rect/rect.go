// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2024  Jochen Voss <voss@seehuhn.de>
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

package rect

import (
	"math"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/geom/vec"
)

// Rect represents an axis-aligned rectangle.
type Rect struct {
	LLx, LLy, URx, URy float64
}

// IsZero reports whether the rectangle is the zero value.
func (r Rect) IsZero() bool {
	return r.LLx == 0 && r.LLy == 0 && r.URx == 0 && r.URy == 0
}

// Dx returns the width of the rectangle.
func (r Rect) Dx() float64 {
	return r.URx - r.LLx
}

// Dy returns the height of the rectangle.
func (r Rect) Dy() float64 {
	return r.URy - r.LLy
}

// Covers reports whether the rectangle contains all of other, counting a
// shared edge as covered.
func (r Rect) Covers(other Rect) bool {
	return r.LLx <= other.LLx && r.LLy <= other.LLy && r.URx >= other.URx && r.URy >= other.URy
}

// Add enlarges the rectangle if necessary to include the point (x, y).
func (r *Rect) Add(x, y float64) {
	if x < r.LLx {
		r.LLx = x
	}
	if y < r.LLy {
		r.LLy = y
	}
	if x > r.URx {
		r.URx = x
	}
	if y > r.URy {
		r.URy = y
	}
}

// Extend enlarges the rectangle if necessary to include all of other.  The
// zero value stands for an empty rectangle here rather than for a point at the
// origin: extending by it changes nothing, and extending it gives other.
func (r *Rect) Extend(other Rect) {
	if other.IsZero() {
		return
	}
	if r.IsZero() {
		*r = other
		return
	}
	if other.LLx < r.LLx {
		r.LLx = other.LLx
	}
	if other.LLy < r.LLy {
		r.LLy = other.LLy
	}
	if other.URx > r.URx {
		r.URx = other.URx
	}
	if other.URy > r.URy {
		r.URy = other.URy
	}
}

// Scale multiplies all four coordinates by factor, moving the rectangle
// towards or away from the origin as well as changing its size.  A negative
// factor reflects the rectangle through the origin and turns it inside out.
func (r *Rect) Scale(factor float64) {
	r.LLx *= factor
	r.LLy *= factor
	r.URx *= factor
	r.URy *= factor
}

// Grow returns the rectangle enlarged by d on each of its four sides, so that
// it becomes 2*d wider and 2*d taller.  A negative d shrinks the rectangle
// instead, and shrinking by more than half the width or height turns the
// rectangle inside out rather than collapsing it to an empty one.
func (r Rect) Grow(d float64) Rect {
	return Rect{
		LLx: r.LLx - d,
		LLy: r.LLy - d,
		URx: r.URx + d,
		URy: r.URy + d,
	}
}

// Rounded returns the smallest rectangle with integer coordinates which
// contains this one.
func (r Rect) Rounded() Rect {
	r.LLx = math.Floor(r.LLx)
	r.LLy = math.Floor(r.LLy)
	r.URx = math.Ceil(r.URx)
	r.URy = math.Ceil(r.URy)
	return r
}

// Transform maps the four corners of r through M and returns their axis-aligned
// bounding box.  When M rotates or shears, the mapped rectangle is no longer
// axis-aligned, so the result is the smallest Rect that contains it.
func (r Rect) Transform(M matrix.Matrix) Rect {
	p0 := M.Apply(vec.Vec2{X: r.LLx, Y: r.LLy})
	p1 := M.Apply(vec.Vec2{X: r.URx, Y: r.LLy})
	p2 := M.Apply(vec.Vec2{X: r.LLx, Y: r.URy})
	p3 := M.Apply(vec.Vec2{X: r.URx, Y: r.URy})
	return Rect{
		LLx: min(p0.X, p1.X, p2.X, p3.X),
		LLy: min(p0.Y, p1.Y, p2.Y, p3.Y),
		URx: max(p0.X, p1.X, p2.X, p3.X),
		URy: max(p0.Y, p1.Y, p2.Y, p3.Y),
	}
}

// IntRect represents an axis-aligned rectangle with integer coordinates.
// This is used for pixel-based operations such as image dimensions.
//
// The rectangle is half-open: it contains the points (x, y) with
// XMin <= x < XMax and YMin <= y < YMax.  A rectangle with XMin >= XMax
// or YMin >= YMax is empty.
type IntRect struct {
	XMin, YMin, XMax, YMax int
}

// Dx returns the width of the rectangle.
func (r IntRect) Dx() int {
	return r.XMax - r.XMin
}

// Dy returns the height of the rectangle.
func (r IntRect) Dy() int {
	return r.YMax - r.YMin
}
