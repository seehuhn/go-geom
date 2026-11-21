// Copyright 2025 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import (
	"iter"
	"math"

	"seehuhn.de/go/geom/matrix"
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
//
// CmdClose leaves the current point at the start of the sub-path it closed.
// A drawing command which follows a CmdClose without an intervening CmdMoveTo
// therefore starts a new sub-path at that point.  This matches the closepath
// operator of the PostScript language and the "h" operator of PDF content
// streams.  A CmdClose with no open sub-path, at the start of the path or
// directly after another CmdClose, does nothing.
//
// Before the first CmdMoveTo, the current point is the origin, and so is the
// start of the current sub-path.  A path which begins with a drawing command
// therefore starts a sub-path at the origin.
type Path iter.Seq2[Command, []vec.Vec2]

// Empty is a path which contains no segments.
var Empty Path = func(yield func(Command, []vec.Vec2) bool) {}

// Transform applies matrix M to the path.  A path which begins with a drawing
// command starts at the image of the origin under M, which the transformed
// path states by an explicit CmdMoveTo.
func (p Path) Transform(M matrix.Matrix) Path {
	return func(yield func(Command, []vec.Vec2) bool) {
		var buf [3]vec.Vec2
		moved := false
		for cmd, pts := range p {
			if !moved && cmd != CmdClose {
				if cmd != CmdMoveTo {
					buf[0] = vec.Vec2{X: M[4], Y: M[5]}
					if !yield(CmdMoveTo, buf[:1]) {
						return
					}
				}
				moved = true
			}
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

// TightBBox returns the smallest axis-aligned rectangle which contains the
// path.  Unlike [Path.BBox], it bounds curves by the points they pass
// through, not by their control points.  The rectangle contains every
// segment of the path, including the zero-length segment of a sub-path which
// consists of a CmdMoveTo and a CmdClose.  A CmdMoveTo which starts no
// segment, because another CmdMoveTo or the end of the path follows it,
// draws nothing and is ignored.  The result is exact up to rounding, as
// described for [Cubic.BBox].
//
// The second result is false for a path without segments, in which case
// there is nothing to bound and the rectangle is zero.
func (p Path) TightBBox() (rect.Rect, bool) {
	var bbox rect.Rect
	first := true
	var cur, start vec.Vec2
	var pending bool // start is from a CmdMoveTo, not yet in bbox
	for cmd, pts := range p {
		if cmd == CmdMoveTo {
			cur, start = pts[0], pts[0]
			pending = true
			continue
		}
		// a sub-path before the first CmdMoveTo starts at the origin
		if pending || first && cmd != CmdClose {
			extendBox(&bbox, first, start)
			first, pending = false, false
		}
		switch cmd {
		case CmdClose:
			cur = start
			continue
		case CmdLineTo:
			bbox.Add(pts[0].X, pts[0].Y)
		case CmdQuadTo:
			extendQuad(&bbox, cur, pts[0], pts[1])
		case CmdCubeTo:
			Cubic{cur, pts[0], pts[1], pts[2]}.extend(&bbox)
		}
		cur = pts[len(pts)-1]
	}
	return bbox, !first
}

// extendQuad enlarges bbox to contain the quadratic Bezier curve with start
// point p0, control point c and end point p2, apart from its start point.
func extendQuad(bbox *rect.Rect, p0, c, p2 vec.Vec2) {
	bbox.Add(p2.X, p2.Y)
	for _, t := range [2]float64{
		quadExtremum(p0.X, c.X, p2.X),
		quadExtremum(p0.Y, c.Y, p2.Y),
	} {
		if t > 0 {
			q := lerp(lerp(p0, c, t), lerp(c, p2, t), t)
			bbox.Add(q.X, q.Y)
		}
	}
}

// quadExtremum returns the parameter in (0, 1] where the one-dimensional
// quadratic Bezier curve with coefficients a0, a1, a2 has its extremum, or 0
// if there is none.  The result is 1 only where rounding moves an extremum
// close to the end point onto it.
func quadExtremum(a0, a1, a2 float64) float64 {
	// Scaling by a power of two leaves the extremum unchanged and introduces
	// no rounding short of underflow.  It keeps the differences below from
	// overflowing.
	_, e := math.Frexp(max(math.Abs(a0), math.Abs(a1), math.Abs(a2)))
	a0, a1, a2 = math.Ldexp(a0, -e), math.Ldexp(a1, -e), math.Ldexp(a2, -e)

	d0, d1 := a1-a0, a2-a1
	if d0 > 0 && d1 < 0 || d0 < 0 && d1 > 0 {
		// the signs differ, so the denominator has no cancellation
		return d0 / (d0 - d1)
	}
	return 0
}

// BBox returns the smallest axis-aligned rectangle which contains all points
// of the path, counting the control points of curves.  The rectangle contains
// the path, but can be larger than [Path.TightBBox].  As for TightBBox, a
// CmdMoveTo which starts no segment is ignored.
//
// The second result is false for a path without segments, in which case
// there is nothing to bound and the rectangle is zero.
func (p Path) BBox() (rect.Rect, bool) {
	var bbox rect.Rect
	first := true
	var start vec.Vec2
	var pending bool // start is from a CmdMoveTo, not yet in bbox
	for cmd, pts := range p {
		if cmd == CmdMoveTo {
			start = pts[0]
			pending = true
			continue
		}
		// a sub-path before the first CmdMoveTo starts at the origin
		if pending || first && cmd != CmdClose {
			extendBox(&bbox, first, start)
			first, pending = false, false
		}
		for _, pt := range pts {
			bbox.Add(pt.X, pt.Y)
		}
	}
	return bbox, !first
}

// extendBox enlarges bbox to contain q.  If first is true, bbox is replaced
// by the rectangle consisting of q alone.
func extendBox(bbox *rect.Rect, first bool, q vec.Vec2) {
	if first {
		*bbox = rect.Rect{LLx: q.X, LLy: q.Y, URx: q.X, URy: q.Y}
		return
	}
	bbox.Add(q.X, q.Y)
}
