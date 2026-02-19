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
	"fmt"
	"math"
	"testing"

	"seehuhn.de/go/geom/vec"
)

func TestIntersection(t *testing.T) {
	type testCase struct {
		a1, a2, b1, b2 vec.Vec2
		expected       vec.Vec2
		ok             bool
	}
	testCases := []testCase{
		{
			a1:       vec.Vec2{X: 0, Y: 0},
			a2:       vec.Vec2{X: 0, Y: 1},
			b1:       vec.Vec2{X: 0, Y: 0},
			b2:       vec.Vec2{X: 1, Y: 0},
			expected: vec.Vec2{X: 0, Y: 0},
			ok:       true,
		},
		{
			a1:       vec.Vec2{X: 0, Y: 0},
			a2:       vec.Vec2{X: 1, Y: 1},
			b1:       vec.Vec2{X: 0, Y: 0},
			b2:       vec.Vec2{X: -1, Y: 1},
			expected: vec.Vec2{X: 0, Y: 0},
			ok:       true,
		},
		{
			a1:       vec.Vec2{X: 0, Y: 1},
			a2:       vec.Vec2{X: 0, Y: 2},
			b1:       vec.Vec2{X: 0, Y: 0},
			b2:       vec.Vec2{X: 1, Y: 0},
			expected: vec.Vec2{X: 0, Y: 0},
			ok:       true,
		},
		{
			a1:       vec.Vec2{X: 1, Y: 1},
			a2:       vec.Vec2{X: 3, Y: 3},
			b1:       vec.Vec2{X: 0, Y: 2},
			b2:       vec.Vec2{X: 10, Y: 2},
			expected: vec.Vec2{X: 2, Y: 2},
			ok:       true,
		},

		{ // identical lines
			a1: vec.Vec2{X: 1, Y: 1},
			a2: vec.Vec2{X: 3, Y: 1},
			b1: vec.Vec2{X: 2, Y: 1},
			b2: vec.Vec2{X: 4, Y: 1},
			ok: false,
		},
		{ // parallel lines
			a1: vec.Vec2{X: -1, Y: -1},
			a2: vec.Vec2{X: -1, Y: 10},
			b1: vec.Vec2{X: 1, Y: 2},
			b2: vec.Vec2{X: 1, Y: 3},
			ok: false,
		},
		{ // one mis-specified line
			a1:       vec.Vec2{X: 1, Y: 1},
			a2:       vec.Vec2{X: 1, Y: 1},
			b1:       vec.Vec2{X: 0, Y: 0},
			b2:       vec.Vec2{X: -1, Y: 1},
			expected: vec.Vec2{X: 0, Y: 0},
			ok:       false,
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("Test%d", i), func(t *testing.T) {
			intersection, ok := Intersection(tc.a1, tc.a2, tc.b1, tc.b2)
			if ok != tc.ok {
				t.Errorf("expected ok=%v, got %v", tc.ok, ok)
			}
			if ok && intersection != tc.expected {
				t.Errorf("expected intersection=%v, got %v", tc.expected, intersection)
			}
		})
	}
}

func TestMiter(t *testing.T) {
	type testCase struct {
		name      string
		a, b, c   vec.Vec2
		lineWidth float64
		outer     bool
		expected  vec.Vec2
		ok        bool
	}

	testCases := []testCase{
		{
			name:      "90 degree turn, outer miter",
			a:         vec.Vec2{X: 0, Y: 0},
			b:         vec.Vec2{X: 1, Y: 0},
			c:         vec.Vec2{X: 1, Y: 1},
			lineWidth: 2.0,
			outer:     true,
			expected:  vec.Vec2{X: 2, Y: -1},
			ok:        true,
		},
		{
			name:      "90 degree turn, inner miter",
			a:         vec.Vec2{X: 0, Y: 0},
			b:         vec.Vec2{X: 1, Y: 0},
			c:         vec.Vec2{X: 1, Y: 1},
			lineWidth: 2.0,
			outer:     false,
			expected:  vec.Vec2{X: 0, Y: 1},
			ok:        true,
		},
		{
			name:      "your example: (10,0)--(0,0)--(0,10), outer miter",
			a:         vec.Vec2{X: 10, Y: 0},
			b:         vec.Vec2{X: 0, Y: 0},
			c:         vec.Vec2{X: 0, Y: 10},
			lineWidth: 2.0,
			outer:     true,
			expected:  vec.Vec2{X: -1, Y: -1},
			ok:        true,
		},
		{
			name:      "your example: (10,0)--(0,0)--(0,10), inner miter",
			a:         vec.Vec2{X: 10, Y: 0},
			b:         vec.Vec2{X: 0, Y: 0},
			c:         vec.Vec2{X: 0, Y: 10},
			lineWidth: 2.0,
			outer:     false,
			expected:  vec.Vec2{X: 1, Y: 1},
			ok:        true,
		},
		{
			name:      "straight line (parallel segments)",
			a:         vec.Vec2{X: 0, Y: 0},
			b:         vec.Vec2{X: 1, Y: 0},
			c:         vec.Vec2{X: 2, Y: 0},
			lineWidth: 2.0,
			outer:     true,
			expected:  vec.Vec2{X: 1, Y: 0}, // returns b when parallel
			ok:        false,
		},
		{
			name:      "degenerate case - same point a and b",
			a:         vec.Vec2{X: 1, Y: 1},
			b:         vec.Vec2{X: 1, Y: 1},
			c:         vec.Vec2{X: 2, Y: 2},
			lineWidth: 2.0,
			outer:     true,
			expected:  vec.Vec2{X: 1, Y: 1}, // returns b when degenerate
			ok:        false,
		},
		{
			name:      "degenerate case - same point b and c",
			a:         vec.Vec2{X: 0, Y: 0},
			b:         vec.Vec2{X: 1, Y: 1},
			c:         vec.Vec2{X: 1, Y: 1},
			lineWidth: 2.0,
			outer:     true,
			expected:  vec.Vec2{X: 1, Y: 1}, // returns b when degenerate
			ok:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			miter, ok := Miter(tc.a, tc.b, tc.c, tc.lineWidth, tc.outer)
			if ok != tc.ok {
				t.Errorf("expected ok=%v, got %v", tc.ok, ok)
			}
			if !ok {
				if miter != tc.b {
					t.Errorf("expected failed miter to return b=%v, got %v", tc.b, miter)
				}
			} else {
				const eps = 1e-9
				if math.Abs(miter.X-tc.expected.X) > eps || math.Abs(miter.Y-tc.expected.Y) > eps {
					t.Errorf("expected miter=%v, got %v", tc.expected, miter)
				}
			}
		})
	}
}

func TestIntersectionProperties(t *testing.T) {
	t.Run("intersection is on both lines", func(t *testing.T) {
		a1 := vec.Vec2{X: 0, Y: 0}
		a2 := vec.Vec2{X: 2, Y: 0}
		b1 := vec.Vec2{X: 1, Y: -1}
		b2 := vec.Vec2{X: 1, Y: 1}

		intersection, ok := Intersection(a1, a2, b1, b2)
		if !ok {
			t.Fatal("expected intersection to succeed")
		}

		expected := vec.Vec2{X: 1, Y: 0}
		if intersection != expected {
			t.Errorf("expected intersection=%v, got %v", expected, intersection)
		}
	})

	t.Run("intersection with negative coordinates", func(t *testing.T) {
		a1 := vec.Vec2{X: -2, Y: -2}
		a2 := vec.Vec2{X: 2, Y: 2}
		b1 := vec.Vec2{X: -2, Y: 2}
		b2 := vec.Vec2{X: 2, Y: -2}

		intersection, ok := Intersection(a1, a2, b1, b2)
		if !ok {
			t.Fatal("expected intersection to succeed")
		}

		expected := vec.Vec2{X: 0, Y: 0}
		if intersection != expected {
			t.Errorf("expected intersection=%v, got %v", expected, intersection)
		}
	})

	t.Run("nearly parallel lines", func(t *testing.T) {
		a1 := vec.Vec2{X: 0, Y: 0}
		a2 := vec.Vec2{X: 1, Y: 0}
		b1 := vec.Vec2{X: 0, Y: 1}
		b2 := vec.Vec2{X: 1, Y: 1.0000000001} // very slight angle

		_, ok := Intersection(a1, a2, b1, b2)
		if ok {
			t.Error("expected nearly parallel lines to be detected as parallel")
		}
	})
}

func TestMiterProperties(t *testing.T) {
	t.Run("miter with zero line width", func(t *testing.T) {
		a := vec.Vec2{X: 0, Y: 0}
		b := vec.Vec2{X: 1, Y: 0}
		c := vec.Vec2{X: 1, Y: 1}

		miter, ok := Miter(a, b, c, 0.0, true)
		if !ok {
			t.Error("expected miter to succeed with zero line width")
		}
		if miter != b {
			t.Errorf("expected miter at corner point %v, got %v", b, miter)
		}
	})

	t.Run("miter is symmetric for outer/inner", func(t *testing.T) {
		a := vec.Vec2{X: 0, Y: 0}
		b := vec.Vec2{X: 1, Y: 0}
		c := vec.Vec2{X: 1, Y: 1}
		lineWidth := 2.0

		outerMiter, outerOk := Miter(a, b, c, lineWidth, true)
		innerMiter, innerOk := Miter(a, b, c, lineWidth, false)

		if outerOk != innerOk {
			t.Error("outer and inner miter should have same success status")
		}

		if outerOk {
			outerDist := outerMiter.Sub(b).Length()
			innerDist := innerMiter.Sub(b).Length()

			const eps = 1e-9
			if math.Abs(outerDist-innerDist) > eps {
				t.Errorf("outer and inner miter should be equidistant from corner: outer=%v, inner=%v", outerDist, innerDist)
			}
		}
	})
}
