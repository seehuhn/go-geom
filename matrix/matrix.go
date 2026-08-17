// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2024  Jochen Voss <voss@seehuhn.de>
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

package matrix

import (
	"math"

	"seehuhn.de/go/geom/vec"
)

// Matrix contains a PDF transformation matrix.
// The elements are stored in the same order as for the "cm" operator.
//
// A matrix object M = [a b c d e f] corresponds to the following
// 3x3 matrix:
//
//	/ a c e \
//	| b d f |
//	\ 0 0 1 /
//
// A vector (x, y, 1) is transformed by M into
//
//	M * (x, y, 1) = (a*x+c*y+e, b*x+d*y+f, 1)
type Matrix [6]float64

// IsZero reports whether all elements of M are zero.
func (M Matrix) IsZero() bool {
	return M == Zero
}

// Apply applies the transformation matrix to the given vector.
func (M Matrix) Apply(v vec.Vec2) vec.Vec2 {
	return vec.Vec2{
		X: v.X*M[0] + v.Y*M[2] + M[4],
		Y: v.X*M[1] + v.Y*M[3] + M[5],
	}
}

// Mul multiplies two transformation matrices and returns the result.
// The result is equivalent to first applying M and then B.
// Mathematically: (B * M) * v = B * (M * v).
func (M Matrix) Mul(B Matrix) Matrix {
	return Matrix{
		M[0]*B[0] + M[1]*B[2],
		M[0]*B[1] + M[1]*B[3],
		M[2]*B[0] + M[3]*B[2],
		M[2]*B[1] + M[3]*B[3],
		M[4]*B[0] + M[5]*B[2] + B[4],
		M[4]*B[1] + M[5]*B[3] + B[5],
	}
}

// Translate applies a translation after the transformation matrix M.
func (M Matrix) Translate(dx, dy float64) Matrix {
	B := Translate(dx, dy)
	return M.Mul(B)
}

// Scale applies a scaling after the transformation matrix M.
func (M Matrix) Scale(xScale, yScale float64) Matrix {
	B := Scale(xScale, yScale)
	return M.Mul(B)
}

// Rotate applies a rotation after the transformation matrix M.
func (M Matrix) Rotate(phi float64) Matrix {
	B := Rotate(phi)
	return M.Mul(B)
}

// RotateDeg applies a rotation (in degrees) after the transformation matrix M.
func (M Matrix) RotateDeg(phi float64) Matrix {
	B := RotateDeg(phi)
	return M.Mul(B)
}

// col returns the i-th column of the linear part of M, the image of the i-th
// basis vector.
func (M Matrix) col(i int) vec.Vec2 {
	return vec.Vec2{X: M[2*i], Y: M[2*i+1]}
}

// offset returns the translation part of M, the image of the origin.
func (M Matrix) offset() vec.Vec2 {
	return vec.Vec2{X: M[4], Y: M[5]}
}

// isFinite reports whether every element of M is a finite number.
func (M Matrix) isFinite() bool {
	for _, x := range M {
		if math.IsInf(x, 0) || math.IsNaN(x) {
			return false
		}
	}
	return true
}

// Det returns the determinant of the linear part of M, the factor by which M
// scales areas.  It is negative if M swaps the two orientations of the plane.
//
// A singular M gives exactly zero whatever the magnitude of its elements.  The
// converse does not hold: the determinant of a matrix with very small elements
// underflows to zero, and that of one with very large elements overflows to
// infinity, even where the matrix is invertible.  Use [Matrix.Inv] to invert a
// matrix and [Matrix.SingularValues] to judge how close to singular it is;
// both cope with a determinant which no float64 can hold.
func (M Matrix) Det() float64 {
	return M.col(0).Cross(M.col(1))
}

// SingularValues returns the smallest and largest singular values of the linear
// part of M: the factors by which M scales lengths along the directions where
// the scaling is weakest and strongest.  The translation part of M does not
// affect either value.
//
// A singular M gives a sigmaMin of zero, since M then flattens some direction
// away entirely.  The converse does not hold: sigmaMin underflows to zero for a
// matrix whose smallest singular value is too small to represent beside its
// largest, and [Matrix.Inv] still inverts some of those exactly.  Read a
// sigmaMin of zero as "singular as far as a float64 can tell", not as proof
// that M has no inverse.
//
// Both values are zero if any element of the linear part of M is infinite or
// not a number.
func (M Matrix) SingularValues() (sigmaMin, sigmaMax float64) {
	// The elements are scaled by a power of two, which is exact, so that the
	// squares below neither overflow nor underflow for a matrix whose own
	// singular values are in range.  This also rejects infinities and NaNs,
	// since the comparison is false for both.
	scale := max(math.Abs(M[0]), math.Abs(M[1]), math.Abs(M[2]), math.Abs(M[3]))
	if !(scale > 0) || math.IsInf(scale, 1) {
		return 0, 0
	}
	_, e := math.Frexp(scale)
	a := math.Ldexp(M[0], -e)
	b := math.Ldexp(M[1], -e)
	c := math.Ldexp(M[2], -e)
	d := math.Ldexp(M[3], -e)

	t := a*a + b*b + c*c + d*d
	det := vec.Vec2{X: a, Y: b}.Cross(vec.Vec2{X: c, Y: d})
	// The largest scaled element has magnitude at least 1/2, so t is at least
	// 1/4 and sigmaMax at least sqrt(1/8), even where the root above vanishes.
	// It is therefore never zero, and the division below is safe.
	sigmaMax = math.Sqrt(max(0, (t+math.Sqrt(max(0, t*t-4*det*det)))/2))

	// sigmaMin is derived from the determinant rather than from the difference
	// under the root, which cancels away to nothing for a nearly singular
	// matrix
	sigmaMin = math.Abs(det) / sigmaMax
	return math.Ldexp(sigmaMin, e), math.Ldexp(sigmaMax, e)
}

// Inv computes the inverse of the transformation matrix M.  ok is false if M
// is singular, if any of its elements is infinite or not a number, or if its
// inverse overflows the range of float64.  The returned matrix is then Zero,
// which collapses every point onto the origin, so that a caller which ignores
// ok produces visibly degenerate output rather than plausible nonsense.  Zero
// is never returned with ok true.
//
// An inverse which is merely inaccurate is still returned with ok true.
// Whether to use it is left to the caller, which alone knows what the result
// is for: [Matrix.SingularValues] reports how close to singular M is, and the
// inverse loses roughly as many digits as separate sigmaMax from sigmaMin.
func (M Matrix) Inv() (Matrix, bool) {
	if !M.isFinite() {
		return Zero, false
	}

	// The determinant of an invertible matrix can overflow or underflow the
	// range of float64 even where the inverse is perfectly representable, and
	// dividing by it then leaves every element of the inverse infinite or
	// zero.  Scaling the linear part by a power of two, which is exact, keeps
	// the determinant in range.  It scales the whole inverse by the reciprocal
	// factor, so the result is scaled back by the same amount.
	scale := max(math.Abs(M[0]), math.Abs(M[1]), math.Abs(M[2]), math.Abs(M[3]))
	if scale > 0 {
		_, k := math.Frexp(scale)
		S := Matrix{
			math.Ldexp(M[0], -k), math.Ldexp(M[1], -k),
			math.Ldexp(M[2], -k), math.Ldexp(M[3], -k),
			M[4], M[5],
		}
		if inv, ok := S.invUnscaled(); ok {
			for i := range inv {
				inv[i] = math.Ldexp(inv[i], -k)
			}
			if inv.isFinite() {
				return inv, true
			}
		}
	}

	// The scaling drops an element more than 2**1074 times smaller than the
	// largest one, which can be the very element the determinant depends on.
	// M itself may still have a determinant in range.
	return M.invUnscaled()
}

// invUnscaled inverts M by dividing through its determinant as a float64.  ok
// is false if the determinant is zero or has left the range of float64, or if
// any element of the inverse overflows.  M must have no infinite or NaN
// element, which is what keeps the determinant from being NaN here.
func (M Matrix) invUnscaled() (Matrix, bool) {
	det := M.Det()
	if det == 0 || math.IsInf(det, 0) {
		return Zero, false
	}

	// The elements are divided by det rather than multiplied by 1/det: the
	// reciprocal costs a rounding error in every element, and overflows to
	// infinity for a subnormal determinant whose inverse is representable.
	inv := Matrix{
		M[3] / det, -M[1] / det,
		-M[2] / det, M[0] / det,
		M.col(1).Cross(M.offset()) / det,
		-M.col(0).Cross(M.offset()) / det,
	}
	if !inv.isFinite() {
		return Zero, false
	}
	return inv, true
}

// Identity is the identity transformation.
var Identity = Matrix{1, 0, 0, 1, 0, 0}

// Zero is an uninitialized transformation.
var Zero = Matrix{0, 0, 0, 0, 0, 0}

// Translate moves the origin of the coordinate system.
//
// Drawing the unit square [0, 1] x [0, 1] after applying this transformation
// is equivalent to drawing the rectangle [dx, dx+1] x [dy, dy+1] in the
// original coordinate system.
func Translate(dx, dy float64) Matrix {
	return Matrix{1, 0, 0, 1, dx, dy}
}

// Scale scales the coordinate system.
//
// Drawing the unit square [0, 1] x [0, 1] after applying this transformation
// is equivalent to drawing the rectangle [0, xScale] x [0, yScale] in the
// original coordinate system.
func Scale(xScale, yScale float64) Matrix {
	return Matrix{xScale, 0, 0, yScale, 0, 0}
}

// Rotate rotates the coordinate system by the given angle (in radians).
func Rotate(phi float64) Matrix {
	c := math.Cos(phi)
	s := math.Sin(phi)
	return Matrix{c, s, -s, c, 0, 0}
}

// RotateDeg rotates the coordinate system by the given angle (in degrees).
func RotateDeg(phi float64) Matrix {
	phi *= math.Pi / 180
	return Rotate(phi)
}
