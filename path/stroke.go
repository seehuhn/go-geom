// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

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
	// An infinite limit gives a miter at every corner except where the path
	// doubles back.
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
// The rectangle grows with the stroke options as described for
// [Path.StrokeBBox].
//
// The second result is false where the stroke draws nothing, in which case
// there is nothing to bound.
func PolylineStrokeBBox(subpaths [][]vec.Vec2, closed bool, opt StrokeOptions) (rect.Rect, bool) {
	b := newStrokeBounds(opt)
	var segs []strokeSeg
	for _, pts := range subpaths {
		if len(pts) == 0 {
			continue
		}
		segs = segs[:0]
		for i := 1; i < len(pts); i++ {
			segs = append(segs, lineSeg(pts[i-1], pts[i]))
		}
		if closed {
			segs = append(segs, lineSeg(pts[len(pts)-1], pts[0]))
		}
		b.subpath(pts[0], segs, closed)
	}
	return b.bbox, b.seeded
}

// StrokeBBox returns an axis-aligned rectangle which contains the stroke
// drawn along the path.  A sub-path ended by a [CmdClose] is closed: it has
// a join at every vertex and no ends to cap.  The other sub-paths are open.
//
// Where the path has no curve segments, the rectangle is the smallest one
// containing the stroke, as for [PolylineStrokeBBox].  Where it has curves,
// the rectangle contains both the ideal stroke and the stroke a renderer
// draws by replacing each curve with a polyline whose vertices lie on the
// curve, joining the pieces with the path's join style and pointing caps
// along the end pieces.  With round caps and joins it is the smallest
// rectangle containing the ideal stroke; otherwise it can be larger.  Both
// hold up to rounding: where a curve's stroke touches the rectangle, it can
// reach beyond it by a few units in the last place.
//
// A sub-path all of whose points coincide, including one which consists of
// a [CmdMoveTo] and a [CmdClose], is a zero-length line.  With round caps it
// is drawn as a disc of the line width.  With butt or square caps it draws
// nothing.  A CmdMoveTo which starts no segment draws nothing either.
//
// With opt.Dashed set, the rectangle contains the stroke for every dash
// pattern and phase.
//
// Changing one option in any of the following ways, with the others kept,
// gives a rectangle which covers the original one: setting opt.Dashed,
// raising the miter limit, replacing bevel joins with miter joins, and
// replacing butt caps with square or round caps.  For the miter limit, zero
// counts as [DefaultMiterLimit].  Where the dash pattern, join style, miter
// limit or cap style is not known, the stroke for every value they can take
// thus lies in the union of four rectangles: miter joins with an infinite
// miter limit or round joins, combined with square or round caps, all with
// opt.Dashed set.
//
// With miter joins and an infinite miter limit, a path with curves can give
// a rectangle of infinite extent.  A renderer which flattens a curve joins
// chords between points of the curve, and where two consecutive chords
// point in nearly opposite directions, there is no bound on the length of
// the miter at the corner between them.  The rectangle is infinite only
// where the control points allow this.  Within a curve, this is where the
// legs of its control polygon span half a turn or more, as near a cusp.
// Where two segments meet at a point q, it is where the vectors to q from
// the other control points of the first segment span half a turn or more,
// where the vectors from q to the other control points of the second
// segment do, or where a combination with non-negative weights of the
// former points opposite to such a combination of the latter.
//
// The second result is false where the stroke draws nothing, in which case
// there is nothing to bound.
func (p Path) StrokeBBox(opt StrokeOptions) (rect.Rect, bool) {
	b := newStrokeBounds(opt)

	var segs []strokeSeg
	var start, cur vec.Vec2
	started := false // a sub-path has begun
	finish := func(closed bool) {
		if len(segs) > 0 {
			b.subpath(start, segs, closed)
		}
		segs = segs[:0]
		started = false
	}
	for cmd, pts := range p.ToCubic() {
		switch cmd {
		case CmdMoveTo:
			finish(false)
			start, cur = pts[0], pts[0]
			started = true
		case CmdLineTo, CmdCubeTo:
			if !started {
				start = cur
				started = true
			}
			if cmd == CmdLineTo {
				segs = append(segs, lineSeg(cur, pts[0]))
			} else {
				segs = append(segs, curveSeg(Cubic{cur, pts[0], pts[1], pts[2]}))
			}
			cur = pts[len(pts)-1]
		case CmdClose:
			if started {
				segs = append(segs, lineSeg(cur, start))
				finish(true)
				cur = start
			}
		}
	}
	finish(false)
	return b.bbox, b.seeded
}

// strokeSeg is one segment of a stroked sub-path: a cubic Bezier curve, or
// a straight line from c[0] to c[3], with c[1] = c[0] and c[2] = c[3].
//
// The directions of the chords between two points of the segment lie in a
// range of width spread.  The range start holds the directions of the
// chords which start at c[0], and end those of the chords which end at c[3].
type strokeSeg struct {
	c          Cubic
	curve      bool
	spread     float64
	start, end dirRange
}

func lineSeg(p, q vec.Vec2) strokeSeg {
	r := directionRange([3]vec.Vec2{q.Sub(p)})
	return strokeSeg{c: Cubic{p, p, q, q}, start: r, end: r}
}

// curveSeg returns the segment for the curve c.
//
// The derivative of the curve is a combination with non-negative weights of
// the legs of the control polygon, and so is every chord, the integral of
// the derivative between its end points.  A point of the curve is a
// combination of the control points with non-negative weights which sum to
// one.  A chord starting at c[0] is thus a combination with non-negative
// weights of the vectors from c[0] to the other control points, and a chord
// ending at c[3] is one of the vectors from the other control points to
// c[3].
func curveSeg(c Cubic) strokeSeg {
	_, lo, hi := relativeAngles([3]vec.Vec2{c[1].Sub(c[0]), c[2].Sub(c[1]), c[3].Sub(c[2])})
	return strokeSeg{
		c:      c,
		curve:  true,
		spread: hi - lo,
		start:  directionRange([3]vec.Vec2{c[1].Sub(c[0]), c[2].Sub(c[0]), c[3].Sub(c[0])}),
		end:    directionRange([3]vec.Vec2{c[3].Sub(c[0]), c[3].Sub(c[1]), c[3].Sub(c[2])}),
	}
}

// startDir returns the direction in which the segment leaves its start
// point.  It is zero for a segment of length zero.
func (s strokeSeg) startDir() vec.Vec2 {
	for _, p := range s.c[1:3] {
		if d := p.Sub(s.c[0]); d != (vec.Vec2{}) {
			return d
		}
	}
	return s.c[3].Sub(s.c[0])
}

// endDir returns the direction in which the segment arrives at its end
// point.  It is zero for a segment of length zero.
func (s strokeSeg) endDir() vec.Vec2 {
	for i := 2; i > 0; i-- {
		if d := s.c[3].Sub(s.c[i]); d != (vec.Vec2{}) {
			return d
		}
	}
	return s.c[3].Sub(s.c[0])
}

// dirRange is the range of directions at angles from base+lo to base+hi,
// with lo <= hi, measured counterclockwise from the positive x axis.  A
// range of half a turn or more gives no bound on the direction.
//
// The offsets lo and hi are kept apart from base, so that the width of the
// range is free from the rounding of the additions.
type dirRange struct {
	base, lo, hi float64
}

// width returns the angle the range spans.
func (r dirRange) width() float64 {
	return r.hi - r.lo
}

// directionRange returns the smallest range which holds the directions of
// the non-zero vectors in vs, or a range of half a turn or more where no
// range shorter than half a turn holds them.  A range shorter than half a
// turn holds the direction of every non-zero combination of the vectors
// with non-negative weights.
func directionRange(vs [3]vec.Vec2) dirRange {
	ref, lo, hi := relativeAngles(vs)
	return dirRange{base: math.Atan2(ref.Y, ref.X), lo: lo, hi: hi}
}

// relativeAngles returns the first non-zero vector ref in vs, together with
// the smallest and largest signed angle from ref to a non-zero vector in
// vs.  These are the offsets of the range found by [directionRange],
// relative to the direction of ref.
func relativeAngles(vs [3]vec.Vec2) (ref vec.Vec2, lo, hi float64) {
	for _, d := range vs {
		if d == (vec.Vec2{}) {
			continue
		}
		if ref == (vec.Vec2{}) {
			ref = d
			continue
		}
		// Relative to a vector in the range, the other vectors lie within
		// half a turn either side, and the signed angles are their
		// positions in the range.  Where the vectors fit into no range
		// shorter than half a turn, the signed angles span half a turn or
		// more as well.
		a := math.Atan2(ref.Cross(d), ref.Dot(d))
		lo, hi = min(lo, a), max(hi, a)
	}
	return ref, lo, hi
}

// strokeBounds accumulates the bounding box of the pieces a stroke is made
// of.
type strokeBounds struct {
	opt      StrokeOptions
	half     float64 // half the line width
	maxMiter float64 // the furthest a join can reach from its vertex
	bbox     rect.Rect
	seeded   bool
}

func newStrokeBounds(opt StrokeOptions) *strokeBounds {
	if opt.MiterLimit == 0 {
		opt.MiterLimit = DefaultMiterLimit
	}
	return &strokeBounds{
		opt:      opt,
		half:     opt.Width / 2,
		maxMiter: linalg.MaxMiterLength(opt.Width, opt.MiterLimit),
	}
}

// subpath adds the stroke of the sub-path which starts at start and
// consists of segs, each starting where the one before it ends.  For a
// closed sub-path, the last segment ends at start.  The function overwrites
// segs.
//
// The stroke is the union of one piece per segment, a join at each corner
// and a cap at each end.
func (b *strokeBounds) subpath(start vec.Vec2, segs []strokeSeg, closed bool) {
	// A zero-length segment adds nothing to the stroke, and a stroker joins
	// the segments on either side of it as though it were not there.
	k := 0
	for _, s := range segs {
		if s.startDir() != (vec.Vec2{}) {
			segs[k] = s
			k++
		}
	}
	segs = segs[:k]
	n := len(segs)
	if n == 0 {
		if b.opt.Cap == CapRound {
			b.disc(start)
		}
		return
	}

	for _, s := range segs {
		b.segment(s)
	}
	for i := range n - 1 {
		b.joinSegs(segs[i], segs[i+1])
	}
	if closed {
		b.joinSegs(segs[n-1], segs[0])
	} else {
		b.capEnd(segs[0].c[0], segs[0].startDir().Neg(), segs[0])
		b.capEnd(segs[n-1].c[3], segs[n-1].endDir(), segs[n-1])
	}
}

// segment adds the stroke along one segment of non-zero length, without the
// joins and caps at its ends.
//
// For a straight segment the piece is a rectangle.  Where two segments
// meet, the rectangles' corners already hold a bevel join, and at an open
// end they hold a butt cap.  A dash has its own caps, which along a
// straight segment stay within the segment's rectangle and reach furthest
// out where the dash ends or starts on one of its end points.
func (b *strokeBounds) segment(s strokeSeg) {
	if !s.curve {
		p, q := s.c[0], s.c[3]
		side := q.Sub(p).Rot90().Normalize().Mul(b.half)
		b.add(p.Add(side))
		b.add(p.Sub(side))
		b.add(q.Add(side))
		b.add(q.Sub(side))
		if b.opt.Dashed {
			b.cap(p, p.Sub(q))
			b.cap(q, q.Sub(p))
		}
		return
	}

	// Every point of the ideal stroke lies within half the line width of
	// the curve.  A renderer which flattens the curve joins the pieces with
	// the path's join style, and since the two chords at a join both point
	// within the range of directions of the curve's chords, the miter there
	// reaches no further than [strokeBounds.miterReach] for the width of
	// that range.  A square dash cap, which can sit anywhere along the
	// curve, reaches out by the square root of two times half the line
	// width.
	reach := b.half
	if b.opt.Join == JoinMiter {
		reach = b.miterReach(s.spread)
	}
	if b.opt.Dashed && b.opt.Cap == CapSquare {
		reach = max(reach, b.half*math.Sqrt2)
	}
	r := s.c.BBox().Grow(reach)
	b.add(vec.Vec2{X: r.LLx, Y: r.LLy})
	b.add(vec.Vec2{X: r.URx, Y: r.URy})
}

// joinSegs adds the join where segment s ends and segment next starts.
func (b *strokeBounds) joinSegs(s, next strokeSeg) {
	q := s.c[3]
	if !s.curve && !next.curve {
		b.join(s.c[0], q, next.c[3])
		return
	}

	// A renderer which flattens a curve joins the chords next to the
	// vertex, and each chord points within the range of directions of the
	// chords of its segment which end or start at the vertex.  A round or
	// bevel join stays within half the line width of the vertex, where the
	// curve's own box already reaches.
	if b.opt.Join == JoinMiter {
		b.discRadius(q, b.miterReach(maxTurn(s, next)))
	}
}

// maxTurn returns the largest angle by which the direction can turn from a
// chord of s which ends at the end point of s to a chord of next which
// starts there, or half a turn where the ranges of directions give no bound
// below this.
func maxTurn(s, next strokeSeg) float64 {
	in, out := s.end, next.start
	if in.width() >= math.Pi || out.width() >= math.Pi {
		return math.Pi
	}

	// The differences between a direction of out and one of in fill the
	// interval [lo, hi], which is shorter than a full turn.  The angle
	// between two directions grows with the distance of their difference
	// from the nearest multiple of a full turn, and reaches half a turn at
	// the odd multiples of half a turn.
	d := out.base - in.base
	lo, hi := d+(out.lo-in.hi), d+(out.hi-in.lo)
	k := math.Ceil((lo - math.Pi) / (2 * math.Pi))
	if math.Pi+2*math.Pi*k <= hi {
		return math.Pi
	}
	return max(math.Abs(math.Remainder(lo, 2*math.Pi)),
		math.Abs(math.Remainder(hi, 2*math.Pi)))
}

// capEnd adds the cap at the end point q of an open sub-path, where segment
// s arrives at q from direction away.
func (b *strokeBounds) capEnd(q, away vec.Vec2, s strokeSeg) {
	if s.curve && b.opt.Cap == CapSquare {
		// a renderer which flattens the curve points the cap along the
		// last chord, not the tangent
		b.discRadius(q, b.half*math.Sqrt2)
		return
	}
	b.cap(q, away)
}

// miterReach returns how far a join can reach from its vertex where the
// path turns by at most theta there.
func (b *strokeBounds) miterReach(theta float64) float64 {
	if theta >= math.Pi {
		return b.maxMiter
	}
	return min(b.maxMiter, b.half/math.Cos(theta/2))
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
	b.discRadius(centre, b.half)
}

// discRadius adds the disc of radius r around centre.
func (b *strokeBounds) discRadius(centre vec.Vec2, r float64) {
	for _, e := range axes {
		b.add(centre.Add(e.Mul(r)))
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
func (b *strokeBounds) join(p, q, r vec.Vec2) {
	switch b.opt.Join {
	case JoinMiter:
		if tip, ok := miterTip(p, q, r, b.opt); ok {
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
func (b *strokeBounds) cap(q, away vec.Vec2) {
	d := away.Normalize()
	switch b.opt.Cap {
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
