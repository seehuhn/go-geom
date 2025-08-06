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

func TestNormalVerySmallVector(t *testing.T) {
	v := Vec2{1e-12, 1e-12}
	got := v.Normal()
	want := Vec2{0, 0}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Normal() for very small vector mismatch (-want +got):\n%s", diff)
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
	for range b.N {
		_ = v.Length()
	}
}

func BenchmarkAdd(b *testing.B) {
	v1 := Vec2{1, 2}
	v2 := Vec2{3, 4}
	for range b.N {
		_ = v1.Add(v2)
	}
}

func BenchmarkDot(b *testing.B) {
	v1 := Vec2{1, 2}
	v2 := Vec2{3, 4}
	for range b.N {
		_ = v1.Dot(v2)
	}
}

func BenchmarkNormal(b *testing.B) {
	v := Vec2{3, 4}
	for range b.N {
		_ = v.Normal()
	}
}
