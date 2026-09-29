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

	// Dashed says that the stroke is drawn with a dash pattern.  Every dash
	// is capped, and a dash can end or start on any vertex, so the stroke
	// can then reach beyond the undashed one near the vertices.
	Dashed bool
}

// PolylineStrokeBBox returns the smallest axis-aligned rectangle containing
// the stroke drawn along the given polylines.  Each sub-path is the list of
// its vertices, joined by straight segments.  closed says whether a segment
// runs from the last vertex of each sub-path back to its first, which gives
// the sub-path a join at every vertex and no ends to cap.
//
// A sub-path whose vertices all coincide, including one with a single vertex,
// is a zero-length line.  With round caps it is drawn as a disc of the line
// width.  With butt or square caps it draws nothing, since the caps have no
// direction to point in.
//
// With opt.Dashed set, the rectangle is the smallest one containing the
// stroke for every dash pattern and phase.
//
// The second result is false where the stroke draws nothing, in which case
// there is nothing to bound.
func PolylineStrokeBBox(subpaths [][]vec.Vec2, closed bool, opt StrokeOptions) (rect.Rect, bool) {
	if opt.MiterLimit == 0 {
		opt.MiterLimit = DefaultMiterLimit
	}

	b := &strokeBounds{half: opt.Width / 2}
	for _, pts := range subpaths {
		pts = distinctVertices(pts, closed)
		n := len(pts)
		if n == 0 {
			continue
		}
		if n == 1 {
			if opt.Cap == CapRound {
				b.disc(pts[0])
			}
			continue
		}

		// The stroke is the union of one rectangle per segment, a join at
		// each corner and a cap at each end.  Where two segments meet, the
		// rectangles' corners already hold a bevel join, and at an open end
		// they hold a butt cap.  A dash has its own caps, which along a
		// segment stay within the segment's rectangle and reach furthest out
		// where the dash ends or starts on one of its end points.
		numSeg := n - 1
		if closed {
			numSeg = n
		}
		for i := range numSeg {
			p, q := pts[i], pts[(i+1)%n]
			side := q.Sub(p).Rot90().Normalize().Mul(b.half)
			b.add(p.Add(side))
			b.add(p.Sub(side))
			b.add(q.Add(side))
			b.add(q.Sub(side))
			if opt.Dashed {
				b.cap(p, p.Sub(q), opt.Cap)
				b.cap(q, q.Sub(p), opt.Cap)
			}
		}

		first, last := 1, n-1
		if closed {
			first, last = 0, n
		}
		for i := first; i < last; i++ {
			b.join(pts[(i+n-1)%n], pts[i], pts[(i+1)%n], opt)
		}
		if !closed {
			b.cap(pts[0], pts[0].Sub(pts[1]), opt.Cap)
			b.cap(pts[n-1], pts[n-1].Sub(pts[n-2]), opt.Cap)
		}
	}
	return b.bbox, b.seeded
}

// strokeBounds accumulates the bounding box of the pieces a stroke is made
// of.
type strokeBounds struct {
	half   float64 // half the line width
	bbox   rect.Rect
	seeded bool
}

func (b *strokeBounds) add(p vec.Vec2) {
	if !b.seeded {
		b.bbox = rect.Rect{LLx: p.X, LLy: p.Y, URx: p.X, URy: p.Y}
		b.seeded = true
		return
	}
	b.bbox.Add(p.X, p.Y)
}

// axes are the four directions in which a circular arc can reach an edge of
// the bounding box somewhere between its two end points.
var axes = [4]vec.Vec2{{X: 1}, {X: -1}, {Y: 1}, {Y: -1}}

// disc adds the disc of half the line width around centre.
func (b *strokeBounds) disc(centre vec.Vec2) {
	for _, e := range axes {
		b.add(centre.Add(e.Mul(b.half)))
	}
}

// arc adds the arc of radius half the line width around centre which is
// centred on the unit direction mid and reaches a half-angle to either side
// of it, given by its cosine.  Only the points where the arc reaches furthest
// along an axis are added: the arc's two end points are left to the caller.
func (b *strokeBounds) arc(centre, mid vec.Vec2, cosHalfAngle float64) {
	for _, e := range axes {
		if e.Dot(mid) > cosHalfAngle {
			b.add(centre.Add(e.Mul(b.half)))
		}
	}
}

// join adds the join at q of a stroke running from p to q to r, beyond the
// corners of the two segments' rectangles.
func (b *strokeBounds) join(p, q, r vec.Vec2, opt StrokeOptions) {
	switch opt.Join {
	case JoinMiter:
		if tip, ok := miterTip(p, q, r, opt); ok {
			b.add(tip)
		}
	case JoinRound:
		// The join is the pie slice on the outer side of the corner,
		// between the two segments' outer edges.  It is centred on the
		// outer bisector, and spans the angle theta the direction turns by;
		// half the sum of the unit directions has length cos(theta/2).
		in, out := q.Sub(p).Normalize(), r.Sub(q).Normalize()
		outer := in.Sub(out)
		if outer == (vec.Vec2{}) {
			return // straight on: no corner to fill
		}
		b.arc(q, outer.Normalize(), in.Add(out).Length()/2)
	}
}

// cap adds the cap at the end point q of an open sub-path, pointing in
// direction away, beyond the corners of the end segment's rectangle.
func (b *strokeBounds) cap(q, away vec.Vec2, style CapStyle) {
	d := away.Normalize()
	switch style {
	case CapSquare:
		ext, side := d.Mul(b.half), d.Rot90().Mul(b.half)
		b.add(q.Add(ext).Add(side))
		b.add(q.Add(ext).Sub(side))
	case CapRound:
		b.arc(q, d, 0)
	}
}

// miterTip returns the point the miter join at b reaches, for a stroke
// running from a to b to c.
//
// ok is false where the corner is bevelled instead, and where the path runs
// straight through b or doubles back on itself; in all of these the join
// reaches no further than the corners of the two segments' rectangles.
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
