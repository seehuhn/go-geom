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

import (
	"fmt"
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLength(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2
		want float64
	}{
		{"zero vector", Vec2{0, 0}, 0},
		{"unit x", Vec2{1, 0}, 1},
		{"unit y", Vec2{0, 1}, 1},
		{"diagonal", Vec2{3, 4}, 5},
		{"negative coords", Vec2{-3, -4}, 5},
		{"mixed signs", Vec2{-3, 4}, 5},
		{"small vector", Vec2{0.6, 0.8}, 1},
		{"large vector", Vec2{30, 40}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Length()
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Length() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		v1   Vec2
		v2   Vec2
		want Vec2
	}{
		{"zero vectors", Vec2{0, 0}, Vec2{0, 0}, Vec2{0, 0}},
		{"add to zero", Vec2{0, 0}, Vec2{3, 4}, Vec2{3, 4}},
		{"positive coords", Vec2{1, 2}, Vec2{3, 4}, Vec2{4, 6}},
		{"negative coords", Vec2{-1, -2}, Vec2{-3, -4}, Vec2{-4, -6}},
		{"mixed signs", Vec2{-1, 2}, Vec2{3, -4}, Vec2{2, -2}},
		{"fractional coords", Vec2{0.5, 0.3}, Vec2{0.2, 0.7}, Vec2{0.7, 1.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Add(tt.v2)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSub(t *testing.T) {
	tests := []struct {
		name string
		v1   Vec2
		v2   Vec2
		want Vec2
	}{
		{"zero vectors", Vec2{0, 0}, Vec2{0, 0}, Vec2{0, 0}},
		{"sub from zero", Vec2{0, 0}, Vec2{3, 4}, Vec2{-3, -4}},
		{"positive coords", Vec2{5, 7}, Vec2{2, 3}, Vec2{3, 4}},
		{"negative coords", Vec2{-1, -2}, Vec2{-3, -4}, Vec2{2, 2}},
		{"mixed signs", Vec2{-1, 2}, Vec2{3, -4}, Vec2{-4, 6}},
		{"same vector", Vec2{3, 4}, Vec2{3, 4}, Vec2{0, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Sub(tt.v2)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Sub() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMul(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2
		c    float64
		want Vec2
	}{
		{"zero vector", Vec2{0, 0}, 5.0, Vec2{0, 0}},
		{"multiply by zero", Vec2{3, 4}, 0.0, Vec2{0, 0}},
		{"multiply by one", Vec2{3, 4}, 1.0, Vec2{3, 4}},
		{"multiply by two", Vec2{3, 4}, 2.0, Vec2{6, 8}},
		{"multiply by half", Vec2{6, 8}, 0.5, Vec2{3, 4}},
		{"multiply by negative", Vec2{3, 4}, -1.0, Vec2{-3, -4}},
		{"negative vector", Vec2{-2, -3}, 2.0, Vec2{-4, -6}},
		{"fractional scalar", Vec2{2, 3}, 1.5, Vec2{3, 4.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Mul(tt.c)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Mul() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestIMul(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2
		c    float64
		want Vec2
	}{
		{"zero vector", Vec2{0, 0}, 5.0, Vec2{0, 0}},
		{"multiply by zero", Vec2{3, 4}, 0.0, Vec2{0, 0}},
		{"multiply by one", Vec2{3, 4}, 1.0, Vec2{3, 4}},
		{"multiply by two", Vec2{3, 4}, 2.0, Vec2{6, 8}},
		{"multiply by negative", Vec2{3, 4}, -1.0, Vec2{-3, -4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.v // copy
			v.IMul(tt.c)
			if diff := cmp.Diff(tt.want, v, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("IMul() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDot(t *testing.T) {
	tests := []struct {
		name string
		v1   Vec2
		v2   Vec2
		want float64
	}{
		{"zero vectors", Vec2{0, 0}, Vec2{0, 0}, 0},
		{"one zero vector", Vec2{3, 4}, Vec2{0, 0}, 0},
		{"orthogonal vectors", Vec2{1, 0}, Vec2{0, 1}, 0},
		{"same vector", Vec2{3, 4}, Vec2{3, 4}, 25},
		{"opposite vectors", Vec2{3, 4}, Vec2{-3, -4}, -25},
		{"unit vectors", Vec2{1, 0}, Vec2{1, 0}, 1},
		{"perpendicular unit vectors", Vec2{1, 0}, Vec2{0, 1}, 0},
		{"general case", Vec2{2, 3}, Vec2{4, 5}, 23}, // 2*4 + 3*5 = 8 + 15 = 23
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Dot(tt.v2)
			if math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("Dot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormal(t *testing.T) {
	tests := []struct {
		name string
		v    Vec2
		want Vec2
	}{
		{"unit x", Vec2{1, 0}, Vec2{0, 1}},
		{"unit y", Vec2{0, 1}, Vec2{-1, 0}},
		{"diagonal", Vec2{1, 1}, Vec2{-1 / math.Sqrt(2), 1 / math.Sqrt(2)}},
		{"scaled vector", Vec2{3, 4}, Vec2{-4.0 / 5, 3.0 / 5}},
		{"negative vector", Vec2{-3, -4}, Vec2{4.0 / 5, -3.0 / 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v.Normal()
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Normal() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNormalZeroVector(t *testing.T) {
	v := Vec2{0, 0}
	got := v.Normal()
	want := Vec2{0, 0}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Normal() for zero vector mismatch (-want +got):\n%s", diff)
	}
}

// verySmallLengths are lengths short enough that an absolute cut-off would
// discard them, down to the smallest positive float64.
var verySmallLengths = []float64{1e-12, 1e-100, 1e-300, 5e-324}

// TestNormalVerySmallVector checks that a short vector still has a direction.
// Its length carries no information about how well that direction is known,
// which depends on the coordinates the vector was formed from.
func TestNormalVerySmallVector(t *testing.T) {
	for _, s := range verySmallLengths {
		v := Vec2{s, 0}
		want := Vec2{0, 1}
		if diff := cmp.Diff(want, v.Normal()); diff != "" {
			t.Errorf("Normal() for length %g mismatch (-want +got):\n%s", s, diff)
		}
	}
}

func TestNormalizeVerySmallVector(t *testing.T) {
	for _, s := range verySmallLengths {
		v := Vec2{s, 0}
		want := Vec2{1, 0}
		if diff := cmp.Diff(want, v.Normalize()); diff != "" {
			t.Errorf("Normalize() for length %g mismatch (-want +got):\n%s", s, diff)
		}
	}
}

func TestNormalizeZeroVector(t *testing.T) {
	if got := (Vec2{0, 0}).Normalize(); got != (Vec2{0, 0}) {
		t.Errorf("Normalize() for zero vector = %v, want the zero vector", got)
	}
}

func TestMiddle(t *testing.T) {
	tests := []struct {
		name string
		a    Vec2
		b    Vec2
		want Vec2
	}{
		{"same point", Vec2{3, 4}, Vec2{3, 4}, Vec2{3, 4}},
		{"origin and point", Vec2{0, 0}, Vec2{4, 6}, Vec2{2, 3}},
		{"negative coords", Vec2{-2, -4}, Vec2{2, 4}, Vec2{0, 0}},
		{"general case", Vec2{1, 2}, Vec2{5, 8}, Vec2{3, 5}},
		{"fractional result", Vec2{1, 1}, Vec2{2, 3}, Vec2{1.5, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Middle(tt.a, tt.b)
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Middle() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestVecProperties(t *testing.T) {
	testVectors := []Vec2{
		{0, 0},     // zero vector
		{1, 0},     // unit x
		{0, 1},     // unit y
		{1, 1},     // diagonal
		{3, 4},     // 3-4-5 triangle
		{-2, 3},    // mixed signs
		{0.1, 0.2}, // small values
		{100, 200}, // large values
	}

	t.Run("add is commutative", func(t *testing.T) {
		for i, v1 := range testVectors {
			for j, v2 := range testVectors {
				t.Run(fmt.Sprintf("vec%d+vec%d", i, j), func(t *testing.T) {
					a := v1.Add(v2)
					b := v2.Add(v1)
					if diff := cmp.Diff(a, b, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
						t.Errorf("addition not commutative (-v1+v2 +v2+v1):\n%s", diff)
					}
				})
			}
		}
	})

	t.Run("add is associative", func(t *testing.T) {
		v1, v2, v3 := Vec2{1, 2}, Vec2{3, 4}, Vec2{5, 6}

		a := v1.Add(v2).Add(v3)
		b := v1.Add(v2.Add(v3))

		if diff := cmp.Diff(a, b, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
			t.Errorf("addition not associative (-(v1+v2)+v3 +v1+(v2+v3)):\n%s", diff)
		}
	})

	t.Run("dot is commutative", func(t *testing.T) {
		for i, v1 := range testVectors {
			for j, v2 := range testVectors {
				t.Run(fmt.Sprintf("vec%d·vec%d", i, j), func(t *testing.T) {
					a := v1.Dot(v2)
					b := v2.Dot(v1)
					if math.Abs(a-b) > 1e-9 {
						t.Errorf("dot product not commutative: %v·%v = %v, %v·%v = %v", v1, v2, a, v2, v1, b)
					}
				})
			}
		}
	})

	t.Run("normal is perpendicular", func(t *testing.T) {
		for i, v := range testVectors {
			t.Run(fmt.Sprintf("vec%d", i), func(t *testing.T) {
				if v.Length() < 1e-9 {
					t.Skip("skip zero/very small vector")
				}

				normal := v.Normal()
				dot := v.Dot(normal)

				if math.Abs(dot) > 1e-9 {
					t.Errorf("normal not perpendicular: %v·%v = %v, want ~0", v, normal, dot)
				}
			})
		}
	})

	t.Run("normal has unit length", func(t *testing.T) {
		for i, v := range testVectors {
			t.Run(fmt.Sprintf("vec%d", i), func(t *testing.T) {
				if v.Length() < 1e-9 {
					t.Skip("skip zero/very small vector")
				}

				normal := v.Normal()
				length := normal.Length()

				if math.Abs(length-1.0) > 1e-9 {
					t.Errorf("normal not unit length: |%v| = %v, want 1", normal, length)
				}
			})
		}
	})

	t.Run("mul and imul are equivalent", func(t *testing.T) {
		scalars := []float64{0, 1, -1, 2, 0.5, -2.5}

		for i, v := range testVectors {
			for _, c := range scalars {
				t.Run(fmt.Sprintf("vec%d*%g", i, c), func(t *testing.T) {
					a := v.Mul(c)

					b := v // copy
					b.IMul(c)

					if diff := cmp.Diff(a, b, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
						t.Errorf("Mul() and IMul() not equivalent (-Mul +IMul):\n%s", diff)
					}
				})
			}
		}
	})

	t.Run("vector operations preserve zero", func(t *testing.T) {
		zero := Vec2{0, 0}
		v := Vec2{3, 4}

		if !cmp.Equal(zero.Add(v), v) {
			t.Error("0 + v ≠ v")
		}
		if !cmp.Equal(v.Add(zero), v) {
			t.Error("v + 0 ≠ v")
		}
		if !cmp.Equal(v.Sub(v), zero, cmpopts.EquateApprox(1e-9, 1e-9)) {
			t.Error("v - v ≠ 0")
		}
		if zero.Dot(v) != 0 {
			t.Error("0 · v ≠ 0")
		}
	})

	t.Run("middle is symmetric", func(t *testing.T) {
		for i, v1 := range testVectors {
			for j, v2 := range testVectors {
				t.Run(fmt.Sprintf("vec%d_vec%d", i, j), func(t *testing.T) {
					a := Middle(v1, v2)
					b := Middle(v2, v1)
					if diff := cmp.Diff(a, b, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
						t.Errorf("Middle() not symmetric (-Middle(v1,v2) +Middle(v2,v1)):\n%s", diff)
					}
				})
			}
		}
	})
}

func BenchmarkLength(b *testing.B) {
	v := Vec2{3, 4}
	for b.Loop() {
		_ = v.Length()
	}
}

func BenchmarkAdd(b *testing.B) {
	v1 := Vec2{1, 2}
	v2 := Vec2{3, 4}
	for b.Loop() {
		_ = v1.Add(v2)
	}
}

func BenchmarkDot(b *testing.B) {
	v1 := Vec2{1, 2}
	v2 := Vec2{3, 4}
	for b.Loop() {
		_ = v1.Dot(v2)
	}
}

func BenchmarkNormal(b *testing.B) {
	v := Vec2{3, 4}
	for b.Loop() {
		_ = v.Normal()
	}
}

func TestCross(t *testing.T) {
	tests := []struct {
		name string
		v1   Vec2
		v2   Vec2
		want float64
	}{
		{"zero vectors", Vec2{0, 0}, Vec2{0, 0}, 0},
		{"one zero vector", Vec2{3, 4}, Vec2{0, 0}, 0},
		{"parallel vectors", Vec2{3, 4}, Vec2{6, 8}, 0},
		{"antiparallel vectors", Vec2{3, 4}, Vec2{-3, -4}, 0},
		{"counter-clockwise", Vec2{1, 0}, Vec2{0, 1}, 1},
		{"clockwise", Vec2{0, 1}, Vec2{1, 0}, -1},
		{"general case", Vec2{2, 3}, Vec2{4, 5}, -2}, // 2*5 - 3*4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.v1.Cross(tt.v2)
			if got != tt.want {
				t.Errorf("Cross() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCrossParallel checks that the cross product of exactly parallel vectors
// is exactly zero, even where the two products cancel only after rounding.
// Evaluating v.X*w.Y-v.Y*w.X directly gives a non-zero result for these.
func TestCrossParallel(t *testing.T) {
	for _, tc := range []struct{ v, w Vec2 }{
		{Vec2{0.1, 0.2}, Vec2{0.30000000000000004, 0.6000000000000001}},
		{Vec2{1e8, 3e8}, Vec2{2e8, 6e8}},
		{Vec2{0.1, 0.2}, Vec2{0.4, 0.8}},
		{Vec2{7, 7}, Vec2{7, 7}},
	} {
		if got := tc.v.Cross(tc.w); got != 0 {
			t.Errorf("Cross(%v, %v) = %v, want 0", tc.v, tc.w, got)
		}
	}
}

// TestCrossExtremeScale checks vectors whose products lie outside the range of
// float64, where forming v.X*w.Y-v.Y*w.X directly gives infinity or NaN even
// though the area itself is representable.
func TestCrossExtremeScale(t *testing.T) {
	for _, tc := range []struct {
		name string
		v, w Vec2
		want float64
	}{
		// each product overflows on its own, but they cancel exactly
		{"large parallel", Vec2{1e300, 1e300}, Vec2{1e300, 1e300}, 0},
		{"large antiparallel", Vec2{3e200, 4e200}, Vec2{-6e200, -8e200}, 0},
		// one component huge, the other tiny
		{"mixed scales", Vec2{1e300, 0}, Vec2{0, 1e-300}, 1e300 * 1e-300},
		// both products underflow, and so does their difference
		{"small parallel", Vec2{1e-200, 2e-200}, Vec2{3e-200, 6e-200}, 0},
		{"subnormal components", Vec2{5e-324, 0}, Vec2{0, 5e-324}, 0},
		// The answer rests entirely on a component too small to survive being
		// scaled against the other, so it has to be reached without scaling.
		{"answer in the small component", Vec2{1e300, 1e-300}, Vec2{1, 0}, -1e-300},
		{"small component against zero", Vec2{1e-300, 1e300}, Vec2{0, 1}, 1e-300},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.Cross(tc.w); got != tc.want {
				t.Errorf("Cross(%v, %v) = %v, want %v", tc.v, tc.w, got, tc.want)
			}
		})
	}
}

// TestCrossScaleExact checks that scaling one argument by a power of two
// scales the result by the same factor, exactly.  Cross skips its internal
// scaling for arguments of ordinary magnitude, so the two paths through the
// function have to agree bit for bit; the exponents here straddle the point
// where it switches over.
func TestCrossScaleExact(t *testing.T) {
	vals := []float64{0, 1, -1, 0.1, 3, 7.5, 1e8, 1e-8}
	for _, a := range vals {
		for _, b := range vals {
			for _, c := range vals {
				for _, d := range vals {
					base := Vec2{a, b}.Cross(Vec2{c, d})
					for _, e := range []int{-900, -600, -501, -499, 0, 499, 501, 600, 900} {
						want := math.Ldexp(base, e)
						if math.IsInf(want, 0) || (want == 0 && base != 0) {
							continue // the scaled result is not representable
						}
						v := Vec2{math.Ldexp(a, e), math.Ldexp(b, e)}
						if got := v.Cross(Vec2{c, d}); got != want {
							t.Fatalf("Cross(%v, %v) = %v, want %v (2**%d times Cross(%v, %v))",
								v, Vec2{c, d}, got, want, e, Vec2{a, b}, Vec2{c, d})
						}
					}
				}
			}
		}
	}
}

func TestFrexp(t *testing.T) {
	inf, nan := math.Inf(1), math.NaN()
	for _, v := range []Vec2{
		{0, 0}, {1, 0}, {0, -1}, {3, 4}, {-0.1, 0.2},
		{1e-320, 0}, {5e-324, 5e-324}, {1e300, 1e300}, {1e300, 1e-5},
	} {
		frac, exp := v.Frexp()
		// the split must be exact for components of comparable magnitude
		back := Vec2{X: math.Ldexp(frac.X, exp), Y: math.Ldexp(frac.Y, exp)}
		if back != v {
			t.Errorf("Frexp(%v) = %v, %d, which scales back to %v", v, frac, exp, back)
		}
		larger := max(math.Abs(frac.X), math.Abs(frac.Y))
		if v != (Vec2{}) && !(larger >= 0.5 && larger < 1) {
			t.Errorf("Frexp(%v) = %v, larger component %v is not in [0.5, 1)", v, frac, larger)
		}
	}

	// a component more than 2**1074 times smaller than the other is lost,
	// which is why the scaling in Cross and Inv is a fallback and not the
	// normal path
	if frac, _ := (Vec2{1e300, 1e-300}).Frexp(); frac.Y != 0 {
		t.Errorf("Frexp({1e300, 1e-300}) = %v, expected the small component to underflow", frac)
	}

	// a vector with no meaningful scale is left alone
	for _, v := range []Vec2{{inf, 0}, {1, nan}, {0, 0}} {
		if frac, exp := v.Frexp(); exp != 0 || frac != v && !math.IsNaN(v.Y) {
			t.Errorf("Frexp(%v) = %v, %d, want the vector unchanged with exponent 0", v, frac, exp)
		}
	}
}

// TestDotExtremeScale checks vectors whose products lie outside the range of
// float64, where forming v.X*w.X+v.Y*w.Y directly gives infinity or NaN even
// though the result itself is representable.
func TestDotExtremeScale(t *testing.T) {
	for _, tc := range []struct {
		name string
		v, w Vec2
		want float64
	}{
		// each product overflows on its own, but they cancel exactly
		{"perpendicular", Vec2{1e200, 1e200}, Vec2{1e200, -1e200}, 0},
		{"antiparallel", Vec2{3e200, 4e200}, Vec2{4e200, -3e200}, 0},
		// one component huge, the other tiny
		{"mixed scales", Vec2{1e300, 0}, Vec2{1e-300, 0}, 1e300 * 1e-300},
		// both products underflow, and so does their sum
		{"small", Vec2{1e-200, 0}, Vec2{1e-200, 0}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.Dot(tc.w); got != tc.want {
				t.Errorf("Dot(%v, %v) = %v, want %v", tc.v, tc.w, got, tc.want)
			}
		})
	}

	// the answer rests on a component too small to survive scaling, so it has
	// to be reached without scaling
	if got := (Vec2{1e300, 1e-300}).Dot(Vec2{0, 1}); got != 1e-300 {
		t.Errorf("Dot({1e300, 1e-300}, {0, 1}) = %v, want 1e-300", got)
	}
}

// TestMiddleExtremeScale checks points whose coordinates sum to outside the
// range of float64, where the midpoint itself is comfortably inside it.
func TestMiddleExtremeScale(t *testing.T) {
	for _, tc := range []struct {
		a, b, want Vec2
	}{
		{Vec2{1e308, 1e308}, Vec2{1e308, 1e308}, Vec2{1e308, 1e308}},
		{Vec2{1e308, -1e308}, Vec2{1.5e308, -1.5e308}, Vec2{1.25e308, -1.25e308}},
		{Vec2{math.MaxFloat64, 0}, Vec2{math.MaxFloat64, 0}, Vec2{math.MaxFloat64, 0}},
		// the ordinary case must be untouched, and exact
		{Vec2{1, 2}, Vec2{3, 5}, Vec2{2, 3.5}},
	} {
		if got := Middle(tc.a, tc.b); got != tc.want {
			t.Errorf("Middle(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

// extremeSample returns a mantissa in [1, 2) with a random sign, at an
// exponent anywhere in the range of float64.  The exponent stops at 1022 so
// that the sample itself stays finite.
func extremeSample(rng *rand.Rand) float64 {
	m := rng.Float64() + 1
	if rng.Intn(2) == 0 {
		m = -m
	}
	return math.Ldexp(m, rng.Intn(2097)-1074)
}

// exactDot computes the dot product of v and other without rounding.
func exactDot(v, other Vec2) *big.Float {
	f := func(x float64) *big.Float { return new(big.Float).SetPrec(300).SetFloat64(x) }
	p := new(big.Float).SetPrec(300).Mul(f(v.X), f(other.X))
	q := new(big.Float).SetPrec(300).Mul(f(v.Y), f(other.Y))
	return new(big.Float).SetPrec(300).Add(p, q)
}

// exactCross computes the cross product of v and other without rounding.
// Products of two float64 values are exact at 300 bits of precision, and so is
// their difference.
func exactCross(v, other Vec2) *big.Float {
	f := func(x float64) *big.Float { return new(big.Float).SetPrec(300).SetFloat64(x) }
	p := new(big.Float).SetPrec(300).Mul(f(v.X), f(other.Y))
	q := new(big.Float).SetPrec(300).Mul(f(v.Y), f(other.X))
	return new(big.Float).SetPrec(300).Sub(p, q)
}

// TestCrossAgainstExact checks Cross against exact arithmetic, over vectors
// whose components are spread across the whole exponent range of float64.
//
// This is what backs the scaling fallback.  Scaling drops a component far
// smaller than the other, so it can only be taken where that component cannot
// carry the answer; the argument is that a product has to have overflowed to
// get there, which bounds how far apart the components can be.  An argument is
// not a proof, and only sampling every scale tests it.
func TestCrossAgainstExact(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	scaled := 0
	for range 100000 {
		v, w := Vec2{extremeSample(rng), extremeSample(rng)},
			Vec2{extremeSample(rng), extremeSample(rng)}
		if r := cross(v, w); math.Abs(r) > math.MaxFloat64 {
			scaled++
		}

		want := exactCross(v, w)
		wantF, _ := want.Float64() // the exact value, correctly rounded
		if math.IsInf(wantF, 0) {
			continue // the area itself is outside the range of float64
		}

		got := v.Cross(w)
		if got == wantF {
			continue
		}

		// The compensated products are accurate to within 1.5 ulp of the
		// exact value, whether or not the scaling was used.
		ulp := math.Abs(math.Nextafter(wantF, math.Inf(1)) - wantF)
		if ulp == 0 {
			ulp = math.SmallestNonzeroFloat64
		}
		diff := new(big.Float).SetPrec(300).Sub(want, new(big.Float).SetPrec(300).SetFloat64(got))
		tol := new(big.Float).SetPrec(300).SetFloat64(1.5 * ulp)
		if diff.Abs(diff).Cmp(tol) > 0 {
			t.Fatalf("Cross(%v, %v) = %v, exact value %v", v, w, got, want.Text('g', 20))
		}
	}

	// The fallback is the point of the test, so fail if the samples stopped
	// reaching it.
	if scaled < 1000 {
		t.Errorf("the scaling fallback ran %d times, too few to test it", scaled)
	}
}

// TestDotAgainstExact checks Dot against exact arithmetic, over vectors whose
// components are spread across the whole exponent range of float64.  As for
// Cross, this is what backs the scaling fallback: the argument that the
// component it drops cannot carry the answer holds only where a product has
// overflowed, and only sampling every scale tests it.
func TestDotAgainstExact(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	scaled := 0
	for range 100000 {
		v, w := Vec2{extremeSample(rng), extremeSample(rng)},
			Vec2{extremeSample(rng), extremeSample(rng)}
		if r := dot(v, w); math.Abs(r) > math.MaxFloat64 {
			scaled++
		}

		want := exactDot(v, w)
		wantF, _ := want.Float64() // the exact value, correctly rounded
		if math.IsInf(wantF, 0) {
			continue // the result itself is outside the range of float64
		}

		got := v.Dot(w)
		if got == wantF {
			continue
		}

		// The compensated products are accurate to within 1.5 ulp of the
		// exact value, whether or not the scaling was used.
		ulp := math.Abs(math.Nextafter(wantF, math.Inf(1)) - wantF)
		if ulp == 0 {
			ulp = math.SmallestNonzeroFloat64
		}
		diff := new(big.Float).SetPrec(300).Sub(want, new(big.Float).SetPrec(300).SetFloat64(got))
		tol := new(big.Float).SetPrec(300).SetFloat64(1.5 * ulp)
		if diff.Abs(diff).Cmp(tol) > 0 {
			t.Fatalf("Dot(%v, %v) = %v, exact value %v", v, w, got, want.Text('g', 20))
		}
	}

	// The fallback is the point of the test, so fail if the samples stopped
	// reaching it.
	if scaled < 1000 {
		t.Errorf("the scaling fallback ran %d times, too few to test it", scaled)
	}
}

// TestDotScaleExact checks that scaling one argument by a power of two scales
// the result by the same factor, exactly.  Dot skips its internal scaling for
// arguments of ordinary magnitude, so the two paths have to agree bit for bit.
func TestDotScaleExact(t *testing.T) {
	vals := []float64{0, 1, -1, 0.1, 3, 7.5, 1e8, 1e-8}
	for _, a := range vals {
		for _, b := range vals {
			for _, c := range vals {
				for _, d := range vals {
					base := Vec2{a, b}.Dot(Vec2{c, d})
					for _, e := range []int{-900, -600, -501, -499, 0, 499, 501, 600, 900} {
						want := math.Ldexp(base, e)
						if math.IsInf(want, 0) || (want == 0 && base != 0) {
							continue // the scaled result is not representable
						}
						v := Vec2{math.Ldexp(a, e), math.Ldexp(b, e)}
						if got := v.Dot(Vec2{c, d}); got != want {
							t.Fatalf("Dot(%v, %v) = %v, want %v (2**%d times Dot(%v, %v))",
								v, Vec2{c, d}, got, want, e, Vec2{a, b}, Vec2{c, d})
						}
					}
				}
			}
		}
	}
}

// TestCrossAntisymmetric checks v x w against -(w x v).  Swapping the
// arguments changes which of the two products is corrected for its rounding
// error, so the results need not be bit-identical, but the sign must always
// agree and the magnitudes must agree to within a rounding error.
func TestCrossAntisymmetric(t *testing.T) {
	vals := []float64{0, 1, -1, 0.1, 1e8, 1e-8, 3, 7.5}
	for _, a := range vals {
		for _, b := range vals {
			for _, c := range vals {
				for _, d := range vals {
					v, w := Vec2{a, b}, Vec2{c, d}
					fwd, rev := v.Cross(w), w.Cross(v)
					if (fwd > 0) != (rev < 0) || (fwd < 0) != (rev > 0) {
						t.Fatalf("Cross(%v, %v) = %v, Cross(%v, %v) = %v, signs disagree",
							v, w, fwd, w, v, rev)
					}
					if math.Abs(fwd+rev) > 4e-16*math.Abs(fwd) {
						t.Fatalf("Cross(%v, %v) = %v, but Cross(%v, %v) = %v",
							v, w, fwd, w, v, rev)
					}
				}
			}
		}
	}
}

// TestLengthExtremeScale checks vectors whose components square to outside the
// range of float64, where sqrt(x*x+y*y) underflows to zero or overflows to
// infinity but the length itself is perfectly representable.  The scales
// between 1e-154 and 1e-162 are the ones where the squares are subnormal but
// not yet zero, so that the sum keeps a plausible magnitude while losing most
// of its significant digits.
func TestLengthExtremeScale(t *testing.T) {
	for _, s := range []float64{1e-320, 5e-324, 1e-200, 1e-165, 1e-162,
		1e-160, 1e-155, 1e-154, 1e-150, 1, 1e200, 1e308} {
		v := Vec2{3 * s, 4 * s}
		want := 5 * s
		got := v.Length()
		if math.Abs(got-want) > 1e-12*want {
			t.Errorf("Vec2{3e, 4e}.Length() for e = %g is %g, want %g", s, got, want)
		}
	}
	if got := (Vec2{0, 0}).Length(); got != 0 {
		t.Errorf("zero vector has length %g", got)
	}
}
