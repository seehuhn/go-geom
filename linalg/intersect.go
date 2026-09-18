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

// parallelTol is the smallest sine of the angle between two lines for which
// their intersection point is still considered meaningful.  The point moves
// by about the inverse of this for a perturbation of the inputs.
const parallelTol = 1e-9

// endTol is how far from either end of a segment a crossing may fall and
// still count as at that end, as a fraction of the segment's length.  The
// parameter of a crossing at an endpoint is formed from two cross products
// and a quotient, and their rounding leaves it a few ulp either side of the
// exact 0 or 1, so an exact range test would reject some such crossings and
// accept others, and an exact comparison would put some at the end and others
// a rounding error short of it.
const endTol = 1e-12

// degenerateTol is the smallest length, as a fraction of the distance of the
// points from the origin, for which a line segment still has a usable
// direction.  Below this the direction is dominated by the rounding error in
// the endpoints.
const degenerateTol = 1e-12

// Intersection calculates the intersection point of two infinite lines, the
// first through a1 and a2, the second through b1 and b2.  ok is false if the
// two lines are parallel, or so nearly parallel that the intersection point
// carries no useful accuracy, or if either pair of points coincides.  The zero
// vector is returned in that case, since no point on either line is a better
// answer than another; a caller which has a fallback in mind should supply it
// itself.
func Intersection(a1, a2, b1, b2 vec.Vec2) (vec.Vec2, bool) {
	// Each direction is split into a normalised vector and a power of two.
	// Every product below is then formed from components of magnitude at most
	// one, so that none of them can overflow on its way to a result which is
	// itself in range, and the powers of two are accounted for separately.
	u, _ := a2.Sub(a1).Frexp()
	v, _ := b2.Sub(b1).Frexp()
	w, ew := b1.Sub(a1).Frexp()

	// The test is on the sine of the angle between the two directions, so
	// that it does not depend on the scale the points are expressed in.
	// Testing the cross product on its own would call perpendicular
	// millimetre-sized segments parallel, and near-parallel kilometre-sized
	// ones distinct.  Coincident points give zero on both sides and are
	// rejected too.  The powers of two divided out of the two directions are
	// a common factor of both sides here, so they cancel and are not needed.
	denom := u.Cross(v)
	if math.Abs(denom) <= parallelTol*u.Length()*v.Length() {
		return vec.Vec2{}, false
	}

	// The intersection is a1 + t*(a2-a1) with t = (w x v)/(u x v).  The power
	// of two of the second line cancels from the quotient, and that of the
	// first cancels against the step taken along it, so only the one divided
	// out of w = b1-a1 is left to put back.
	t := w.Cross(v) / denom
	return a1.Add(u.Mul(math.Ldexp(t, ew))), true
}

// Miter calculates the miter point for two line segments (a,b) and (b,c).
// This is the intersection point of the outer edges when stroked with the given
// line width.
//
// ok is false if either segment is too short to have a direction which can be
// told apart from the rounding error of its endpoints, in which case b is
// returned, or if the two outer edges run parallel and never meet, in which
// case the midpoint of the two offset corner vertices is returned.  That
// midpoint is b itself for an exactly straight corner, and a point on the
// offset edge for one which only just counts as straight.
//
// How short is too short is relative to the distance of the corner from
// the origin, since that is what fixes the precision the coordinates are held
// to: the same corner shape needs longer segments far from the origin than
// close to it.
func Miter(a, b, c vec.Vec2, lineWidth float64, outer bool) (vec.Vec2, bool) {
	ba := vec.Vec2{X: a.X - b.X, Y: a.Y - b.Y}
	bc := vec.Vec2{X: c.X - b.X, Y: c.Y - b.Y}
	baLength := ba.Length()
	bcLength := bc.Length()

	// A coincident pair has no direction at all.  This is also the only case
	// where all three points sit at the origin, leaving no scale to measure
	// the segments against below.
	if baLength == 0 || bcLength == 0 {
		return b, false
	}

	// A segment this much shorter than the coordinates has no direction left
	// to speak of, since its endpoints differ by little more than their own
	// rounding error.  The lengths are divided by the scale rather than the
	// bound multiplied by it: that product underflows to zero for coordinates
	// in the subnormal range, which would let any segment through.
	scale := max(a.Length(), b.Length(), c.Length())
	if baLength/scale <= degenerateTol || bcLength/scale <= degenerateTol {
		return b, false
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

	p1 := b.Add(baNorm)
	p2 := b.Add(bcNorm)
	if p, ok := Intersection(a.Add(baNorm), p1, p2, c.Add(bcNorm)); ok {
		return p, true
	}

	// The corner runs straight through, so the two outer edges are parallel
	// and never meet.  The midpoint of the two offset corner vertices stands
	// in for the miter point: it lies on the offset edge where both segments
	// put it on the same side, and collapses onto b where they put it on
	// opposite sides, as they do for an exactly straight corner.
	return vec.Middle(p1, p2), false
}

// SegmentIntersection calculates where the line segment from a1 to a2 crosses
// the line segment from b1 to b2.  The crossing point is returned together
// with the two parameters which locate it on the segments, so that the point
// is both a1 + t*(a2-a1) and b1 + s*(b2-b1), with t and s in the range [0, 1].
// The point returned is the one formed from the first segment; the one formed
// from the second differs from it by a rounding error.
//
// Segments which meet at an endpoint count as crossing.  The parameter of
// such a crossing does not always work out as exactly 0 or 1, so a crossing
// within a small fraction of a segment's length of one of its ends, on either
// side, counts as at that end: the parameter is snapped to 0 or 1, and the
// point returned is then that endpoint itself.
//
// ok is false if the two segments do not meet, if they are parallel, or so
// nearly parallel that the crossing point carries no useful accuracy, or if
// either pair of points coincides.  The zero vector and zero parameters are
// returned in that case, since no point on either segment is a better answer
// than another; a caller which has a fallback in mind should supply it itself.
func SegmentIntersection(a1, a2, b1, b2 vec.Vec2) (vec.Vec2, float64, float64, bool) {
	// The reasoning matches [Intersection]; only the powers of two need more
	// care here, because the two parameters are compared against a fixed
	// range instead of cancelling against a step along one of the lines.
	u, eu := a2.Sub(a1).Frexp()
	v, ev := b2.Sub(b1).Frexp()
	w, ew := b1.Sub(a1).Frexp()

	denom := u.Cross(v)
	if math.Abs(denom) <= parallelTol*u.Length()*v.Length() {
		return vec.Vec2{}, 0, 0, false
	}

	// The crossing is at a1 + t*(a2-a1) = b1 + s*(b2-b1).  Each quotient is
	// formed from the normalised directions, so the powers of two divided out
	// of them are put back afterwards.  An exponent too large for float64
	// gives an infinite parameter, which the range test below rejects just as
	// it would a large finite one.
	t := snapToEnd(math.Ldexp(w.Cross(v)/denom, ew-eu))
	s := snapToEnd(math.Ldexp(w.Cross(u)/denom, ew-ev))
	if !(t >= 0 && t <= 1 && s >= 0 && s <= 1) {
		return vec.Vec2{}, 0, 0, false
	}
	switch t {
	case 0:
		return a1, t, s, true
	case 1:
		return a2, t, s, true
	}
	return a1.Add(u.Mul(math.Ldexp(t, eu))), t, s, true
}

// snapToEnd moves a segment parameter within endTol of 0 or 1 onto that
// value.  A negative zero, which a vanishing cross product can leave behind,
// is replaced by a positive one on the way, so that a caller which divides
// by the parameter or looks at its sign sees the same thing whichever way
// the cross product rounded.
func snapToEnd(t float64) float64 {
	switch {
	case math.Abs(t) <= endTol:
		return 0
	case math.Abs(t-1) <= endTol:
		return 1
	}
	return t
}

// miterTol is the relative slack allowed on the miter limit.  A corner which
// sits exactly on the limit is meant to be mitered, but the ratio it is
// compared against is computed from the two directions and carries their
// rounding error, so an exact comparison would miter some such corners and
// bevel others.  The slack is relative, since the limit is a ratio which may
// be anywhere from 1 upwards and an absolute bound would mean nothing at the
// top of that range.
const miterTol = 1e-12

// MiterLength returns how far the miter join at a corner reaches from the
// corner itself, for a stroke of the given width.  in is the direction of the
// segment arriving at the corner and out the direction of the one leaving it;
// neither has to be a unit vector.
//
// ok is false where the join is bevelled instead: where either direction is
// zero and the corner has no angle to speak of, and where the miter is longer
// than miterLimit times the line width.  That limit is on the whole miter,
// from the inner corner of the join across to its tip, which is twice the
// length returned here.  The length returned when ok is false is half the line
// width, which is how far a bevelled join reaches.  A corner whose miter is
// within a rounding error of the limit counts as within it, and a miterLimit
// below 1 is shorter than any miter and bevels every corner.
func MiterLength(in, out vec.Vec2, lineWidth, miterLimit float64) (float64, bool) {
	// The directions are scaled to a common range first, so that their
	// lengths neither overflow for a corner given in large coordinates nor
	// underflow for one given in small ones.
	u, _ := in.Frexp()
	v, _ := out.Frexp()
	uLen, vLen := u.Length(), v.Length()
	if uLen == 0 || vLen == 0 {
		return lineWidth / 2, false
	}

	// The miter is lineWidth/sin(phi/2) long, where phi is the interior angle
	// the two segments enclose at the corner.  The direction turns by
	// theta = pi - phi there, so sin(phi/2) = cos(theta/2), and the sum of the
	// two unit directions has length 2*cos(theta/2).
	//
	// Half that sum is taken as the sine rather than sqrt((1+cos(theta))/2),
	// because a sharp corner leaves the latter with no significant digits:
	// there cos(theta) is close to -1, and adding one to it cancels away
	// everything the cosine had.  The sum of the directions cancels as well,
	// but component by component, and a subtraction of two nearby floats is
	// exact.  A corner which doubles back leaves no sum and no miter at all.
	sinHalf := u.Mul(1/uLen).Add(v.Mul(1/vLen)).Length() / 2
	if sinHalf <= 0 || 1/sinHalf > miterLimit*(1+miterTol) {
		return lineWidth / 2, false
	}
	return lineWidth / 2 / sinHalf, true
}
