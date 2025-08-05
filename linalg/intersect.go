// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2025  Jochen Voss <voss@seehuhn.de>
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

package linalg

import (
	"math"

	"seehuhn.de/go/geom/vec"
)

// Intersection calculates the intersection point of two infinite line defined
// by points a1, a2 and b1, b2.  If the lines are parallel, it returns false.
// Otherwise, it returns the intersection point and true.
func Intersection(a1, a2, b1, b2 vec.Vec2) (vec.Vec2, bool) {
	denom := (a2.X-a1.X)*(b2.Y-b1.Y) - (a2.Y-a1.Y)*(b2.X-b1.X)
	if math.Abs(denom) < 1e-9 {
		return vec.Middle(a2, b1), false // Lines are parallel
	}
	t := ((b1.X-a1.X)*(b2.Y-b1.Y) - (b1.Y-a1.Y)*(b2.X-b1.X)) / denom
	intersection := vec.Vec2{
		X: a1.X + t*(a2.X-a1.X),
		Y: a1.Y + t*(a2.Y-a1.Y),
	}
	return intersection, true
}

// Miter calculates the miter point for two line segments (a,b) and (b,c). This
// is the intersection point of the outer edges of the line segments when they
// are stroked with the given line width.  The function returns false if (a,b)
// and (b,c) are parallel.  Otherwise, it returns the miter point and true.
func Miter(a, b, c vec.Vec2, lineWidth float64, outer bool) (vec.Vec2, bool) {
	ba := vec.Vec2{X: a.X - b.X, Y: a.Y - b.Y}
	bc := vec.Vec2{X: c.X - b.X, Y: c.Y - b.Y}
	baLength := ba.Length()
	bcLength := bc.Length()

	if baLength < 1e-9 || bcLength < 1e-9 {
		return b, false // Degenerate case
	}

	baNorm := ba.Normal()
	if baNorm.Dot(bc) > 0 == outer {
		baNorm.IMul(-1)
	}
	baNorm.IMul(lineWidth / 2)

	bcNorm := bc.Normal()
	if bcNorm.Dot(ba) > 0 == outer {
		bcNorm.IMul(-1)
	}
	bcNorm.IMul(lineWidth / 2)

	return Intersection(a.Add(baNorm), b.Add(baNorm), b.Add(bcNorm), c.Add(bcNorm))
}
