// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import (
	"math"
	"slices"

	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

// Cubic is a cubic Bezier curve, given by its start point, its two control
// points and its end point, in this order.  The curve is parametrised by t,
// which runs from 0 at the start point to 1 at the end point.
type Cubic [4]vec.Vec2

// At returns the point of the curve at the parameter t.  The curve is
// defined outside the range [0, 1] as well, and the method extrapolates
// there.
//
// The curve is evaluated by repeated linear interpolation, using de
// Casteljau's algorithm.  Every intermediate point is a weighted average of
// two others, with weights which are non-negative for t in [0, 1], so the
// result never leaves the convex hull of the four points.  It is exactly
// c[0] at t == 0 and exactly c[3] at t == 1.
func (c Cubic) At(t float64) vec.Vec2 {
	a := lerp(c[0], c[1], t)
	b := lerp(c[1], c[2], t)
	d := lerp(c[2], c[3], t)
	return lerp(lerp(a, b, t), lerp(b, d, t), t)
}

// Extrema returns the parameters in the open interval (0, 1) at which the
// x or the y coordinate of the curve has a local extremum, that is where
// its derivative changes sign.  The parameters are listed in increasing
// order, without repeats.  A parameter where the derivative only touches
// zero is not listed, and neither are the ends of the curve.
//
// The curve at each listed parameter is within rounding of its extreme
// value, for coordinates anywhere in the float64 range.
func (c Cubic) Extrema() []float64 {
	ts := c.appendExtrema(nil)
	slices.Sort(ts)
	return slices.Compact(ts)
}

// BBox returns the smallest axis-aligned rectangle which contains the curve
// for t in [0, 1].  This is exact up to rounding: the box may miss the
// curve by a few units in the last place where the curve touches it.
func (c Cubic) BBox() rect.Rect {
	bbox := rect.Rect{LLx: c[0].X, LLy: c[0].Y, URx: c[0].X, URy: c[0].Y}
	c.extend(&bbox)
	return bbox
}

// extend enlarges bbox to contain the curve, apart from its start point.
func (c Cubic) extend(bbox *rect.Rect) {
	bbox.Add(c[3].X, c[3].Y)
	var buf [4]float64
	for _, t := range c.appendExtrema(buf[:0]) {
		p := c.At(t)
		bbox.Add(p.X, p.Y)
	}
}

// appendExtrema appends the parameters of [Cubic.Extrema] to ts, unsorted
// and possibly with repeats.
func (c Cubic) appendExtrema(ts []float64) []float64 {
	ts = appendExtrema(ts, c[0].X, c[1].X, c[2].X, c[3].X)
	return appendExtrema(ts, c[0].Y, c[1].Y, c[2].Y, c[3].Y)
}

// appendExtrema appends the parameters in (0, 1) where the derivative of the
// one-dimensional cubic Bezier curve with coefficients a0, ..., a3 changes
// sign.
func appendExtrema(ts []float64, a0, a1, a2, a3 float64) []float64 {
	// Scaling by a power of two leaves the roots unchanged and introduces no
	// rounding short of underflow.  It keeps the differences below from
	// overflowing, and the products in the discriminant from overflowing or
	// underflowing.
	_, e := math.Frexp(max(math.Abs(a0), math.Abs(a1), math.Abs(a2), math.Abs(a3)))
	a0, a1, a2, a3 = math.Ldexp(a0, -e), math.Ldexp(a1, -e), math.Ldexp(a2, -e), math.Ldexp(a3, -e)

	// the derivative, divided by three, is A t^2 + B t + C
	d0, d1, d2 := a1-a0, a2-a1, a3-a2
	A := d0 - 2*d1 + d2
	B := 2 * (d1 - d0)
	C := d0

	_, e = math.Frexp(max(math.Abs(A), math.Abs(B), math.Abs(C)))
	A, B, C = math.Ldexp(A, -e), math.Ldexp(B, -e), math.Ldexp(C, -e)

	// A discriminant within the rounding error of zero belongs to a double
	// root, where the derivative touches zero without changing sign.  Two
	// genuine roots this close together enclose a bump far smaller than the
	// rounding error of the coordinates, so nothing is lost by dropping
	// them.
	BB, AC4 := B*B, 4*A*C
	disc := BB - AC4
	if disc <= 16*epsilon*(BB+math.Abs(AC4)) {
		return ts
	}

	// The stable form of the quadratic formula avoids cancellation between
	// B and the root of the discriminant.  Where A is zero, or nearly so,
	// q/A is infinite or large and falls outside (0, 1).
	q := -(B + math.Copysign(math.Sqrt(disc), B)) / 2
	for _, t := range [2]float64{q / A, C / q} {
		if t > 0 && t < 1 {
			ts = append(ts, t)
		}
	}
	return ts
}

// epsilon is the relative rounding error of a float64 operation.
const epsilon = 0x1p-53

// lerp interpolates linearly between a and b.
func lerp(a, b vec.Vec2, t float64) vec.Vec2 {
	s := 1 - t
	return vec.Vec2{
		X: s*a.X + t*b.X,
		Y: s*a.Y + t*b.Y,
	}
}
