// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import (
	"math"
	"math/rand/v2"
	"testing"

	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

// TestCubicAtEndpoints checks that the ends of the curve are reproduced
// exactly, since a caller which joins segments end to end relies on it.
func TestCubicAtEndpoints(t *testing.T) {
	p0 := vec.Vec2{X: 1.5, Y: -2.25}
	c1 := vec.Vec2{X: 17, Y: 3}
	c2 := vec.Vec2{X: -4, Y: 11}
	p3 := vec.Vec2{X: 6.125, Y: 0.5}

	if got := (Cubic{p0, c1, c2, p3}).At(0); got != p0 {
		t.Errorf("at t=0: got %v, want %v", got, p0)
	}
	if got := (Cubic{p0, c1, c2, p3}).At(1); got != p3 {
		t.Errorf("at t=1: got %v, want %v", got, p3)
	}
}

// TestCubicAtBernstein checks the interpolation against the Bernstein form
// of the same curve.
func TestCubicAtBernstein(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	c1 := vec.Vec2{X: 1, Y: 3}
	c2 := vec.Vec2{X: 4, Y: -1}
	p3 := vec.Vec2{X: 5, Y: 2}

	for i := range 21 {
		u := float64(i) / 20
		s := 1 - u
		want := vec.Vec2{
			X: s*s*s*p0.X + 3*s*s*u*c1.X + 3*s*u*u*c2.X + u*u*u*p3.X,
			Y: s*s*s*p0.Y + 3*s*s*u*c1.Y + 3*s*u*u*c2.Y + u*u*u*p3.Y,
		}
		got := Cubic{p0, c1, c2, p3}.At(u)
		if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
			t.Errorf("at t=%g: got %v, want %v", u, got, want)
		}
	}
}

// TestCubicAtLine checks that a curve whose control points are spaced
// evenly along a straight line is traversed at a constant rate.
func TestCubicAtLine(t *testing.T) {
	p0 := vec.Vec2{X: 0, Y: 0}
	c1 := vec.Vec2{X: 1, Y: 2}
	c2 := vec.Vec2{X: 2, Y: 4}
	p3 := vec.Vec2{X: 3, Y: 6}

	for i := range 11 {
		u := float64(i) / 10
		got := Cubic{p0, c1, c2, p3}.At(u)
		want := vec.Vec2{X: 3 * u, Y: 6 * u}
		if math.Abs(got.X-want.X) > 1e-12 || math.Abs(got.Y-want.Y) > 1e-12 {
			t.Errorf("at t=%g: got %v, want %v", u, got, want)
		}
	}
}

// TestCubicAtConvexHull checks that the curve stays inside the bounding box
// of its control points.
func TestCubicAtConvexHull(t *testing.T) {
	p0 := vec.Vec2{X: -1, Y: 0}
	c1 := vec.Vec2{X: 0, Y: 100}
	c2 := vec.Vec2{X: 2, Y: -50}
	p3 := vec.Vec2{X: 3, Y: 1}

	bbox, _ := DataFromPath(func(yield func(Command, []vec.Vec2) bool) {
		yield(CmdMoveTo, []vec.Vec2{p0})
		yield(CmdCubeTo, []vec.Vec2{c1, c2, p3})
	}).Iter().BBox()

	for i := range 101 {
		u := float64(i) / 100
		p := Cubic{p0, c1, c2, p3}.At(u)
		if p.X < bbox.LLx || p.X > bbox.URx || p.Y < bbox.LLy || p.Y > bbox.URy {
			t.Errorf("at t=%g: point %v outside %v", u, p, bbox)
		}
	}
}

// deriv returns the derivative of the one-dimensional cubic Bezier curve with
// coefficients a at t, from the Bernstein form, together with the size of the
// terms it sums, against which a result near zero can be judged.
func deriv(a [4]float64, t float64) (d, scale float64) {
	s := 1 - t
	u, v, w := 3*s*s*(a[1]-a[0]), 6*s*t*(a[2]-a[1]), 3*t*t*(a[3]-a[2])
	return u + v + w, math.Abs(u) + math.Abs(v) + math.Abs(w)
}

func TestCubicExtrema(t *testing.T) {
	r := math.Sqrt(12) / 12
	big := math.Ldexp(1, 600)
	cases := []struct {
		name string
		x, y [4]float64
		want []float64
	}{
		{"one in x", [4]float64{0, 50, 50, 0}, [4]float64{0, 1, 2, 3}, []float64{0.5}},
		{"two in y", [4]float64{0, 1, 2, 3}, [4]float64{0, 1, -1, 0}, []float64{0.5 - r, 0.5 + r}},
		{"merged", [4]float64{0, 50, 50, 0}, [4]float64{0, 1, -1, 0}, []float64{0.5 - r, 0.5, 0.5 + r}},
		{"shared", [4]float64{0, 50, 50, 0}, [4]float64{0, 7, 7, 0}, []float64{0.5}},
		{"near-linear derivative", [4]float64{0, 0.5, 0.5, 1e-9}, [4]float64{}, []float64{0.5 + 0.25e-9}},
		{"at the ends only", [4]float64{0, 0, 1, 1}, [4]float64{1, 1, 0, 0}, nil},
		{"line", [4]float64{0, 1, 2, 3}, [4]float64{5, 4, 3, 2}, nil},
		{"point", [4]float64{1, 1, 1, 1}, [4]float64{2, 2, 2, 2}, nil},
		{"huge", [4]float64{0, 50 * big, 50 * big, 0}, [4]float64{0, big, -big, 0}, []float64{0.5 - r, 0.5, 0.5 + r}},
		{"tiny", [4]float64{0, 50 / big, 50 / big, 0}, [4]float64{0, 1 / big, -1 / big, 0}, []float64{0.5 - r, 0.5, 0.5 + r}},
		{"far from the origin", [4]float64{1e9, 1e9 + 50, 1e9 + 50, 1e9}, [4]float64{-1e9, -1e9, -1e9, -1e9}, []float64{0.5}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var curve Cubic
			for i := range 4 {
				curve[i] = vec.Vec2{X: c.x[i], Y: c.y[i]}
			}
			got := curve.Extrema()
			if len(got) != len(c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
			for i := range got {
				if math.Abs(got[i]-c.want[i]) > 1e-15 {
					t.Errorf("got %v, want %v", got, c.want)
				}
			}
		})
	}
}

// TestCubicExtremaTouch checks that a derivative which only touches zero
// gives no extremum, also where rounding perturbs the discriminant.
func TestCubicExtremaTouch(t *testing.T) {
	for _, s := range []float64{1, 0.1, 1.0 / 3, 7.3, 1e-200, 1e200} {
		// x' is proportional to (t - 1/2)^2
		curve := Cubic{{}, {X: 0.25 * s, Y: 1}, {X: 0, Y: 2}, {X: 0.25 * s, Y: 3}}
		if got := curve.Extrema(); len(got) != 0 {
			t.Errorf("scale %g: got %v, want none", s, got)
		}
	}
}

// TestCubicExtremaProperties checks the contract on random curves: the
// parameters lie in (0, 1), increase, and the derivative vanishes there.
func TestCubicExtremaProperties(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for range 1000 {
		var x, y [4]float64
		var curve Cubic
		for i := range 4 {
			x[i], y[i] = rng.NormFloat64()*100, rng.NormFloat64()*100
			curve[i] = vec.Vec2{X: x[i], Y: y[i]}
		}
		ts := curve.Extrema()
		for i, u := range ts {
			if u <= 0 || u >= 1 || i > 0 && u <= ts[i-1] {
				t.Fatalf("%v: invalid parameters %v", curve, ts)
			}
			dx, sx := deriv(x, u)
			dy, sy := deriv(y, u)
			if math.Abs(dx) > 1e-12*sx && math.Abs(dy) > 1e-12*sy {
				t.Errorf("%v: derivative (%g, %g) at %g", curve, dx, dy, u)
			}
		}
	}
}

// TestCubicBBox checks that sampled points of random curves lie in the box,
// and that the box is touched on all four sides.
func TestCubicBBox(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	const n = 10000
	for range 100 {
		var curve Cubic
		for i := range curve {
			curve[i] = vec.Vec2{X: rng.NormFloat64() * 100, Y: rng.NormFloat64() * 100}
		}
		box := curve.BBox()
		tol := 1e-12 * max(box.Dx(), box.Dy())
		var seen rect.Rect
		for i := range n + 1 {
			p := curve.At(float64(i) / n)
			if p.X < box.LLx-tol || p.X > box.URx+tol || p.Y < box.LLy-tol || p.Y > box.URy+tol {
				t.Fatalf("%v: point %v outside %v", curve, p, box)
			}
			if i == 0 {
				seen = rect.Rect{LLx: p.X, LLy: p.Y, URx: p.X, URy: p.Y}
			}
			seen.Add(p.X, p.Y)
		}
		// samples 1/n apart come within (1/n)^2 of a smooth maximum
		tol = 1e-6 * max(box.Dx(), box.Dy())
		if seen.LLx > box.LLx+tol || seen.LLy > box.LLy+tol || seen.URx < box.URx-tol || seen.URy < box.URy-tol {
			t.Errorf("%v: samples span %v, box %v", curve, seen, box)
		}
	}
}
