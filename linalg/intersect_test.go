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
	"math/rand"
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

func TestSegmentIntersection(t *testing.T) {
	type testCase struct {
		name           string
		a1, a2, b1, b2 vec.Vec2
		p              vec.Vec2
		tt, ss         float64
		ok             bool
	}
	testCases := []testCase{
		{
			name: "crossing at the middle",
			a1:   vec.Vec2{X: -1, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: -1}, b2: vec.Vec2{X: 0, Y: 1},
			p: vec.Vec2{X: 0, Y: 0}, tt: 0.5, ss: 0.5, ok: true,
		},
		{
			name: "crossing at a quarter",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 4, Y: 0},
			b1: vec.Vec2{X: 1, Y: -1}, b2: vec.Vec2{X: 1, Y: 3},
			p: vec.Vec2{X: 1, Y: 0}, tt: 0.25, ss: 0.25, ok: true,
		},
		{
			name: "touching at the start of the first segment",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: -1}, b2: vec.Vec2{X: 0, Y: 1},
			p: vec.Vec2{X: 0, Y: 0}, tt: 0, ss: 0.5, ok: true,
		},
		{
			name: "touching at the end of the first segment",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 1, Y: -1}, b2: vec.Vec2{X: 1, Y: 1},
			p: vec.Vec2{X: 1, Y: 0}, tt: 1, ss: 0.5, ok: true,
		},
		{
			name: "touching at the end of the second segment",
			a1:   vec.Vec2{X: -1, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: -1}, b2: vec.Vec2{X: 0, Y: 0},
			p: vec.Vec2{X: 0, Y: 0}, tt: 0.5, ss: 1, ok: true,
		},
		{
			name: "touching at the start of the second segment",
			a1:   vec.Vec2{X: -1, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: 0}, b2: vec.Vec2{X: 0, Y: 1},
			p: vec.Vec2{X: 0, Y: 0}, tt: 0.5, ss: 0, ok: true,
		},
		{
			// the first segment starts on the second, where the cross
			// product which gives its parameter can come out as a
			// negative zero
			name: "starting on the second segment",
			a1:   vec.Vec2{X: 0.1, Y: 0}, a2: vec.Vec2{X: 0.1, Y: 1},
			b1: vec.Vec2{X: 0, Y: 0}, b2: vec.Vec2{X: 1, Y: 0},
			p: vec.Vec2{X: 0.1, Y: 0}, tt: 0, ss: 0.1, ok: true,
		},
		{
			name: "lines cross beyond the end of the first segment",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 2, Y: -1}, b2: vec.Vec2{X: 2, Y: 1},
		},
		{
			name: "lines cross before the start of the first segment",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: -2, Y: -1}, b2: vec.Vec2{X: -2, Y: 1},
		},
		{
			name: "lines cross beyond the end of the second segment",
			a1:   vec.Vec2{X: -1, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: 1}, b2: vec.Vec2{X: 0, Y: 2},
		},
		{
			name: "parallel segments",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 1, Y: 0},
			b1: vec.Vec2{X: 0, Y: 1}, b2: vec.Vec2{X: 1, Y: 1},
		},
		{
			name: "collinear overlapping segments",
			a1:   vec.Vec2{X: 0, Y: 0}, a2: vec.Vec2{X: 2, Y: 0},
			b1: vec.Vec2{X: 1, Y: 0}, b2: vec.Vec2{X: 3, Y: 0},
		},
		{
			name: "first segment degenerate",
			a1:   vec.Vec2{X: 1, Y: 1}, a2: vec.Vec2{X: 1, Y: 1},
			b1: vec.Vec2{X: 0, Y: 1}, b2: vec.Vec2{X: 2, Y: 1},
		},
		{
			name: "second segment degenerate",
			a1:   vec.Vec2{X: 0, Y: 1}, a2: vec.Vec2{X: 2, Y: 1},
			b1: vec.Vec2{X: 1, Y: 1}, b2: vec.Vec2{X: 1, Y: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p, tParam, sParam, ok := SegmentIntersection(tc.a1, tc.a2, tc.b1, tc.b2)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if !ok {
				if p != (vec.Vec2{}) || tParam != 0 || sParam != 0 {
					t.Errorf("failed call returned %v, %g, %g", p, tParam, sParam)
				}
				return
			}
			if math.Abs(tParam-tc.tt) > 1e-12 || math.Abs(sParam-tc.ss) > 1e-12 {
				t.Errorf("t, s = %g, %g, want %g, %g", tParam, sParam, tc.tt, tc.ss)
			}
			// a parameter of zero must not carry a sign, so that a caller
			// which divides by it or tests its sign sees the same thing
			// whichever way the cross product rounded
			if math.Signbit(tParam) || math.Signbit(sParam) {
				t.Errorf("t, s = %g, %g, want no sign on a zero", tParam, sParam)
			}
			if math.Abs(p.X-tc.p.X) > 1e-12 || math.Abs(p.Y-tc.p.Y) > 1e-12 {
				t.Errorf("point = %v, want %v", p, tc.p)
			}
			// the point must lie on the second segment too
			q := tc.b1.Add(tc.b2.Sub(tc.b1).Mul(sParam))
			if math.Abs(p.X-q.X) > 1e-12 || math.Abs(p.Y-q.Y) > 1e-12 {
				t.Errorf("point = %v, but s puts it at %v", p, q)
			}
		})
	}
}

// TestSegmentIntersectionScale checks that the crossing is still found when
// the segments are far from the origin and short compared to their distance
// from it, where forming the products from the raw coordinates would lose
// every significant digit.
func TestSegmentIntersectionScale(t *testing.T) {
	for _, scale := range []float64{1, 1e6, 1e12, 1e100} {
		base := vec.Vec2{X: scale, Y: scale}
		// The segments are short compared to the distance from the origin,
		// but not so short that their endpoints round onto each other.
		d := scale * 1e-6
		a1 := base.Add(vec.Vec2{X: -d, Y: 0})
		a2 := base.Add(vec.Vec2{X: d, Y: 0})
		b1 := base.Add(vec.Vec2{X: 0, Y: -d})
		b2 := base.Add(vec.Vec2{X: 0, Y: d})

		p, tParam, sParam, ok := SegmentIntersection(a1, a2, b1, b2)
		if !ok {
			t.Errorf("scale %g: no intersection found", scale)
			continue
		}
		// The endpoints are only held to a relative precision, so the
		// parameters can be no more accurate than that precision measured
		// against the length of the segments.
		if math.Abs(tParam-0.5) > 1e-8 || math.Abs(sParam-0.5) > 1e-8 {
			t.Errorf("scale %g: t, s = %g, %g, want 0.5, 0.5", scale, tParam, sParam)
		}
		tol := 1e-12 * scale
		if math.Abs(p.X-base.X) > tol || math.Abs(p.Y-base.Y) > tol {
			t.Errorf("scale %g: point = %v, want %v", scale, p, base)
		}
	}
}

// TestSegmentIntersectionSharedEndpoint checks that a segment which ends
// exactly on the end of another is always found to cross it, and that the
// crossing reported is that endpoint itself.  The parameter of such a
// crossing rounds to either side of 1, so an exact range test would reject
// about a third of these, and an exact comparison would leave others a
// rounding error short of the endpoint.
func TestSegmentIntersectionSharedEndpoint(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	random := func() vec.Vec2 {
		return vec.Vec2{X: rng.Float64()*10 - 5, Y: rng.Float64()*10 - 5}
	}
	for i := range 10000 {
		a1, a2, b1 := random(), random(), random()
		p, tParam, sParam, ok := SegmentIntersection(a1, a2, b1, a2)
		if !ok {
			t.Errorf("trial %d: shared endpoint %v not found", i, a2)
			continue
		}
		if p != a2 || tParam != 1 || sParam != 1 {
			t.Errorf("trial %d: got %v, %g, %g, want %v, 1, 1", i, p, tParam, sParam, a2)
		}

		// the same, with the shared endpoint at the start of both
		p, tParam, sParam, ok = SegmentIntersection(a2, a1, a2, b1)
		if !ok {
			t.Errorf("trial %d: shared start %v not found", i, a2)
			continue
		}
		if p != a2 || tParam != 0 || sParam != 0 {
			t.Errorf("trial %d: got %v, %g, %g, want %v, 0, 0", i, p, tParam, sParam, a2)
		}
	}
}

// TestSegmentIntersectionNearMiss checks that a crossing which falls beyond
// an end by more than the slack is still rejected, while one within the slack
// on either side is snapped to the end.
func TestSegmentIntersectionNearMiss(t *testing.T) {
	a1 := vec.Vec2{X: 0, Y: 0}
	a2 := vec.Vec2{X: 1, Y: 0}
	b1 := vec.Vec2{X: 1, Y: -1}

	// the second segment stops short of the first by 10 times the slack
	b2 := vec.Vec2{X: 1, Y: -10 * endTol}
	if _, _, _, ok := SegmentIntersection(a1, a2, b1, b2); ok {
		t.Error("a crossing beyond the slack was accepted")
	}

	// and by a tenth of it
	b2 = vec.Vec2{X: 1, Y: -endTol / 10}
	p, tParam, sParam, ok := SegmentIntersection(a1, a2, b1, b2)
	if !ok {
		t.Fatal("a crossing within the slack was rejected")
	}
	if p != a2 || tParam != 1 || sParam != 1 {
		t.Errorf("got %v, %g, %g, want %v, 1, 1", p, tParam, sParam, a2)
	}

	// crossing a tenth of the slack inside the end snaps to it as well
	b2 = vec.Vec2{X: 1, Y: endTol / 10}
	p, tParam, sParam, ok = SegmentIntersection(a1, a2, b1, b2)
	if !ok {
		t.Fatal("a crossing inside the end was rejected")
	}
	if p != a2 || tParam != 1 || sParam != 1 {
		t.Errorf("got %v, %g, %g, want %v, 1, 1", p, tParam, sParam, a2)
	}
}

// TestSegmentIntersectionSymmetry checks that swapping the two segments finds
// the same crossing point.
func TestSegmentIntersectionSymmetry(t *testing.T) {
	a1 := vec.Vec2{X: -3, Y: -1}
	a2 := vec.Vec2{X: 5, Y: 7}
	b1 := vec.Vec2{X: -2, Y: 6}
	b2 := vec.Vec2{X: 4, Y: -2}

	p1, t1, s1, ok1 := SegmentIntersection(a1, a2, b1, b2)
	p2, t2, s2, ok2 := SegmentIntersection(b1, b2, a1, a2)
	if !ok1 || !ok2 {
		t.Fatalf("ok = %v, %v, want true, true", ok1, ok2)
	}
	if math.Abs(p1.X-p2.X) > 1e-12 || math.Abs(p1.Y-p2.Y) > 1e-12 {
		t.Errorf("points differ: %v vs %v", p1, p2)
	}
	// swapping the segments swaps the roles of the two parameters
	if math.Abs(t1-s2) > 1e-12 || math.Abs(s1-t2) > 1e-12 {
		t.Errorf("parameters %g, %g do not mirror %g, %g", t1, s1, t2, s2)
	}
}

func TestMiterLength(t *testing.T) {
	type testCase struct {
		name       string
		in, out    vec.Vec2
		miterLimit float64
		want       float64
		ok         bool
	}
	testCases := []testCase{
		{
			// a straight join reaches exactly half the line width
			name: "straight",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: 1, Y: 0},
			miterLimit: 10, want: 1, ok: true,
		},
		{
			// a right angle turn: miter ratio sqrt(2)
			name: "right angle",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: 0, Y: 1},
			miterLimit: 10, want: math.Sqrt2, ok: true,
		},
		{
			// 60 degrees of turn: miter ratio 2/sqrt(3)
			name: "sixty degree turn",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: 0.5, Y: math.Sqrt(3) / 2},
			miterLimit: 10, want: 2 / math.Sqrt(3), ok: true,
		},
		{
			name: "right angle below the limit",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: 0, Y: 1},
			miterLimit: 1.2, want: 1, ok: false,
		},
		{
			// the segments double back, so the miter runs off to infinity
			name: "reversal",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: -1, Y: 0},
			miterLimit: 1e6, want: 1, ok: false,
		},
		{
			name: "zero incoming direction",
			in:   vec.Vec2{}, out: vec.Vec2{X: 0, Y: 1},
			miterLimit: 10, want: 1, ok: false,
		},
		{
			name: "zero outgoing direction",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{},
			miterLimit: 10, want: 1, ok: false,
		},
		{
			// a limit below 1 is shorter than any miter
			name: "limit below one",
			in:   vec.Vec2{X: 1, Y: 0}, out: vec.Vec2{X: 1, Y: 0},
			miterLimit: 0.5, want: 1, ok: false,
		},
	}

	const lineWidth = 2
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := MiterLength(tc.in, tc.out, lineWidth, tc.miterLimit)
			if ok != tc.ok {
				t.Fatalf("ok = %v, want %v", ok, tc.ok)
			}
			if math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("length = %g, want %g", got, tc.want)
			}
		})
	}
}

// TestMiterLengthAtLimit checks that a corner sitting exactly on the miter
// limit is mitered, whichever side the rounding error of the ratio falls on.
func TestMiterLengthAtLimit(t *testing.T) {
	// A right angle has a miter ratio of exactly sqrt(2).
	in := vec.Vec2{X: 1, Y: 0}
	out := vec.Vec2{X: 0, Y: 1}

	if _, ok := MiterLength(in, out, 2, math.Sqrt2); !ok {
		t.Error("a corner exactly on the limit should be mitered")
	}
	// Just below the limit by more than the slack, it bevels.
	if _, ok := MiterLength(in, out, 2, math.Sqrt2*(1-1e-9)); ok {
		t.Error("a corner past the limit should be bevelled")
	}
}

// TestMiterLengthSharpCorner checks the length reported for corners which
// nearly double back.  Taking the half angle from the cosine of the turn
// leaves no significant digits there: the cosine is close to -1, and adding
// one to it cancels away everything the cosine had.
func TestMiterLengthSharpCorner(t *testing.T) {
	const lineWidth = 2
	in := vec.Vec2{X: 1, Y: 0}
	for _, turn := range []float64{1e-2, 1e-4, 1e-6, 1e-8} {
		// the direction turns by pi-turn, so the interior angle is turn and
		// the miter reaches lineWidth/2/sin(turn/2) from the corner
		out := vec.Vec2{X: -math.Cos(turn), Y: math.Sin(turn)}
		want := lineWidth / 2 / math.Sin(turn/2)

		got, ok := MiterLength(in, out, lineWidth, 1e12)
		if !ok {
			t.Errorf("turn %g: expected a miter", turn)
			continue
		}
		if math.Abs(got-want) > 1e-9*want {
			t.Errorf("turn %g: length = %g, want %g", turn, got, want)
		}

		// The limit is on the whole miter, twice the length returned, so
		// this corner sits exactly on a limit of 2*want/lineWidth.
		if _, ok := MiterLength(in, out, lineWidth, 2*want/lineWidth); !ok {
			t.Errorf("turn %g: a corner on the limit should be mitered", turn)
		}
	}
}

// TestMiterLengthScale checks that the result does not depend on how long the
// direction vectors are, nor on the scale they are expressed in.
func TestMiterLengthScale(t *testing.T) {
	want, ok := MiterLength(vec.Vec2{X: 1, Y: 0}, vec.Vec2{X: 0, Y: 1}, 2, 10)
	if !ok {
		t.Fatal("expected a miter")
	}
	for _, scale := range []float64{1e-150, 1e-6, 1, 1e6, 1e150} {
		in := vec.Vec2{X: scale, Y: 0}
		out := vec.Vec2{X: 0, Y: scale}
		got, ok := MiterLength(in, out, 2, 10)
		if !ok {
			t.Errorf("scale %g: expected a miter", scale)
			continue
		}
		if math.Abs(got-want) > 1e-12 {
			t.Errorf("scale %g: length = %g, want %g", scale, got, want)
		}
	}
}

// TestMiterLengthSymmetry checks that reversing the corner, so that the two
// segments swap roles, gives the same miter length.
func TestMiterLengthSymmetry(t *testing.T) {
	in := vec.Vec2{X: 3, Y: 1}
	out := vec.Vec2{X: -1, Y: 4}

	a, okA := MiterLength(in, out, 2, 10)
	b, okB := MiterLength(out.Neg(), in.Neg(), 2, 10)
	if okA != okB {
		t.Fatalf("ok = %v, %v", okA, okB)
	}
	if math.Abs(a-b) > 1e-12 {
		t.Errorf("lengths differ: %g vs %g", a, b)
	}
}
