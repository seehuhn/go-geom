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

import "seehuhn.de/go/geom/vec"

// EvalCubic evaluates the cubic Bezier curve with start point p0, control
// points c1 and c2, and end point p3 at the parameter t, which runs from 0 at
// p0 to 1 at p3.  The curve is defined outside that range as well, and the
// function extrapolates there.
//
// The curve is evaluated by repeated linear interpolation, using de
// Casteljau's algorithm.  Every intermediate point is a weighted average of
// two others, with weights which are non-negative for t in [0, 1], so the
// result never leaves the convex hull of the four given points.  It is
// exactly p0 at t == 0 and exactly p3 at t == 1.
func EvalCubic(p0, c1, c2, p3 vec.Vec2, t float64) vec.Vec2 {
	a := lerp(p0, c1, t)
	b := lerp(c1, c2, t)
	c := lerp(c2, p3, t)
	return lerp(lerp(a, b, t), lerp(b, c, t), t)
}

// lerp interpolates linearly between a and b.
func lerp(a, b vec.Vec2, t float64) vec.Vec2 {
	s := 1 - t
	return vec.Vec2{
		X: s*a.X + t*b.X,
		Y: s*a.Y + t*b.Y,
	}
}
