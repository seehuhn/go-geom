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
	"fmt"
	"math"

	"seehuhn.de/go/geom/linalg"
	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

// CapStyle is the shape of the end of a stroked open path.
type CapStyle uint8

// The possible cap styles.
const (
	// CapButt ends the stroke square at the end point, so that it covers no
	// area beyond it.
	CapButt CapStyle = 0

	// CapRound ends the stroke with a half-disc of the line width, centred
	// on the end point.
	CapRound CapStyle = 1

	// CapSquare ends the stroke with a square of the line width, reaching
	// half the line width beyond the end point.
	CapSquare CapStyle = 2
)

func (c CapStyle) String() string {
	switch c {
	case CapRound:
		return "round"
	case CapSquare:
		return "square"
	case CapButt:
		return "butt"
	default:
		return fmt.Sprintf("CapStyle(%d)", c)
	}
}

// JoinStyle is the shape of the corner where two segments of a stroke meet.
type JoinStyle uint8

// The possible join styles.
const (
	// JoinMiter extends the outer edges of the two segments until they meet
	// in a point.  A corner too sharp for the miter limit is bevelled
	// instead.
	JoinMiter JoinStyle = 0

	// JoinRound fills the corner with a disc of the line width, centred on
	// the vertex.
	JoinRound JoinStyle = 1

	// JoinBevel closes the corner with a straight edge between the two outer
	// corners of the segments.
	JoinBevel JoinStyle = 2
)

func (j JoinStyle) String() string {
	switch j {
	case JoinRound:
		return "round"
	case JoinBevel:
		return "bevel"
	case JoinMiter:
		return "miter"
	default:
		return fmt.Sprintf("JoinStyle(%d)", j)
	}
}

// DefaultMiterLimit is the miter limit used where none is given.  It bevels
// a corner of less than about 11.5 degrees.
const DefaultMiterLimit = 10

// StrokeOptions are the parameters a stroke is drawn with.
type StrokeOptions struct {
	// Width is the thickness of the stroke.
	Width float64

	// Cap is the shape of the two ends of an open path.
	Cap CapStyle

	// Join is the shape of the corners where segments meet.
	Join JoinStyle

	// MiterLimit is the longest miter drawn as a miter rather than bevelled,
	// as a multiple of the line width; see [linalg.MiterLength] for how the
	// two are compared.  Zero stands for [DefaultMiterLimit], since a limit
	// below one is shorter than any miter and would bevel every corner.
	MiterLimit float64
}

// PolylineStrokeBBox returns an axis-aligned rectangle containing the stroke
// drawn along the given polylines.  Each sub-path is the list of its
// vertices, joined by straight segments.  closed says whether a segment runs
// from the last vertex of each sub-path back to its first, which gives the
// sub-path a join at every vertex and no ends to cap.
//
// Each vertex is bounded by the square around it which holds everything the
// stroke draws there, and the tip of a miter join is taken in as the single
// point it is.  The result is exact for a miter join and for a projecting
// square cap.  For a butt or round cap and for a round or bevel join it is
// the square around a disc of half the line width, which can reach up to
// half the line width further than the stroke itself does.
//
// The second result is false where no sub-path has a vertex, in which case
// there is nothing to bound.
func PolylineStrokeBBox(subpaths [][]vec.Vec2, closed bool, opt StrokeOptions) (rect.Rect, bool) {
	var bbox rect.Rect
	seeded := false

	if opt.MiterLimit == 0 {
		opt.MiterLimit = DefaultMiterLimit
	}

	half := opt.Width / 2
	extend := func(p vec.Vec2, r float64) {
		if !seeded {
			bbox = rect.Rect{LLx: p.X - r, LLy: p.Y - r, URx: p.X + r, URy: p.Y + r}
			seeded = true
			return
		}
		bbox.Add(p.X-r, p.Y-r)
		bbox.Add(p.X+r, p.Y+r)
	}

	for _, pts := range subpaths {
		pts = distinctVertices(pts, closed)
		n := len(pts)
		for i, p := range pts {
			if !closed && (i == 0 || i == n-1) {
				// an end of an open path carries a cap, not a join
				r := half
				if opt.Cap == CapSquare {
					r = squareCapReach(capDirection(pts, i), half)
				}
				extend(p, r)
				continue
			}

			extend(p, half)
			if opt.Join == JoinMiter {
				if tip, ok := miterTip(pts[(i+n-1)%n], p, pts[(i+1)%n], opt); ok {
					extend(tip, 0)
				}
			}
		}
	}
	return bbox, seeded
}

// miterTip returns the point the miter join at b reaches, for a stroke
// running from a to b to c.
//
// ok is false where the corner is bevelled instead, and where the path runs
// straight through b or doubles back on itself; in all of these the join
// reaches no further from b than half the line width, which the caller has
// allowed for already.
func miterTip(a, b, c vec.Vec2, opt StrokeOptions) (vec.Vec2, bool) {
	in, out := b.Sub(a), c.Sub(b)
	length, ok := linalg.MiterLength(in, out, opt.Width, opt.MiterLimit)
	if !ok {
		return vec.Vec2{}, false
	}
	// The tip lies on the bisector of the outer angle, which is the
	// difference of the two unit directions: it points away from the turn,
	// and is the zero vector where there is no turn.
	u := in.Normalize().Sub(out.Normalize())
	if u.Length() == 0 {
		return vec.Vec2{}, false
	}
	return b.Add(u.Normalize().Mul(length)), true
}

// distinctVertices returns the vertices of a sub-path with each run of
// coincident consecutive vertices collapsed to one.  For a closed sub-path
// a last vertex which coincides with the first is dropped as well, since
// the closing segment already returns there.
//
// A zero-length segment adds nothing to the stroke, and a stroker joins the
// segments on either side of it as though it were not there.  The join at
// such a vertex therefore has to be computed from the nearest distinct
// neighbours, which this makes the adjacent ones.  The result shares the
// input's storage where there is nothing to collapse.
func distinctVertices(pts []vec.Vec2, closed bool) []vec.Vec2 {
	n := len(pts)
	if closed && n > 1 && pts[n-1] == pts[0] {
		n--
	}
	// k is the index of the first repeated vertex
	k := 1
	for k < n && pts[k] != pts[k-1] {
		k++
	}
	if k >= n {
		return pts[:n]
	}

	out := make([]vec.Vec2, k, n)
	copy(out, pts)
	for _, p := range pts[k:n] {
		if p != out[len(out)-1] {
			out = append(out, p)
		}
	}
	if closed && len(out) > 1 && out[len(out)-1] == out[0] {
		out = out[:len(out)-1]
	}
	return out
}

// capDirection returns the direction the cap at the end vertex i of an open
// sub-path points in: away from the vertex, along the segment which ends
// there.  The vector is not normalised, and it is zero for a sub-path with
// a single vertex, which has no direction to take.  pts must have no two
// consecutive vertices coincide.
func capDirection(pts []vec.Vec2, i int) vec.Vec2 {
	if len(pts) < 2 {
		return vec.Vec2{}
	}
	if i == 0 {
		return pts[0].Sub(pts[1])
	}
	return pts[i].Sub(pts[i-1])
}

// squareCapReach returns how far a projecting square cap pointing in
// direction away reaches from the end point, measured along either axis.
// half is half the line width.
//
// The cap is a square of the line width with the end point in the middle of
// one side, so its two far corners lie half the line width along the
// direction and half of it to either side.  Measured along an axis that is
// half the line width times the sum of the two components of the unit
// direction, which comes to half the width for a cap along an axis and to
// half the width times the square root of two for one at 45 degrees.
func squareCapReach(away vec.Vec2, half float64) float64 {
	length := away.Length()
	if length == 0 {
		// no direction to point the cap in: allow for the worst case
		return half * math.Sqrt2
	}
	return half * (math.Abs(away.X) + math.Abs(away.Y)) / length
}
