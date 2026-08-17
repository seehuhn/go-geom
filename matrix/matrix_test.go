// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2023  Jochen Voss <voss@seehuhn.de>
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
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"seehuhn.de/go/geom/vec"
)

// TestMulOrder verifies that the matrix A.Mul(B) is equivalent to first
// applying A and then B.
func TestMulOrder(t *testing.T) {
	v0 := vec.Vec2{X: 1, Y: 2}
	for i, A := range testMatrices {
		for j, B := range testMatrices {
			t.Run(fmt.Sprintf("mat%d*mat%d", i, j), func(t *testing.T) {
				v1 := A.Mul(B).Apply(v0)
				v3 := B.Apply(A.Apply(v0))
				if math.Abs(v1.X-v3.X) > 1e-6 || math.Abs(v1.Y-v3.Y) > 1e-6 {
					t.Errorf("expected (%f, %f), got (%f, %f)", v3.X, v3.Y, v1.X, v1.Y)
				}
			})
		}
	}
}

func TestIdentityMatrix(t *testing.T) {
	for i, A := range testMatrices {
		t.Run(fmt.Sprintf("mat%d", i), func(t *testing.T) {
			B := A.Mul(Identity)
			if d := cmp.Diff(A, B); d != "" {
				t.Error(d)
			}
			C := Identity.Mul(A)
			if d := cmp.Diff(A, C); d != "" {
				t.Error(d)
			}
		})
	}
}

// TestMatrixInverse1 checks that a matrix multiplied by its inverse is the
// identity matrix.
func TestMatrixInverse1(t *testing.T) {
	for i, A := range testMatrices {
		t.Run(fmt.Sprintf("mat%d", i), func(t *testing.T) {
			Ainv, ok := A.Inv()
			if !ok {
				t.Fatal("not invertible")
			}

			B := Ainv.Mul(A)
			if d := cmp.Diff(Identity, B, cmpopts.EquateApprox(1e-6, 1e-6)); d != "" {
				t.Error(d)
			}

			B = A.Mul(Ainv)
			if d := cmp.Diff(Identity, B, cmpopts.EquateApprox(1e-6, 1e-6)); d != "" {
				t.Error(d)
			}
		})
	}
}

// TestMatrixInverse2 checks that the inverse of the inverse of a matrix is the
// original matrix.
func TestMatrixInverse2(t *testing.T) {
	for i, A := range testMatrices {
		t.Run(fmt.Sprintf("mat%d", i), func(t *testing.T) {
			Ainv, ok := A.Inv()
			if !ok {
				t.Fatal("not invertible")
			}
			B, ok := Ainv.Inv()
			if !ok {
				t.Fatal("inverse not invertible")
			}
			if d := cmp.Diff(A, B, cmpopts.EquateApprox(1e-6, 1e-6)); d != "" {
				t.Error(d)
			}
		})
	}
}

var testMatrices = []Matrix{
	Identity,
	{2, 3, 4, 5, 6, 7},
	Translate(-0.5, 0.5),
	Translate(0, 1),
	Translate(1, 0),
	Translate(1, 2),
	Scale(0.5, 0.5),
	Scale(2, 1),
	Scale(1, 2),
	Scale(3, 4),
	Scale(-1, -1),
	Rotate(0.1),
	Rotate(math.Pi / 2),
	Rotate(math.Pi),
}

func BenchmarkApply(b *testing.B) {
	M := Rotate(1)
	v := vec.Vec2{X: 2, Y: 3}
	for b.Loop() {
		v = M.Apply(v)
	}
}

// TestDetAccurate checks the determinant against exact arithmetic.  The
// result is a float64, so it is only correct to within a rounding error of
// the exact value; the point of the test is that no accuracy is lost where
// the two products cancel, which is where evaluating a*d-b*c directly fails.
func TestDetAccurate(t *testing.T) {
	eps := math.Ldexp(1, -52)
	for _, M := range []Matrix{
		Identity,
		Zero,
		{2, 0, 0, 3, 5, 7},
		{1, 1, 1, 1, 0, 0},
		Rotate(0.7),
		{0.1, 0.2, 0.30000000000000004, 0.6000000000000001, 0, 0},
		{1 + eps, 1 + 2*eps, 1, 1, 0, 0},
	} {
		exact := exactDet(M)
		got := M.Det()

		// the error must be at most half an ulp of the exact value
		diff := new(big.Float).SetPrec(200).Sub(exact, big.NewFloat(got).SetPrec(200))
		tol := new(big.Float).SetPrec(200).SetFloat64(math.Abs(math.Nextafter(got, math.Inf(1))-got) / 2)
		if diff.Abs(diff).Cmp(tol) > 0 {
			t.Errorf("Det(%v) = %v, want %v", M, got, exact)
		}
	}
}

// exactDet computes the determinant of the linear part of M without rounding.
// Products of two float64 values are exact at 200 bits of precision.
func exactDet(M Matrix) *big.Float {
	f := func(x float64) *big.Float { return new(big.Float).SetPrec(200).SetFloat64(x) }
	ad := new(big.Float).SetPrec(200).Mul(f(M[0]), f(M[3]))
	bc := new(big.Float).SetPrec(200).Mul(f(M[1]), f(M[2]))
	return new(big.Float).SetPrec(200).Sub(ad, bc)
}

// TestDetCancellation pins the determinant of a matrix which is exactly
// singular but whose two products only cancel after rounding.  Evaluating
// a*d-b*c directly gives a non-zero result here, and which non-zero result
// depends on whether the compiler contracts the expression, so this also
// keeps Det the same across architectures.
func TestDetCancellation(t *testing.T) {
	M := Matrix{0.1, 0.2, 0.30000000000000004, 0.6000000000000001, 0, 0}
	if naive := M[0]*M[3] - M[1]*M[2]; naive == 0 {
		t.Skip("this platform already evaluates a*d-b*c exactly for this matrix")
	}
	if got := M.Det(); got != 0 {
		t.Errorf("Det = %v, want 0", got)
	}
	if _, ok := M.Inv(); ok {
		t.Error("singular matrix reported as invertible")
	}
}

// TestInvSubnormalDet checks a matrix whose determinant is subnormal as a
// float64, and so has lost most of its significant digits.  Inv keeps the
// determinant as a mantissa and an exponent and never forms it, so the inverse
// is as accurate as for any other matrix.
func TestInvSubnormalDet(t *testing.T) {
	M := Matrix{1e-160, 0, 0, 1e-160, 0, 0}
	if det := M.Det(); det >= math.SmallestNonzeroFloat64*(1<<52) {
		t.Fatalf("determinant %v is not subnormal, the test is no longer meaningful", det)
	}
	inv, ok := M.Inv()
	if !ok {
		t.Fatal("not invertible")
	}
	want := 1 / 1e-160
	if inv[0] != want || inv[3] != want || inv[1] != 0 || inv[2] != 0 {
		t.Errorf("Inv = %v, want {%v, 0, 0, %v, 0, 0}", inv, want, want)
	}
}

// TestInvExtremeScale checks matrices whose determinant lies outside the range
// of float64 in either direction, but whose inverse is perfectly
// representable.  Forming the determinant as a float64 gives infinity or zero
// for these, and every element of the inverse then follows it.
func TestInvExtremeScale(t *testing.T) {
	for _, s := range []float64{1e-300, 1e-200, 1e-160, 1, 1e160, 1e200, 1e300} {
		M := Matrix{s, 0, 0, s, 7 * s, -3 * s}
		inv, ok := M.Inv()
		if !ok {
			t.Errorf("Scale(%g, %g) reported as not invertible", s, s)
			continue
		}
		// M maps (x, y) to (s*x+7*s, s*y-3*s), so the inverse maps (x, y)
		// back to (x/s-7, y/s+3)
		want := Matrix{1 / s, 0, 0, 1 / s, -7, 3}
		for i, got := range inv {
			if math.Abs(got-want[i]) > 1e-12*math.Abs(want[i]) {
				t.Errorf("scale %g: Inv = %v, want %v", s, inv, want)
				break
			}
		}
	}
}

// TestInvWideRange checks a matrix whose determinant rests on an element far
// smaller than the largest one.  Scaling the matrix to keep the determinant in
// range loses that element, so Inv has to fall back to M itself here.
func TestInvWideRange(t *testing.T) {
	// the determinant is 1e300*1e-300 - 1e-300*2e300 = 1 - 2 = -1
	M := Matrix{1e300, 1e-300, 2e300, 1e-300, 0, 0}
	if det := M.Det(); det != -1 {
		t.Fatalf("Det = %v, want -1", det)
	}
	inv, ok := M.Inv()
	if !ok {
		t.Fatal("not invertible")
	}
	want := Matrix{-1e-300, 1e-300, 2e300, -1e300, 0, 0}
	if inv != want {
		t.Errorf("Inv = %v, want %v", inv, want)
	}

	// SingularValues scales all four elements by one exponent and so loses the
	// small ones, where Inv falls back to M itself and keeps them.  The two
	// therefore disagree about this matrix, and both are right: its smaller
	// singular value is about 4e-601 and no float64 holds that.  Neither side
	// should be "fixed" to agree with the other without moving the other.
	if sigmaMin, _ := M.SingularValues(); sigmaMin != 0 {
		t.Errorf("sigmaMin = %v, want 0", sigmaMin)
	}
}

// TestDetExtremeScale checks that a singular matrix is recognised whatever the
// magnitude of its elements.  Forming the products without scaling gives NaN
// for the large cases, since each product on its own overflows.
func TestDetExtremeScale(t *testing.T) {
	for _, s := range []float64{1e-300, 1e-160, 1, 1e160, 1e300} {
		// the two columns are equal, so the matrix is singular
		M := Matrix{s, 2 * s, s, 2 * s, 0, 0}
		if got := M.Det(); got != 0 {
			t.Errorf("scale %g: Det = %v, want 0", s, got)
		}
		if inv, ok := M.Inv(); ok {
			t.Errorf("scale %g: singular matrix inverted to %v", s, inv)
		}
	}

	// a determinant which genuinely leaves the range of float64
	if got := (Matrix{1e200, 0, 0, 1e200, 0, 0}).Det(); !math.IsInf(got, 1) {
		t.Errorf("Det = %v, want +Inf", got)
	}
	if got := (Matrix{1e-200, 0, 0, 1e-200, 0, 0}).Det(); got != 0 {
		t.Errorf("Det = %v, want 0", got)
	}
}

// TestInvOverflow checks a matrix which is not singular, but whose inverse
// is too large to represent.
func TestInvOverflow(t *testing.T) {
	// the determinant is subnormal, so 1/det exceeds the range of float64
	M := Matrix{1e-320, 0, 0, 1, 0, 0}
	if M.Det() == 0 {
		t.Fatal("determinant underflowed to zero, the test is no longer meaningful")
	}
	if inv, ok := M.Inv(); ok {
		t.Errorf("Inv = %v, ok = true, want ok = false", inv)
	}
}

// TestInvNotInvertible checks that matrices without a usable inverse are
// reported, and that the returned matrix is Zero.
func TestInvNotInvertible(t *testing.T) {
	inf, nan := math.Inf(1), math.NaN()
	for _, M := range []Matrix{
		Zero,
		{1, 1, 1, 1, 0, 0},
		{2, 0, 4, 0, 10, 20},
		{0.1, 0.2, 0.30000000000000004, 0.6000000000000001, 0, 0},
		{0, 0, 0, 0, 3, 4}, // zero linear part, non-zero translation
		{inf, 0, 0, 1, 0, 0},
		{1, 0, 0, nan, 0, 0},
	} {
		inv, ok := M.Inv()
		if ok {
			t.Errorf("Inv(%v) = %v, ok = true, want ok = false", M, inv)
		} else if inv != Zero {
			t.Errorf("Inv(%v) returned %v, want Zero", M, inv)
		}
	}
}

// TestInvSingular checks that a singular matrix is never inverted, over a
// range of element magnitudes wide enough that the products which make up the
// determinant leave the range of float64 in both directions.
func TestInvSingular(t *testing.T) {
	vals := []float64{0, 1, -1, 0.5, 1e-320, 1e-160, 1e160, 1e300,
		math.Inf(1), math.Inf(-1), math.NaN()}
	for _, a := range vals {
		for _, b := range vals {
			for _, k := range []float64{0, 1, -1, 2, 0.5, 1e-8} {
				for _, e := range []float64{0, 3} {
					// the two columns are multiples of each other, so the
					// matrix is singular by construction
					M := Matrix{a, b, k * a, k * b, e, 4}
					if M.Det() != 0 {
						continue // rounding made it non-singular
					}
					if inv, ok := M.Inv(); ok {
						t.Fatalf("singular %v reported invertible, inverse %v", M, inv)
					}
				}
			}
		}
	}
}

func TestSingularValues(t *testing.T) {
	const tol = 1e-12
	cases := []struct {
		M                Matrix
		wantMin, wantMax float64
	}{
		{Identity, 1, 1},
		{Zero, 0, 0},
		{Scale(2, 5), 2, 5},
		{Scale(5, 2), 2, 5},
		{Scale(-3, 3), 3, 3},
		{Rotate(0.9), 1, 1},
		{Rotate(0.9).Scale(3, 7), 3, 7},
		{Matrix{1, 1, 1, 1, 0, 0}, 0, 2},
		// the translation must not affect the result
		{Matrix{2, 0, 0, 5, 1e6, -1e6}, 2, 5},
		// outside the range where a*a+b*b+c*c+d*d can be formed directly
		{Scale(1e200, 1e200), 1e200, 1e200},
		{Scale(1e-200, 1e-200), 1e-200, 1e-200},
		{Scale(1e300, 1e-300), 1e-300, 1e300},
	}
	for _, tc := range cases {
		gotMin, gotMax := tc.M.SingularValues()
		if math.Abs(gotMin-tc.wantMin) > tol*max(1, tc.wantMin) ||
			math.Abs(gotMax-tc.wantMax) > tol*max(1, tc.wantMax) {
			t.Errorf("SingularValues(%v) = %v, %v, want %v, %v",
				tc.M, gotMin, gotMax, tc.wantMin, tc.wantMax)
		}
	}
}

// TestSingularValuesAgainstApply checks sigmaMin and sigmaMax against the
// lengths the matrix actually produces, by sampling directions on the unit
// circle.  The sampled extremes approach the true ones from the inside, so
// the tolerance only needs to cover the gap between samples.
func TestSingularValuesAgainstApply(t *testing.T) {
	const nSamples = 4000
	for _, M := range testMatrices {
		sigmaMin, sigmaMax := M.SingularValues()
		linear := Matrix{M[0], M[1], M[2], M[3], 0, 0}
		obsMin, obsMax := math.Inf(1), 0.0
		for i := range nSamples {
			phi := math.Pi * float64(i) / nSamples
			v := linear.Apply(vec.Vec2{X: math.Cos(phi), Y: math.Sin(phi)})
			l := math.Hypot(v.X, v.Y)
			obsMin = min(obsMin, l)
			obsMax = max(obsMax, l)
		}
		// the length varies quadratically near an extremum, so the sampled
		// value is within O(step^2) of the true one
		step := math.Pi / nSamples
		tol := 4 * step * step * max(1, sigmaMax)
		if obsMin < sigmaMin-tol || obsMin > sigmaMin+tol ||
			obsMax < sigmaMax-tol || obsMax > sigmaMax+tol {
			t.Errorf("SingularValues(%v) = %v, %v, sampled %v, %v",
				M, sigmaMin, sigmaMax, obsMin, obsMax)
		}
	}
}

func TestSingularValuesNonFinite(t *testing.T) {
	inf, nan := math.Inf(1), math.NaN()
	for _, M := range []Matrix{
		{inf, 0, 0, 1, 0, 0},
		{1, 0, 0, nan, 0, 0},
		{nan, nan, nan, nan, 0, 0},
	} {
		gotMin, gotMax := M.SingularValues()
		if gotMin != 0 || gotMax != 0 {
			t.Errorf("SingularValues(%v) = %v, %v, want 0, 0", M, gotMin, gotMax)
		}
	}
}

func BenchmarkInv(b *testing.B) {
	M := Rotate(1).Scale(2, 3).Translate(4, 5)
	for b.Loop() {
		M.Inv()
	}
}

func BenchmarkSingularValues(b *testing.B) {
	M := Rotate(1).Scale(2, 3).Translate(4, 5)
	for b.Loop() {
		M.SingularValues()
	}
}
