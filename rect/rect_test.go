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

package rect

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestIsZero(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want bool
	}{
		{"zero rect", Rect{0, 0, 0, 0}, true},
		{"non-zero rect", Rect{1, 2, 3, 4}, false},
		{"partially zero", Rect{0, 0, 1, 1}, false},
		{"negative coords", Rect{-1, -1, -1, -1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.IsZero(); got != tt.want {
				t.Errorf("IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDxDy(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		dx   float64
		dy   float64
	}{
		{"unit square", Rect{0, 0, 1, 1}, 1, 1},
		{"rectangle", Rect{1, 2, 5, 7}, 4, 5},
		{"negative dimensions", Rect{5, 7, 1, 2}, -4, -5},
		{"zero width", Rect{3, 1, 3, 4}, 0, 3},
		{"zero height", Rect{1, 3, 4, 3}, 3, 0},
		{"point", Rect{2, 3, 2, 3}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Dx(); got != tt.dx {
				t.Errorf("Dx() = %v, want %v", got, tt.dx)
			}
			if got := tt.r.Dy(); got != tt.dy {
				t.Errorf("Dy() = %v, want %v", got, tt.dy)
			}
		})
	}
}

func TestCovers(t *testing.T) {
	tests := []struct {
		name string
		r1   Rect
		r2   Rect
		want bool
	}{
		{"identical rects", Rect{0, 0, 10, 10}, Rect{0, 0, 10, 10}, true},
		{"larger covers smaller", Rect{0, 0, 10, 10}, Rect{2, 2, 8, 8}, true},
		{"smaller doesn't cover larger", Rect{2, 2, 8, 8}, Rect{0, 0, 10, 10}, false},
		{"overlapping but not covering", Rect{0, 0, 5, 5}, Rect{3, 3, 8, 8}, false},
		{"adjacent rects", Rect{0, 0, 5, 5}, Rect{5, 0, 10, 5}, false},
		{"point in rect", Rect{0, 0, 10, 10}, Rect{5, 5, 5, 5}, true},
		{"zero rect covers zero rect", Rect{0, 0, 0, 0}, Rect{0, 0, 0, 0}, true},
		{"negative coords", Rect{-5, -5, 5, 5}, Rect{-2, -2, 2, 2}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r1.Covers(tt.r2); got != tt.want {
				t.Errorf("Covers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		x, y float64
		want Rect
	}{
		{"add to zero rect", Rect{0, 0, 0, 0}, 5, 3, Rect{0, 0, 5, 3}},
		{"point inside", Rect{0, 0, 10, 10}, 5, 5, Rect{0, 0, 10, 10}},
		{"expand left", Rect{5, 5, 10, 10}, 2, 7, Rect{2, 5, 10, 10}},
		{"expand right", Rect{0, 0, 5, 5}, 8, 3, Rect{0, 0, 8, 5}},
		{"expand bottom", Rect{0, 5, 10, 10}, 5, 2, Rect{0, 2, 10, 10}},
		{"expand top", Rect{0, 0, 10, 5}, 5, 8, Rect{0, 0, 10, 8}},
		{"expand all directions", Rect{5, 5, 10, 10}, 2, 12, Rect{2, 5, 10, 12}},
		{"negative coords", Rect{0, 0, 5, 5}, -3, -2, Rect{-3, -2, 5, 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.r // copy
			r.Add(tt.x, tt.y)
			if diff := cmp.Diff(tt.want, r); diff != "" {
				t.Errorf("Add() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExtend(t *testing.T) {
	tests := []struct {
		name string
		r1   Rect
		r2   Rect
		want Rect
	}{
		{"extend with zero rect", Rect{1, 2, 3, 4}, Rect{0, 0, 0, 0}, Rect{1, 2, 3, 4}},
		{"zero rect extend with non-zero", Rect{0, 0, 0, 0}, Rect{1, 2, 3, 4}, Rect{1, 2, 3, 4}},
		{"identical rects", Rect{1, 2, 3, 4}, Rect{1, 2, 3, 4}, Rect{1, 2, 3, 4}},
		{"overlapping rects", Rect{0, 0, 5, 5}, Rect{3, 3, 8, 8}, Rect{0, 0, 8, 8}},
		{"adjacent rects", Rect{0, 0, 5, 5}, Rect{5, 0, 10, 5}, Rect{0, 0, 10, 5}},
		{"separated rects", Rect{0, 0, 2, 2}, Rect{5, 5, 7, 7}, Rect{0, 0, 7, 7}},
		{"negative coords", Rect{-5, -5, 0, 0}, Rect{0, 0, 5, 5}, Rect{-5, -5, 5, 5}},
		{"r2 inside r1", Rect{0, 0, 10, 10}, Rect{2, 3, 7, 8}, Rect{0, 0, 10, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.r1 // copy
			r.Extend(tt.r2)
			if diff := cmp.Diff(tt.want, r); diff != "" {
				t.Errorf("Extend() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestScale(t *testing.T) {
	tests := []struct {
		name   string
		r      Rect
		factor float64
		want   Rect
	}{
		{"scale by 1", Rect{1, 2, 3, 4}, 1.0, Rect{1, 2, 3, 4}},
		{"scale by 2", Rect{1, 2, 3, 4}, 2.0, Rect{2, 4, 6, 8}},
		{"scale by 0.5", Rect{2, 4, 6, 8}, 0.5, Rect{1, 2, 3, 4}},
		{"scale by 0", Rect{1, 2, 3, 4}, 0.0, Rect{0, 0, 0, 0}},
		{"scale by -1", Rect{1, 2, 3, 4}, -1.0, Rect{-1, -2, -3, -4}},
		{"scale negative rect", Rect{-2, -4, -1, -2}, 2.0, Rect{-4, -8, -2, -4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.r // copy
			r.Scale(tt.factor)
			if diff := cmp.Diff(tt.want, r, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
				t.Errorf("Scale() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRounded(t *testing.T) {
	tests := []struct {
		name string
		r    Rect
		want Rect
	}{
		{"already integer", Rect{1, 2, 3, 4}, Rect{1, 2, 3, 4}},
		{"fractional coords", Rect{1.3, 2.7, 3.2, 4.8}, Rect{1, 2, 4, 5}},
		{"negative coords", Rect{-2.3, -1.7, -0.2, 0.8}, Rect{-3, -2, 0, 1}},
		{"zero rect", Rect{0, 0, 0, 0}, Rect{0, 0, 0, 0}},
		{"small values", Rect{0.1, 0.9, 1.1, 1.9}, Rect{0, 0, 2, 2}},
		{"exact half values", Rect{0.5, 1.5, 2.5, 3.5}, Rect{0, 1, 3, 4}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.r.Rounded()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Rounded() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRectProperties(t *testing.T) {
	testRects := []Rect{
		{0, 0, 0, 0},         // zero rect
		{1, 2, 3, 4},         // normal rect
		{-5, -3, -1, 2},      // mixed signs
		{0.1, 0.2, 0.3, 0.4}, // small values
		{100, 200, 300, 400}, // large values
	}

	t.Run("scale round-trip", func(t *testing.T) {
		for i, r := range testRects {
			t.Run(fmt.Sprintf("rect%d", i), func(t *testing.T) {
				if r.IsZero() {
					t.Skip("skip zero rect for scaling")
				}
				r2 := r
				r2.Scale(2.0)
				r2.Scale(0.5)
				if diff := cmp.Diff(r, r2, cmpopts.EquateApprox(1e-9, 1e-9)); diff != "" {
					t.Errorf("scale round-trip failed (-want +got):\n%s", diff)
				}
			})
		}
	})

	t.Run("extend is commutative", func(t *testing.T) {
		for i, r1 := range testRects {
			for j, r2 := range testRects {
				t.Run(fmt.Sprintf("rect%d_rect%d", i, j), func(t *testing.T) {
					a := r1
					a.Extend(r2)
					b := r2
					b.Extend(r1)

					if diff := cmp.Diff(a, b); diff != "" {
						t.Errorf("extend not commutative (-r1.extend(r2) +r2.extend(r1)):\n%s", diff)
					}
				})
			}
		}
	})

	t.Run("covers is reflexive", func(t *testing.T) {
		for i, r := range testRects {
			t.Run(fmt.Sprintf("rect%d", i), func(t *testing.T) {
				if !r.Covers(r) {
					t.Errorf("rect does not cover itself")
				}
			})
		}
	})

	t.Run("covers is transitive", func(t *testing.T) {
		r1 := Rect{0, 0, 10, 10}
		r2 := Rect{2, 2, 8, 8}
		r3 := Rect{4, 4, 6, 6}

		if !r1.Covers(r2) {
			t.Fatal("test setup: r1 should cover r2")
		}
		if !r2.Covers(r3) {
			t.Fatal("test setup: r2 should cover r3")
		}
		if !r1.Covers(r3) {
			t.Error("covers is not transitive: r1 covers r2 and r2 covers r3, but r1 does not cover r3")
		}
	})

	t.Run("rounded always expands or maintains", func(t *testing.T) {
		fractionalRects := []Rect{
			{1.1, 2.2, 3.3, 4.4},
			{-2.7, -1.3, 0.8, 1.9},
			{0.1, 0.1, 0.9, 0.9},
		}

		for i, r := range fractionalRects {
			t.Run(fmt.Sprintf("fractional%d", i), func(t *testing.T) {
				rounded := r.Rounded()
				if !rounded.Covers(r) {
					t.Errorf("rounded rect should cover original: rounded=%v, original=%v", rounded, r)
				}
			})
		}
	})
}

func BenchmarkAdd(b *testing.B) {
	r := Rect{1, 2, 3, 4}
	for i := range b.N {
		r.Add(float64(i), float64(i))
	}
}

func BenchmarkExtend(b *testing.B) {
	r := Rect{1, 2, 3, 4}
	other := Rect{5, 6, 7, 8}
	for range b.N {
		r.Extend(other)
	}
}

func BenchmarkRounded(b *testing.B) {
	r := Rect{1.23, 2.34, 3.45, 4.56}
	for range b.N {
		_ = r.Rounded()
	}
}
