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

import (
	"iter"

	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

type Command byte

const (
	CmdMoveTo Command = iota + 1
	CmdLineTo
	CmdQuadTo
	CmdCubeTo
	CmdClose
)

// NumPoints returns the number of points for this command.
func (c Command) NumPoints() int {
	switch c {
	case CmdMoveTo, CmdLineTo:
		return 1
	case CmdQuadTo:
		return 2
	case CmdCubeTo:
		return 3
	default:
		return 0
	}
}

// Path iterates over individual segments.
// The first argument is the segment type, which determines the number of points:
//   - CmdMoveTo: 1 point (starts a new sub-path at the given point)
//   - CmdLineTo: 1 point (endpoint of the line segment)
//   - CmdQuadTo: 2 points (control point and endpoint of the quadratic Bezier curve)
//   - CmdCubeTo: 3 points (control point 1, control point 2, and endpoint of the cubic Bezier curve)
//   - CmdClose: 0 points (straight line from the current point to start of the current sub-path)
type Path iter.Seq2[Command, []vec.Vec2]

// Transform applies matrix M to the path.
func (p Path) Transform(M [6]float64) Path {
	return func(yield func(Command, []vec.Vec2) bool) {
		var buf [3]vec.Vec2
		for cmd, pts := range p {
			for i, p := range pts {
				buf[i] = vec.Vec2{
					X: p.X*M[0] + p.Y*M[2] + M[4],
					Y: p.X*M[1] + p.Y*M[3] + M[5],
				}
			}
			if !yield(cmd, buf[:len(pts)]) {
				return
			}
		}
	}
}

// ToCubic converts all quadratic segments to cubic segments.
// The resulting representation describes the same path but uses more control points.
func (p Path) ToCubic() Path {
	return func(yield func(Command, []vec.Vec2) bool) {
		var current vec.Vec2
		var start vec.Vec2
		var buf [3]vec.Vec2
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

// BBox computes the bounding box for a path.
// For curves, it includes all control points, providing a conservative approximation.
func (p Path) BBox() rect.Rect {
	var bbox rect.Rect
	first := true

	for _, pts := range p {
		for _, pt := range pts {
			if first {
				bbox.LLx, bbox.LLy = pt.X, pt.Y
				bbox.URx, bbox.URy = pt.X, pt.Y
				first = false
			} else {
				bbox.Add(pt.X, pt.Y)
			}
		}
	}

	return bbox
}

// Data is a compact, mutable representation of a path.
type Data struct {
	Cmds   []Command
	Coords []vec.Vec2
}

// MoveTo starts a new sub-path at the given point.
func (d *Data) MoveTo(p vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdMoveTo)
	d.Coords = append(d.Coords, p)
	return d
}

// LineTo adds a line segment to the given point.
func (d *Data) LineTo(p vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdLineTo)
	d.Coords = append(d.Coords, p)
	return d
}

// QuadTo adds a quadratic Bezier curve.
func (d *Data) QuadTo(ctrl, end vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdQuadTo)
	d.Coords = append(d.Coords, ctrl, end)
	return d
}

// CubeTo adds a cubic Bezier curve.
func (d *Data) CubeTo(ctrl1, ctrl2, end vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdCubeTo)
	d.Coords = append(d.Coords, ctrl1, ctrl2, end)
	return d
}

// Close closes the current sub-path.
func (d *Data) Close() *Data {
	d.Cmds = append(d.Cmds, CmdClose)
	return d
}

// Iter returns a Path iterator over the path data.
func (d *Data) Iter() Path {
	return func(yield func(Command, []vec.Vec2) bool) {
		i := 0
		for _, cmd := range d.Cmds {
			n := cmd.NumPoints()
			if !yield(cmd, d.Coords[i:i+n]) {
				return
			}
			i += n
		}
	}
}
