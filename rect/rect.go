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

func (r *Rect) Scale(factor float64) {
	r.LLx *= factor
	r.LLy *= factor
	r.URx *= factor
	r.URy *= factor
}

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
