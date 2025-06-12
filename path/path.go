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

package path

import "iter"

type Point struct {
	X, Y float64
}

type Command byte

const (
	CmdMoveTo Command = iota + 1
	CmdLineTo
	CmdQuadTo
	CmdCubeTo
	CmdClose
)

// Path iterates over individual segments.
// The first argument is the segment type, which determines the number of points:
//   - CmdMoveTo: 1 point (starts a new sub-path at the given point)
//   - CmdLineTo: 1 point (endpoint of the line segment)
//   - CmdQuadTo: 2 points (control point and endpoint of the quadratic Bezier curve)
//   - CmdCubeTo: 3 points (control point 1, control point 2, and endpoint of the cubic Bezier curve)
//   - CmdClose: 0 points (straight line from the current point to start of the current sub-path)
type Path iter.Seq2[Command, []Point]

// ToCubic converts all quadratic segments in the path to cubic segments.
// The resulting representation describes exactly the same path,
// but is less efficient (since two control points are used instead of one).
func ToCubic(p Path) Path {
	// https://pomax.github.io/bezierinfo/#reordering
	return func(yield func(Command, []Point) bool) {
		var current Point
		var start Point
		var buf [3]Point
		for cmd, pts := range p {
			if cmd == CmdQuadTo {
				buf[0].X = current.X*1/3 + pts[0].X*2/3
				buf[0].Y = current.Y*1/3 + pts[0].Y*2/3
				buf[1].X = pts[0].X*2/3 + pts[1].X*1/3
				buf[1].Y = pts[0].Y*2/3 + pts[1].Y*1/3
				buf[2] = pts[1]
				if !yield(CmdCubeTo, buf[:]) {
					return
				}
			} else {
				if !yield(cmd, pts) {
					return
				}
			}
			if len(pts) > 0 {
				current = pts[len(pts)-1]
				if cmd == CmdMoveTo {
					start = current
				}
			} else if cmd == CmdClose {
				current = start
			}
		}
	}
}
