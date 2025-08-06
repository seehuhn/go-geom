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

package vec

import "math"

type Vec2 struct {
	X, Y float64
}

// Length returns the length of the vector.
func (v Vec2) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

func (v Vec2) Mul(c float64) Vec2 {
	return Vec2{X: v.X * c, Y: v.Y * c}
}

func (v *Vec2) IMul(c float64) {
	v.X *= c
	v.Y *= c
}

// Dot calculates the dot product of two vectors.
func (v Vec2) Dot(other Vec2) float64 {
	return v.X*other.X + v.Y*other.Y
}

// Normal returns a unit vector perpendicular to v, rotated 90 degrees counter-clockwise.
func (v Vec2) Normal() Vec2 {
	length := v.Length()
	if length < 1e-9 {
		return Vec2{0, 0}
	}
	return Vec2{X: -v.Y / length, Y: v.X / length}
}

func Middle(a, b Vec2) Vec2 {
	return Vec2{X: (a.X + b.X) / 2, Y: (a.Y + b.Y) / 2}
}
