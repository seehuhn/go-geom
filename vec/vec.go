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

// minNormal is the smallest positive normalised float64.  A sum of squares
// below this has lost precision to underflow, even where it is not yet zero.
const minNormal = math.SmallestNonzeroFloat64 * (1 << 52)

// Length returns the length of the vector.
func (v Vec2) Length() float64 {
	s := v.X*v.X + v.Y*v.Y
	if s >= minNormal && s <= math.MaxFloat64 {
		return math.Sqrt(s)
	}
	if v == (Vec2{}) {
		return 0
	}
	// The squares overflowed, or lost precision by falling into the subnormal
	// range.  math.Hypot scales the components first and so keeps the length
	// of any vector whose length is representable, at about ten times the
	// cost.  Non-finite components reach this point too, and Hypot gives the
	// IEEE 754 result for them.
	return math.Hypot(v.X, v.Y)
}

// Add returns the vector sum of v and other.
func (v Vec2) Add(other Vec2) Vec2 {
	return Vec2{X: v.X + other.X, Y: v.Y + other.Y}
}

// Sub returns the vector difference of v and other.
func (v Vec2) Sub(other Vec2) Vec2 {
	return Vec2{X: v.X - other.X, Y: v.Y - other.Y}
}

// Mul returns the vector scaled by c.
func (v Vec2) Mul(c float64) Vec2 {
	return Vec2{X: v.X * c, Y: v.Y * c}
}

// Neg returns the negation of the vector.
func (v Vec2) Neg() Vec2 {
	return Vec2{X: -v.X, Y: -v.Y}
}

// IMul scales the vector in place by c.
func (v *Vec2) IMul(c float64) {
	v.X *= c
	v.Y *= c
}

// Dot calculates the dot product of two vectors.  It is zero exactly when the
// two are perpendicular, unless the true value is too small to represent as a
// float64, in which case it underflows to zero as well.
//
// The two products are combined with fused multiply-adds, so that the rounding
// error of one is carried into the other instead of being lost.  The result is
// therefore accurate to within a rounding error even for nearly perpendicular
// vectors, where the two products cancel.  It also does not depend on whether
// the compiler contracts a*b+c*d into a fused operation, which differs between
// architectures.
//
// Where a product overflows, the vectors are scaled by a power of two and the
// result scaled back, so that an intermediate value outside the range of
// float64 does not turn a result which is inside it into infinity or NaN.
func (v Vec2) Dot(other Vec2) float64 {
	r := dot(v, other)
	if math.Abs(r) <= math.MaxFloat64 {
		return r
	}

	// One of the products overflowed, either on its own or into a NaN against
	// the other.  As in [Vec2.Cross], scaling has to stay off the normal path:
	// it drops a component far smaller than its partner, and only here, where
	// both products must be enormous, can that component not matter.
	a, ea := v.Frexp()
	b, eb := other.Frexp()
	return math.Ldexp(dot(a, b), ea+eb)
}

// dot calculates the dot product of two vectors without guarding against an
// intermediate overflow.
func dot(v, other Vec2) float64 {
	w := v.X * other.X
	e := math.FMA(v.X, other.X, -w) // the part of v.X*other.X which w lost
	f := math.FMA(v.Y, other.Y, w)
	return f + e
}

// Frexp splits v into a normalised vector and a binary exponent, so that v is
// recovered by scaling each component of frac by 2**exp.  The larger component
// of frac has magnitude in [0.5, 1), so that no product formed from the
// components of two such vectors can overflow.
//
// The split is exact unless one component is more than 2**1074 times smaller
// than the other, in which case the smaller one underflows to zero.  It is
// then too small to affect the direction of v by any representable angle, so
// the loss does not matter where v stands for a direction.  It does matter
// where the components are treated separately, as in a determinant.
//
// The zero vector is returned unchanged with exponent zero, as is a vector with
// an infinite or NaN component, neither of which has a meaningful scale to
// divide out.
func (v Vec2) Frexp() (frac Vec2, exp int) {
	scale := max(math.Abs(v.X), math.Abs(v.Y))
	if !(scale > 0) || scale > math.MaxFloat64 {
		return v, 0
	}
	_, exp = math.Frexp(scale)
	return Vec2{X: math.Ldexp(v.X, -exp), Y: math.Ldexp(v.Y, -exp)}, exp
}

// Cross calculates the cross product of two vectors: the signed area of the
// parallelogram they span, positive if other lies counter-clockwise from v.
// It is zero exactly when the two are parallel, unless the true area is too
// small to represent as a float64, in which case it underflows to zero as well.
//
// The two products are combined with fused multiply-adds, so that the rounding
// error of one is subtracted from the other instead of accumulating.  The
// result is therefore accurate to within a rounding error even for nearly
// parallel vectors, where evaluating the expression directly can lose every
// significant digit.  It also does not depend on whether the compiler
// contracts a*b-c*d into a fused operation, which differs between
// architectures.
//
// Where a product overflows, the vectors are scaled by a power of two and the
// result scaled back, so that an intermediate value outside the range of
// float64 does not turn an area which is inside it into infinity or NaN.
func (v Vec2) Cross(other Vec2) float64 {
	r := cross(v, other)
	if math.Abs(r) <= math.MaxFloat64 {
		return r
	}

	// One of the products overflowed, leaving an infinity which the other
	// product turned into a NaN, or an infinity of its own.  The vectors are
	// brought into a range where the products stay finite; the area itself
	// may still overflow, and then the result is infinite for good reason.
	//
	// Scaling cannot be the normal path.  It reduces each vector by its
	// larger component, which drops a component small enough to be lost
	// beside it, and the dropped one can still carry the answer where the
	// large products cancel.  That case reaches here only with both products
	// overflowing, where the small term cannot matter.
	a, ea := v.Frexp()
	b, eb := other.Frexp()
	return math.Ldexp(cross(a, b), ea+eb)
}

// cross calculates the cross product of two vectors without guarding against
// an intermediate overflow.
func cross(v, other Vec2) float64 {
	w := v.Y * other.X
	e := math.FMA(-v.Y, other.X, w) // the part of v.Y*other.X which w lost
	f := math.FMA(v.X, other.Y, -w)
	return f + e
}

// Rot90 returns the vector rotated 90 degrees counter-clockwise.
func (v Vec2) Rot90() Vec2 {
	return Vec2{X: -v.Y, Y: v.X}
}

// Normalize returns a unit vector in the same direction as v.  Only the zero
// vector has no direction, and the zero vector is returned for it.
//
// The elements are divided by the length rather than multiplied by its
// reciprocal, which would overflow to infinity for a vector short enough that
// its length is subnormal.
func (v Vec2) Normalize() Vec2 {
	length := v.Length()
	if !(length > 0) {
		return Vec2{0, 0}
	}
	return Vec2{X: v.X / length, Y: v.Y / length}
}

// Normal returns a unit vector perpendicular to v, rotated 90 degrees
// counter-clockwise.  Only the zero vector has no direction, and the zero
// vector is returned for it.
func (v Vec2) Normal() Vec2 {
	length := v.Length()
	if !(length > 0) {
		return Vec2{0, 0}
	}
	return Vec2{X: -v.Y / length, Y: v.X / length}
}

// Middle returns the midpoint between vectors a and b.
func Middle(a, b Vec2) Vec2 {
	return Vec2{X: middle(a.X, b.X), Y: middle(a.Y, b.Y)}
}

// middle returns the midpoint of two coordinates.  The sum is halved where it
// can be formed at all, which is exact, and each coordinate is halved on its
// own where the sum would overflow two points which both lie in the range of
// float64.
func middle(x, y float64) float64 {
	if s := x + y; math.Abs(s) <= math.MaxFloat64 {
		return s / 2
	}
	return x/2 + y/2
}
