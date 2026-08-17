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
