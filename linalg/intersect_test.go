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
			// the expected point is the zero vector on the failing cases,
			// which is what Intersection returns when it has no answer
			if intersection != tc.expected {
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

// TestIntersectionScaleInvariant checks that the same configuration of lines
// gives the same answer whatever units the points are expressed in.  A test on
// the cross product alone is not scale invariant: it calls perpendicular
// millimetre-sized segments parallel.
func TestIntersectionScaleInvariant(t *testing.T) {
	for _, s := range []float64{1e300, 1e150, 1e6, 1e3, 1, 1e-3, 1e-6, 1e-9,
		1e-12, 1e-150, 1e-300, 1e-320} {
		a1 := vec.Vec2{X: -s, Y: 0}
		a2 := vec.Vec2{X: s, Y: 0}
		b1 := vec.Vec2{X: 0, Y: -s}
		b2 := vec.Vec2{X: 0, Y: s}

		p, ok := Intersection(a1, a2, b1, b2)
		if !ok {
			t.Errorf("scale %g: perpendicular lines reported as parallel", s)
			continue
		}
		if math.Abs(p.X) > 1e-12*s || math.Abs(p.Y) > 1e-12*s {
			t.Errorf("scale %g: intersection = %v, want the origin", s, p)
		}
	}
}

// TestIntersectionNearlyParallel checks that the rejection threshold is on the
// angle between the lines, independent of scale.
func TestIntersectionNearlyParallel(t *testing.T) {
	for _, s := range []float64{1e300, 1e6, 1, 1e-6, 1e-300} {
		for _, sin := range []float64{1e-6, 1e-10} {
			a1 := vec.Vec2{X: 0, Y: 0}
			a2 := vec.Vec2{X: s, Y: 0}
			b1 := vec.Vec2{X: 0, Y: s}
			b2 := vec.Vec2{X: s, Y: s + s*sin}

			_, ok := Intersection(a1, a2, b1, b2)
			if want := sin > parallelTol; ok != want {
				t.Errorf("scale %g, sin %g: ok = %v, want %v", s, sin, ok, want)
			}
		}
	}
}

func TestIntersectionDegenerate(t *testing.T) {
	o := vec.Vec2{X: 3, Y: 4}
	for _, tc := range []struct{ a1, a2, b1, b2 vec.Vec2 }{
		{o, o, vec.Vec2{X: 0, Y: 0}, vec.Vec2{X: 1, Y: 1}}, // first is a point
		{vec.Vec2{X: 0, Y: 0}, vec.Vec2{X: 1, Y: 1}, o, o}, // second is a point
		{o, o, o, o}, // both are points
	} {
		p, ok := Intersection(tc.a1, tc.a2, tc.b1, tc.b2)
		if ok {
			t.Errorf("Intersection(%v, %v, %v, %v) reported a crossing",
				tc.a1, tc.a2, tc.b1, tc.b2)
		}
		if p != (vec.Vec2{}) {
			t.Errorf("Intersection(%v, %v, %v, %v) returned %v, want the zero vector",
				tc.a1, tc.a2, tc.b1, tc.b2, p)
		}
	}
}

// TestMiterScaleInvariant checks that a right-angle corner produces a miter
// point whatever units it is expressed in.  Comparing the segment lengths
// against an absolute bound rejects small corners outright.
func TestMiterScaleInvariant(t *testing.T) {
	for _, s := range []float64{1e300, 1e150, 1e6, 1e3, 1, 1e-3, 1e-6, 1e-9,
		1e-12, 1e-100, 1e-300} {
		a := vec.Vec2{X: -s, Y: 0}
		b := vec.Vec2{X: 0, Y: 0}
		c := vec.Vec2{X: 0, Y: -s}

		p, ok := Miter(a, b, c, s/10, true)
		if !ok {
			t.Errorf("scale %g: right-angle corner has no miter point", s)
			continue
		}
		// the miter of a right angle sits at half the line width along each
		// axis from the vertex
		want := s / 20
		if math.Abs(math.Abs(p.X)-want) > 1e-9*s || math.Abs(math.Abs(p.Y)-want) > 1e-9*s {
			t.Errorf("scale %g: miter = %v, want distance %g from the vertex", s, p, want)
		}
	}
}

// TestMiterDegenerate checks that a corner whose arms are too short to have a
// direction is rejected, at any scale.
func TestMiterDegenerate(t *testing.T) {
	// a coincident pair has no direction whatever the scale, down into the
	// subnormal range where the coordinates have lost most of their own
	// precision
	for _, s := range []float64{1e300, 1e6, 1, 1e-6, 1e-300, 1e-315} {
		b := vec.Vec2{X: s, Y: s}
		if _, ok := Miter(b, b, vec.Vec2{X: s, Y: 0}, s/10, true); ok {
			t.Errorf("scale %g: coincident points accepted", s)
		}
	}

	// An arm below the tolerance but still non-zero.  This stops at 1e-300:
	// further down the gap between representable coordinates is itself a
	// larger fraction of them than the tolerance, so no such arm exists.
	for _, s := range []float64{1e300, 1e6, 1, 1e-6, 1e-300} {
		b := vec.Vec2{X: s, Y: s}
		tiny := s * 1e-13
		a := vec.Vec2{X: s + tiny, Y: s}
		c := vec.Vec2{X: s, Y: 2 * s}
		if arm := a.Sub(b).Length(); !(arm > 0) {
			t.Fatalf("scale %g: the short arm rounded away, the test is no longer meaningful", s)
		}
		if _, ok := Miter(a, b, c, s/10, true); ok {
			t.Errorf("scale %g: arm of length %g accepted", s, tiny)
		}
	}

	// all three points at the origin, the one case with no scale to measure
	// the arms against
	o := vec.Vec2{X: 0, Y: 0}
	if _, ok := Miter(o, o, o, 1, true); ok {
		t.Error("a corner at the origin with no extent was accepted")
	}
}

// TestMiterNearlyStraight checks the fallback for a corner which is too
// straight for its outer edges to meet, but not exactly straight.  Both
// segments then put the offset on the same side, and the fallback belongs on
// the offset edge, half a line width from the centre line.  Only for an
// exactly straight corner do the two sides disagree and the fallback collapse
// onto b.  Callers of Miter which ignore ok use this point as an outline
// vertex, so moving it onto the centre line would pinch the outline.
func TestMiterNearlyStraight(t *testing.T) {
	const lineWidth = 2.0
	a := vec.Vec2{X: -1, Y: 0}
	b := vec.Vec2{X: 0, Y: 0}

	for _, sin := range []float64{1e-12, 1e-10, 1e-9} {
		c := vec.Vec2{X: 1, Y: sin}
		p, ok := Miter(a, b, c, lineWidth, true)
		if ok {
			t.Errorf("sin %g: a corner this straight should have no miter point", sin)
			continue
		}
		if got := math.Abs(p.Y); math.Abs(got-lineWidth/2) > 1e-9 {
			t.Errorf("sin %g: fallback %v is %g from the centre line, want %g",
				sin, p, got, lineWidth/2)
		}
	}

	// exactly straight: the two sides disagree and the fallback is b
	c := vec.Vec2{X: 1, Y: 0}
	if p, ok := Miter(a, b, c, lineWidth, true); ok || p != b {
		t.Errorf("straight corner gave %v, ok = %v, want %v and false", p, ok, b)
	}
}

// TestMiterAwayFromOrigin checks the property that makes the degeneracy bound
// relative: a corner is judged against the precision its own coordinates are
// held to, so one and the same shape is usable near the origin and degenerate
// far from it.
func TestMiterAwayFromOrigin(t *testing.T) {
	// long beside coordinates of its own size, and 800 times the gap between
	// representable coordinates at 1e12, so the corner survives being moved
	// out there
	const arm = 0.1
	a := vec.Vec2{X: -arm, Y: 0}
	b := vec.Vec2{X: 0, Y: 0}
	c := vec.Vec2{X: 0, Y: -arm}

	if _, ok := Miter(a, b, c, arm/10, true); !ok {
		t.Errorf("corner with arms of %g at the origin was rejected", arm)
	}

	const far = 1e12
	off := vec.Vec2{X: far, Y: far}
	fa, fb, fc := a.Add(off), b.Add(off), c.Add(off)
	// the arms are rounded to a multiple of the coordinate spacing out here,
	// but they must not collapse: the rejection has to come from the bound
	// and not from a coincident pair
	if got := fa.Sub(fb).Length(); math.Abs(got-arm) > arm/100 {
		t.Fatalf("the moved corner has arms of %g, not about %g", got, arm)
	}
	if _, ok := Miter(fa, fb, fc, arm/10, true); ok {
		t.Errorf("corner with arms of %g at %g was accepted", arm, far)
	}
}
