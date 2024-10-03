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
