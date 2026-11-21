// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

// nearlyEqual reports whether two rectangles agree to within tol on every
// edge.
func nearlyEqual(a, b rect.Rect, tol float64) bool {
	return math.Abs(a.LLx-b.LLx) <= tol && math.Abs(a.LLy-b.LLy) <= tol &&
		math.Abs(a.URx-b.URx) <= tol && math.Abs(a.URy-b.URy) <= tol
}

func checkBBox(t *testing.T, got rect.Rect, ok bool, want rect.Rect) {
	t.Helper()
	if !ok {
		t.Fatal("no bounding box")
	}
	if !nearlyEqual(got, want, 1e-9) {
		t.Errorf("bbox = %v, want %v", got, want)
	}
}

// TestPolylineStrokeBBoxSegment checks the box round a single segment with
// butt caps: half the line width to either side, and nothing beyond the ends.
func TestPolylineStrokeBBoxSegment(t *testing.T) {
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, false, StrokeOptions{
		Width: 4, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	checkBBox(t, got, ok, rect.Rect{LLx: 0, LLy: -2, URx: 10, URy: 2})

	// on a diagonal, the butt ends reach only to their corners
	diagonal := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 10}}
	got, ok = PolylineStrokeBBox([][]vec.Vec2{diagonal}, false, StrokeOptions{Width: 2})
	r := math.Sqrt2 / 2
	checkBBox(t, got, ok, rect.Rect{LLx: -r, LLy: -r, URx: 10 + r, URy: 10 + r})
}

// TestPolylineStrokeBBoxRoundCap checks that a round cap reaches half the
// line width beyond the end point, in the directions the cap faces.
func TestPolylineStrokeBBoxRoundCap(t *testing.T) {
	opt := StrokeOptions{Width: 2, Cap: CapRound}

	flat := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{flat}, false, opt)
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 11, URy: 1})

	diagonal := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 10}}
	got, ok = PolylineStrokeBBox([][]vec.Vec2{diagonal}, false, opt)
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 11, URy: 11})
}

// TestPolylineStrokeBBoxShortEndSegment checks a square cap on an end segment
// shorter than half the line width.  The cap reaches out from the end point
// only, and not back past the segment's other end, where the stroke is
// bounded by the vertical segment before it.
func TestPolylineStrokeBBoxShortEndSegment(t *testing.T) {
	const lw = 20
	pts := []vec.Vec2{{X: 0, Y: -100}, {X: 0, Y: 0}, {X: -math.Sqrt2 / 2, Y: math.Sqrt2 / 2}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, false, StrokeOptions{
		Width: lw, Cap: CapSquare, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	if !ok {
		t.Fatal("no bounding box")
	}
	if want := lw / 2.0; math.Abs(got.URx-want) > 1e-9 {
		t.Errorf("bbox ends at x = %g, want %g", got.URx, want)
	}
}

// TestPolylineStrokeBBoxRoundJoin checks that a round join reaches half the
// line width beyond the vertex on the outer side of the corner, and only
// there.
func TestPolylineStrokeBBoxRoundJoin(t *testing.T) {
	// a sharp turn at (10, 0), with the outer side facing the positive x
	// axis, the lower right, and the upper right respectively
	opt := StrokeOptions{Width: 2, Join: JoinRound}
	for _, tc := range []struct {
		pts  []vec.Vec2
		want rect.Rect
	}{
		{
			pts:  []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 0, Y: 1}},
			want: rect.Rect{LLx: -1 / math.Hypot(10, 1), LLy: -1, URx: 11, URy: 1 + 10/math.Hypot(10, 1)},
		},
		{
			pts:  []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}},
			want: rect.Rect{LLx: 0, LLy: -1, URx: 11, URy: 10},
		},
		{
			pts:  []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: -10}},
			want: rect.Rect{LLx: 0, LLy: -10, URx: 11, URy: 1},
		},
	} {
		got, ok := PolylineStrokeBBox([][]vec.Vec2{tc.pts}, false, opt)
		checkBBox(t, got, ok, tc.want)
	}
}

// TestPolylineStrokeBBoxSquareCap checks the extra reach of a projecting
// square cap: none beyond the half width for a cap along an axis, since the
// cap then fills the box the butt cap was already given, and half the width
// times the square root of two at 45 degrees.
func TestPolylineStrokeBBoxSquareCap(t *testing.T) {
	opt := StrokeOptions{Width: 2, Cap: CapSquare, Join: JoinMiter, MiterLimit: DefaultMiterLimit}

	flat := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{flat}, false, opt)
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 11, URy: 1})

	diagonal := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 10}}
	got, ok = PolylineStrokeBBox([][]vec.Vec2{diagonal}, false, opt)
	r := math.Sqrt2
	checkBBox(t, got, ok, rect.Rect{LLx: -r, LLy: -r, URx: 10 + r, URy: 10 + r})

	// the cap has to reach further than the butt cap it replaces
	butt, _ := PolylineStrokeBBox([][]vec.Vec2{diagonal}, false, StrokeOptions{
		Width: 2, Cap: CapButt, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	if !got.Covers(butt) || got == butt {
		t.Errorf("square cap %v does not reach beyond butt cap %v", got, butt)
	}
}

// TestPolylineStrokeBBoxCapsOnlyAtEnds checks that a cap is taken into
// account at the two ends of an open path and nowhere else: a closed path
// has no ends, and an interior vertex carries a join.
func TestPolylineStrokeBBoxCapsOnlyAtEnds(t *testing.T) {
	// a diagonal zig-zag, so that a square cap would show up as a wider box
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 20, Y: 0}}
	opt := StrokeOptions{Width: 2, Join: JoinRound}

	butt, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, true, opt)
	if !ok {
		t.Fatal("no bounding box")
	}
	opt.Cap = CapSquare
	square, _ := PolylineStrokeBBox([][]vec.Vec2{pts}, true, opt)
	if square != butt {
		t.Errorf("a closed path was capped: %v, want %v", square, butt)
	}

	// open, the two ends are capped but the apex is not
	open, _ := PolylineStrokeBBox([][]vec.Vec2{pts}, false, opt)
	if want := 10.0 + 1; open.URy != want {
		t.Errorf("apex reaches to %v, want %v: the join was capped", open.URy, want)
	}
	if want := -math.Sqrt2; !nearlyEqual(open, rect.Rect{
		LLx: want, LLy: want, URx: 20 - want, URy: 11,
	}, 1e-9) {
		t.Errorf("open bbox = %v, want its ends capped", open)
	}
}

// spikeTriangle is a closed path with one corner sharp enough for its miter
// to reach well beyond half the line width, but not so sharp that the
// default miter limit bevels it.  The sharp corner is at the origin and its
// bisector is the negative x axis.
const spikeHalfWidth = 30.0

var spikeTriangle = []vec.Vec2{
	{X: 0, Y: 0},
	{X: 100, Y: spikeHalfWidth},
	{X: 100, Y: -spikeHalfWidth},
}

// spikeTipX is how far the miter at the sharp corner of [spikeTriangle]
// reaches along the negative x axis, for a stroke of width lw: half the line
// width divided by the sine of the corner's half-angle.
func spikeTipX(lw float64) float64 {
	return -lw / 2 / math.Sin(math.Atan2(spikeHalfWidth, 100))
}

// spikeCornerX is how far the outer corners of the two segments meeting at
// the sharp corner of [spikeTriangle] reach along the negative x axis, for a
// stroke of width lw.
func spikeCornerX(lw float64) float64 {
	return -lw / 2 * spikeHalfWidth / math.Hypot(100, spikeHalfWidth)
}

// TestPolylineStrokeBBoxMiter checks that a miter join is allowed for where
// it reaches beyond the line width, that a round join is not, and that a
// corner too sharp for the miter limit is bevelled.
func TestPolylineStrokeBBoxMiter(t *testing.T) {
	const lw = 4

	miter, ok := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	if !ok {
		t.Fatal("no bounding box")
	}
	if want := spikeTipX(lw); math.Abs(miter.LLx-want) > 1e-9 {
		t.Errorf("bbox starts at x = %g, want the miter tip at %g", miter.LLx, want)
	}

	round, _ := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinRound,
	})
	if want := -lw / 2.0; math.Abs(round.LLx-want) > 1e-9 {
		t.Errorf("a round join starts at x = %g, want %g", round.LLx, want)
	}

	// a limit too short for any miter leaves the bevelled corner, which
	// reaches only to the outer corners of the two segments
	bevelled, _ := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinMiter, MiterLimit: 1,
	})
	if want := spikeCornerX(lw); math.Abs(bevelled.LLx-want) > 1e-9 {
		t.Errorf("a bevelled join starts at x = %g, want %g", bevelled.LLx, want)
	}
}

// TestPolylineStrokeBBoxDefaultMiterLimit checks that a zero miter limit
// stands for the default one, rather than bevelling every corner.
func TestPolylineStrokeBBoxDefaultMiterLimit(t *testing.T) {
	const lw = 4

	implied, ok := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinMiter,
	})
	if !ok {
		t.Fatal("no bounding box")
	}
	given, _ := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	if implied != given {
		t.Errorf("bbox = %v, want the default limit's %v", implied, given)
	}
}

// TestPolylineStrokeBBoxDashed checks that a dashed stroke is allowed the
// caps of a dash ending or starting on a corner.  At the bevelled sharp corner
// of [spikeTriangle], a square cap pointing along either segment reaches
// further than the bevel, and a round cap reaches half the line width.
func TestPolylineStrokeBBoxDashed(t *testing.T) {
	const lw = 4
	h := lw / 2.0
	hyp := math.Hypot(100, spikeHalfWidth)
	for _, tc := range []struct {
		cap  CapStyle
		want float64
	}{
		{CapButt, spikeCornerX(lw)},
		{CapSquare, -h * (100 + spikeHalfWidth) / hyp},
		{CapRound, -h},
	} {
		opt := StrokeOptions{Width: lw, Cap: tc.cap, Join: JoinBevel}
		undashed, ok := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, opt)
		if !ok {
			t.Fatal("no bounding box")
		}
		opt.Dashed = true
		dashed, _ := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, opt)
		if math.Abs(dashed.LLx-tc.want) > 1e-9 {
			t.Errorf("%v caps: dashed bbox starts at x = %g, want %g", tc.cap, dashed.LLx, tc.want)
		}
		if !dashed.Covers(undashed) {
			t.Errorf("%v caps: dashed bbox %v does not cover the undashed %v", tc.cap, dashed, undashed)
		}
	}

	// a zero-length line draws nothing without round caps, dashed or not
	p := vec.Vec2{X: 1, Y: 2}
	if _, ok := PolylineStrokeBBox([][]vec.Vec2{{p, p}}, false, StrokeOptions{
		Width: lw, Cap: CapSquare, Dashed: true,
	}); ok {
		t.Error("a dashed zero-length line reported a bounding box")
	}
}

// TestPolylineStrokeBBoxStraightVertex checks a vertex the stroke runs
// straight through: there is no corner there for a join to fill, so the
// stroke is the rectangle along the whole path.
func TestPolylineStrokeBBoxStraightVertex(t *testing.T) {
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 3, Y: 4}, {X: 6, Y: 8}}
	for _, join := range []JoinStyle{JoinMiter, JoinRound, JoinBevel} {
		got, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, false, StrokeOptions{
			Width: 2, Join: join, MiterLimit: DefaultMiterLimit,
		})
		checkBBox(t, got, ok, rect.Rect{LLx: -0.8, LLy: -0.6, URx: 6.8, URy: 8.6})
	}
}

// TestPolylineStrokeBBoxMiterIsExact checks that a miter join is taken in as
// the point it is, rather than as a disc of the miter length: the stroke
// round an axis-aligned square reaches exactly half the line width beyond
// it, although each corner's miter is longer than that.
func TestPolylineStrokeBBoxMiterIsExact(t *testing.T) {
	const lw = 10
	square := []vec.Vec2{{X: 20, Y: 20}, {X: 80, Y: 20}, {X: 80, Y: 80}, {X: 20, Y: 80}}

	got, ok := PolylineStrokeBBox([][]vec.Vec2{square}, true, StrokeOptions{
		Width: lw, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	checkBBox(t, got, ok, rect.Rect{LLx: 15, LLy: 15, URx: 85, URy: 85})
}

// TestPolylineStrokeBBoxOpenPathEnds checks that the first and last vertex
// of an open path carry a cap rather than a join, so that the sharp corner
// of [spikeTriangle] is not mitered when the path is left open.
func TestPolylineStrokeBBoxOpenPathEnds(t *testing.T) {
	const lw = 4

	// the spike is the first vertex, and the last one closes back onto it
	open, ok := PolylineStrokeBBox([][]vec.Vec2{
		{{X: 0, Y: 0}, {X: 100, Y: spikeHalfWidth}, {X: 100, Y: -spikeHalfWidth}, {X: 0, Y: 0}},
	}, false, StrokeOptions{Width: lw, Join: JoinMiter, MiterLimit: DefaultMiterLimit})
	if !ok {
		t.Fatal("no bounding box")
	}
	if want := spikeCornerX(lw); math.Abs(open.LLx-want) > 1e-9 {
		t.Errorf("bbox starts at x = %g, want %g: an end point carries no join",
			open.LLx, want)
	}
}

// TestPolylineStrokeBBoxDegenerate checks the paths which have no direction
// to work from: no vertices at all, and zero-length lines, which only round
// caps draw.
func TestPolylineStrokeBBoxDegenerate(t *testing.T) {
	opt := StrokeOptions{Width: 2, Join: JoinMiter, MiterLimit: DefaultMiterLimit}

	if _, ok := PolylineStrokeBBox(nil, false, opt); ok {
		t.Error("an empty path reported a bounding box")
	}
	if _, ok := PolylineStrokeBBox([][]vec.Vec2{{}}, false, opt); ok {
		t.Error("a sub-path with no vertices reported a bounding box")
	}

	p := vec.Vec2{X: 5, Y: 7}
	zeroLength := [][][]vec.Vec2{{{p}}, {{p, p}}}
	for _, closed := range []bool{false, true} {
		for _, subpaths := range zeroLength {
			for _, lineCap := range []CapStyle{CapButt, CapSquare} {
				opt.Cap = lineCap
				if got, ok := PolylineStrokeBBox(subpaths, closed, opt); ok {
					t.Errorf("zero-length line %v, closed=%v, with %v caps: bbox %v, want none",
						subpaths, closed, lineCap, got)
				}
			}

			opt.Cap = CapRound
			got, ok := PolylineStrokeBBox(subpaths, closed, opt)
			checkBBox(t, got, ok, rect.Rect{LLx: 4, LLy: 6, URx: 6, URy: 8})
		}
	}

	// a zero-length line beside a drawn one adds nothing without round caps
	opt.Cap = CapSquare
	got, ok := PolylineStrokeBBox([][]vec.Vec2{
		{{X: 0, Y: 0}, {X: 10, Y: 0}},
		{p, p},
	}, false, opt)
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 11, URy: 1})

	// a zero-length first segment still leaves a direction to cap along
	got, ok = PolylineStrokeBBox([][]vec.Vec2{
		{{X: 0, Y: 0}, {X: 0, Y: 0}, {X: 10, Y: 0}},
	}, false, opt)
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 11, URy: 1})
}

// TestPolylineStrokeBBoxRepeatedVertex checks that a vertex given twice in
// a row, and a closed path which returns to its first vertex explicitly,
// keep the miter join a stroker draws there: the zero-length segment is
// skipped, and the segments on either side of it are joined.
func TestPolylineStrokeBBoxRepeatedVertex(t *testing.T) {
	const lw = 4
	opt := StrokeOptions{Width: lw, Join: JoinMiter, MiterLimit: DefaultMiterLimit}
	tip := spikeTipX(lw)

	closedTwice := append(append([]vec.Vec2{}, spikeTriangle...), spikeTriangle[0])
	got, ok := PolylineStrokeBBox([][]vec.Vec2{closedTwice}, true, opt)
	if !ok {
		t.Fatal("no bounding box")
	}
	if math.Abs(got.LLx-tip) > 1e-9 {
		t.Errorf("closed path repeating its start: bbox starts at x = %g, want the miter tip at %g",
			got.LLx, tip)
	}

	// the spike is an interior vertex of an open path, given twice
	open := []vec.Vec2{
		{X: 100, Y: spikeHalfWidth}, {X: 0, Y: 0}, {X: 0, Y: 0}, {X: 100, Y: -spikeHalfWidth},
	}
	got, _ = PolylineStrokeBBox([][]vec.Vec2{open}, false, opt)
	if math.Abs(got.LLx-tip) > 1e-9 {
		t.Errorf("repeated interior vertex: bbox starts at x = %g, want the miter tip at %g",
			got.LLx, tip)
	}

	// an end vertex given twice still carries a cap, pointing along the
	// segment behind the repeat
	opt.Cap = CapSquare
	got, _ = PolylineStrokeBBox([][]vec.Vec2{
		{{X: 0, Y: 0}, {X: 10, Y: 10}, {X: 10, Y: 10}},
	}, false, opt)
	r := math.Sqrt2 * lw / 2
	checkBBox(t, got, true, rect.Rect{LLx: -r, LLy: -r, URx: 10 + r, URy: 10 + r})
}

// TestPolylineStrokeBBoxSubpaths checks that every sub-path is taken in.
func TestPolylineStrokeBBoxSubpaths(t *testing.T) {
	got, ok := PolylineStrokeBBox([][]vec.Vec2{
		{{X: 0, Y: 0}, {X: 10, Y: 0}},
		{{X: 20, Y: 20}, {X: 30, Y: 20}},
	}, false, StrokeOptions{Width: 2, Join: JoinRound})
	checkBBox(t, got, ok, rect.Rect{LLx: 0, LLy: -1, URx: 30, URy: 21})
}

// TestCapStyleString and TestJoinStyleString check the names of the styles.
func TestCapStyleString(t *testing.T) {
	for _, tc := range []struct {
		style CapStyle
		want  string
	}{{CapButt, "butt"}, {CapRound, "round"}, {CapSquare, "square"}} {
		if got := tc.style.String(); got != tc.want {
			t.Errorf("CapStyle(%d) = %q, want %q", tc.style, got, tc.want)
		}
	}
}

func TestJoinStyleString(t *testing.T) {
	for _, tc := range []struct {
		style JoinStyle
		want  string
	}{{JoinMiter, "miter"}, {JoinRound, "round"}, {JoinBevel, "bevel"}} {
		if got := tc.style.String(); got != tc.want {
			t.Errorf("JoinStyle(%d) = %q, want %q", tc.style, got, tc.want)
		}
	}
}

// TestStrokeBBoxPolyline checks that a path of straight segments gets the
// same box as the polylines through its vertices.
func TestStrokeBBoxPolyline(t *testing.T) {
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 10, Y: 10}, {X: 10, Y: 10}, {X: 3, Y: 4}}
	for _, closed := range []bool{false, true} {
		for _, join := range []JoinStyle{JoinMiter, JoinRound, JoinBevel} {
			for _, lineCap := range []CapStyle{CapButt, CapRound, CapSquare} {
				opt := StrokeOptions{Width: 3, Cap: lineCap, Join: join}
				p := &Data{}
				p.MoveTo(pts[0])
				for _, q := range pts[1:] {
					p.LineTo(q)
				}
				if closed {
					p.Close()
				}
				want, _ := PolylineStrokeBBox([][]vec.Vec2{pts}, closed, opt)
				got, ok := p.Iter().StrokeBBox(opt)
				if !ok || !nearlyEqual(got, want, 1e-9) {
					t.Errorf("closed=%v %v %v: bbox %v, want %v", closed, join, lineCap, got, want)
				}
			}
		}
	}
}

// TestStrokeBBoxRoundCurve checks that with round caps and joins a curve's
// box is its tight box grown by half the line width.
func TestStrokeBBoxRoundCurve(t *testing.T) {
	c := Cubic{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}, {X: 10, Y: 0}}
	p := (&Data{}).MoveTo(c[0]).CubeTo(c[1], c[2], c[3])
	got, ok := p.Iter().StrokeBBox(StrokeOptions{Width: 2, Cap: CapRound, Join: JoinRound})
	checkBBox(t, got, ok, c.BBox().Grow(1))
}

// TestStrokeBBoxZeroLength checks sub-paths which draw no segment of
// non-zero length.
func TestStrokeBBoxZeroLength(t *testing.T) {
	pt := vec.Vec2{X: 5, Y: 5}
	round := StrokeOptions{Width: 2, Cap: CapRound}
	butt := StrokeOptions{Width: 2, Cap: CapButt}

	// a lone move draws nothing
	lone := (&Data{}).MoveTo(pt).MoveTo(vec.Vec2{X: 50, Y: 50}).LineTo(vec.Vec2{X: 50, Y: 50})
	got, ok := lone.Iter().StrokeBBox(round)
	checkBBox(t, got, ok, rect.Rect{LLx: 49, LLy: 49, URx: 51, URy: 51})

	// a closed single point is a zero-length line
	closed := (&Data{}).MoveTo(pt).Close()
	got, ok = closed.Iter().StrokeBBox(round)
	checkBBox(t, got, ok, rect.Rect{LLx: 4, LLy: 4, URx: 6, URy: 6})
	if _, ok := closed.Iter().StrokeBBox(butt); ok {
		t.Error("zero-length line with butt caps has a bounding box")
	}

	// a curve whose points all coincide
	curve := (&Data{}).MoveTo(pt).CubeTo(pt, pt, pt)
	got, ok = curve.Iter().StrokeBBox(round)
	checkBBox(t, got, ok, rect.Rect{LLx: 4, LLy: 4, URx: 6, URy: 6})
}

// TestStrokeBBoxCurveMiter checks that a miter join between a curve and a
// line is taken in, and that with round caps and joins the box is exact.
func TestStrokeBBoxCurveMiter(t *testing.T) {
	// a straight curve along the x axis, then a sharp turn back
	p := (&Data{}).MoveTo(vec.Vec2{X: 0, Y: 0}).
		CubeTo(vec.Vec2{X: 3, Y: 0}, vec.Vec2{X: 6, Y: 0}, vec.Vec2{X: 10, Y: 0}).
		LineTo(vec.Vec2{X: 0, Y: 2})
	poly := [][]vec.Vec2{{{X: 0, Y: 0}, {X: 10, Y: 0}, {X: 0, Y: 2}}}
	opt := StrokeOptions{Width: 2, Join: JoinMiter, MiterLimit: 20}
	got, ok := p.Iter().StrokeBBox(opt)
	if !ok {
		t.Fatal("no bounding box")
	}
	miter, _ := PolylineStrokeBBox(poly, false, opt)
	bevel, _ := PolylineStrokeBBox(poly, false, StrokeOptions{Width: 2, Join: JoinBevel})
	if miter == bevel {
		t.Fatal("the corner is bevelled, the test needs a sharper miter")
	}
	if !got.Covers(miter) {
		t.Errorf("bbox %v does not cover the polyline box %v", got, miter)
	}

	opt.Join, opt.Cap = JoinRound, CapRound
	got, _ = p.Iter().StrokeBBox(opt)
	want, _ := PolylineStrokeBBox(poly, false, opt)
	checkBBox(t, got, true, want)
}

// TestStrokeBBoxMiterBendsAway checks that a curve which bends away from
// the corner it meets gives a finite box without a miter limit, which
// contains the stroke of every flattening of the curve.
func TestStrokeBBoxMiterBendsAway(t *testing.T) {
	// a quarter circle from (1, 0) to (0, 1), then a right turn upwards
	const k = 0.5523
	arc := Cubic{{X: 1, Y: 0}, {X: 1, Y: k}, {X: k, Y: 1}, {X: 0, Y: 1}}
	end := vec.Vec2{X: 0, Y: 3}
	p := (&Data{}).MoveTo(arc[0]).CubeTo(arc[1], arc[2], arc[3]).LineTo(end)
	opt := StrokeOptions{Width: 0.1, Join: JoinMiter, MiterLimit: math.Inf(1)}

	got, ok := p.Iter().StrokeBBox(opt)
	if !ok {
		t.Fatal("no bounding box")
	}
	if !got.IsFinite() {
		t.Fatalf("bbox %v is infinite", got)
	}
	for pieces := 1; pieces <= 20; pieces++ {
		poly := []vec.Vec2{arc[0]}
		for i := 1; i <= pieces; i++ {
			poly = append(poly, arc.At(float64(i)/float64(pieces)))
		}
		poly = append(poly, end)
		want, _ := PolylineStrokeBBox([][]vec.Vec2{poly}, false, opt)
		if !got.Grow(1e-9).Covers(want) {
			t.Errorf("%d pieces: bbox %v does not cover %v", pieces, got, want)
		}
	}
}

// TestStrokeBBoxMiterCircle checks that a circle made of four quarter arcs
// gives a finite box without a miter limit, which contains the stroke of
// every flattening of the arcs.
func TestStrokeBBoxMiterCircle(t *testing.T) {
	const k = 0.5523
	arcs := []Cubic{
		{{X: 1, Y: 0}, {X: 1, Y: k}, {X: k, Y: 1}, {X: 0, Y: 1}},
		{{X: 0, Y: 1}, {X: -k, Y: 1}, {X: -1, Y: k}, {X: -1, Y: 0}},
		{{X: -1, Y: 0}, {X: -1, Y: -k}, {X: -k, Y: -1}, {X: 0, Y: -1}},
		{{X: 0, Y: -1}, {X: k, Y: -1}, {X: 1, Y: -k}, {X: 1, Y: 0}},
	}
	p := (&Data{}).MoveTo(arcs[0][0])
	for _, c := range arcs {
		p.CubeTo(c[1], c[2], c[3])
	}
	p.Close()
	opt := StrokeOptions{Width: 0.1, Join: JoinMiter, MiterLimit: math.Inf(1)}

	got, ok := p.Iter().StrokeBBox(opt)
	if !ok {
		t.Fatal("no bounding box")
	}
	if !got.IsFinite() {
		t.Fatalf("bbox %v is infinite", got)
	}

	r := rand.New(rand.NewPCG(5, 6))
	for trial := range 200 {
		var poly []vec.Vec2
		for _, c := range arcs {
			pieces := 1 + r.IntN(10)
			ts := make([]float64, pieces)
			for i := range pieces - 1 {
				ts[i] = r.Float64()
			}
			ts[pieces-1] = 1
			slices.Sort(ts)
			poly = append(poly, c[0])
			for _, u := range ts[:pieces-1] {
				poly = append(poly, c.At(u))
			}
		}
		want, _ := PolylineStrokeBBox([][]vec.Vec2{poly}, true, opt)
		if !got.Grow(1e-9).Covers(want) {
			t.Fatalf("trial %d: bbox %v does not cover %v", trial, got, want)
		}
	}
}

// TestStrokeBBoxDegenerateCurve checks curves which run straight because
// their control points sit on the end points.
func TestStrokeBBoxDegenerateCurve(t *testing.T) {
	a, b, c := vec.Vec2{X: 0, Y: 0}, vec.Vec2{X: 10, Y: 0}, vec.Vec2{X: 10, Y: 5}
	p := (&Data{}).MoveTo(a).CubeTo(a, a, b).CubeTo(c, c, c)
	opt := StrokeOptions{Width: 2, Cap: CapRound, Join: JoinRound}
	got, ok := p.Iter().StrokeBBox(opt)
	want, _ := PolylineStrokeBBox([][]vec.Vec2{{a, b, c}}, false, opt)
	checkBBox(t, got, ok, want)

	opt.Join, opt.MiterLimit = JoinMiter, 10
	got, _ = p.Iter().StrokeBBox(opt)
	want, _ = PolylineStrokeBBox([][]vec.Vec2{{a, b, c}}, false, opt)
	if !got.Covers(want) {
		t.Errorf("bbox %v does not cover the polyline box %v", got, want)
	}
}

// TestStrokeBBoxSubpathStart checks where sub-paths start which do not
// begin with a CmdMoveTo: at the origin for the first, and at the start of
// the sub-path just closed after a CmdClose.
func TestStrokeBBoxSubpathStart(t *testing.T) {
	opt := StrokeOptions{Width: 2, Cap: CapButt, Join: JoinBevel}

	p := (&Data{}).LineTo(vec.Vec2{X: 10, Y: 0})
	got, ok := p.Iter().StrokeBBox(opt)
	checkBBox(t, got, ok, rect.Rect{LLx: 0, LLy: -1, URx: 10, URy: 1})

	p = (&Data{}).MoveTo(vec.Vec2{X: 20, Y: 0}).LineTo(vec.Vec2{X: 30, Y: 0}).Close().
		LineTo(vec.Vec2{X: 20, Y: 10})
	got, ok = p.Iter().StrokeBBox(opt)
	checkBBox(t, got, ok, rect.Rect{LLx: 19, LLy: -1, URx: 30, URy: 10})
}

// TestStrokeBBoxFlattened checks that the box contains the stroke of the
// path with each curve replaced by a polyline whose vertices lie on the
// curve, for random paths, vertices and stroke options.
func TestStrokeBBoxFlattened(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	point := func() vec.Vec2 {
		return vec.Vec2{X: r.Float64()*20 - 10, Y: r.Float64()*20 - 10}
	}
	for range 2000 {
		closed := r.IntN(2) == 0
		pieces := 1 + r.IntN(40)

		cur := point()
		p := (&Data{}).MoveTo(cur)
		poly := []vec.Vec2{cur}
		for range 1 + r.IntN(3) {
			if r.IntN(2) == 0 {
				cur = point()
				p.LineTo(cur)
				poly = append(poly, cur)
				continue
			}
			c := Cubic{cur, point(), point(), point()}
			p.CubeTo(c[1], c[2], c[3])
			ts := make([]float64, pieces)
			for k := range pieces - 1 {
				ts[k] = r.Float64()
			}
			ts[pieces-1] = 1
			slices.Sort(ts)
			for _, u := range ts {
				poly = append(poly, c.At(u))
			}
			cur = c[3]
		}
		if closed {
			p.Close()
		}

		opt := StrokeOptions{
			Width:      0.1 + 3*r.Float64(),
			Cap:        CapStyle(r.IntN(3)),
			Join:       JoinStyle(r.IntN(3)),
			MiterLimit: 1 + 20*r.Float64(),
			Dashed:     r.IntN(2) == 0,
		}
		if r.IntN(3) == 0 {
			opt.MiterLimit = 1e4
		}
		got, ok := p.Iter().StrokeBBox(opt)
		want, _ := PolylineStrokeBBox([][]vec.Vec2{poly}, closed, opt)
		const tol = 1e-9
		grown := got.Grow(tol)
		if !ok || !grown.Covers(want) {
			t.Fatalf("closed=%v %+v, %d pieces: bbox %v does not cover %v",
				closed, opt, pieces, got, want)
		}
	}
}

// TestStrokeBBoxWorstCase checks the orderings between stroke options which
// let a caller bound a stroke whose options are not known.
func TestStrokeBBoxWorstCase(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	point := func() vec.Vec2 {
		return vec.Vec2{X: r.Float64()*20 - 10, Y: r.Float64()*20 - 10}
	}
	randomOptions := func() StrokeOptions {
		return StrokeOptions{
			Width:      0.1 + 3*r.Float64(),
			Cap:        CapStyle(r.IntN(3)),
			Join:       JoinStyle(r.IntN(3)),
			MiterLimit: 1 + 20*r.Float64(),
		}
	}
	for range 2000 {
		p := &Data{}
		var polylines [][]vec.Vec2
		for range 1 + r.IntN(2) {
			cur := point()
			p.MoveTo(cur)
			pts := []vec.Vec2{cur}
			for range 1 + r.IntN(4) {
				switch r.IntN(3) {
				case 0:
					cur = point()
					p.LineTo(cur)
				case 1:
					c := Cubic{cur, point(), point(), point()}
					p.CubeTo(c[1], c[2], c[3])
					cur = c[3]
				default:
					p.LineTo(cur) // a segment of length zero
				}
				pts = append(pts, cur)
			}
			if r.IntN(2) == 0 {
				p.Close()
			}
			polylines = append(polylines, pts)
		}
		closed := r.IntN(2) == 0

		checkWorstCase(t, "StrokeBBox", p.Iter().StrokeBBox, randomOptions())
		checkWorstCase(t, "PolylineStrokeBBox", func(opt StrokeOptions) (rect.Rect, bool) {
			return PolylineStrokeBBox(polylines, closed, opt)
		}, randomOptions())
	}
}

// checkWorstCase checks that each documented change of the stroke options
// gives a rectangle covering the one for opt.
func checkWorstCase(t *testing.T, name string, bbox func(StrokeOptions) (rect.Rect, bool), opt StrokeOptions) {
	t.Helper()
	base, ok := bbox(opt)
	if !ok {
		return
	}
	covers := func(what string, wider StrokeOptions, want rect.Rect) rect.Rect {
		got, ok := bbox(wider)
		if !ok {
			t.Fatalf("%s %+v: %s draws nothing", name, opt, what)
		}
		if !got.Grow(1e-9).Covers(want) {
			t.Fatalf("%s %+v: %s bbox %v does not cover %v", name, opt, what, got, want)
		}
		return got
	}

	dashed := opt
	dashed.Dashed = true
	covers("dashed", dashed, base)

	if opt.Join != JoinRound {
		// a limit above any in randomOptions keeps curved strokes finite
		miter := opt
		miter.Join = JoinMiter
		miter.MiterLimit = 1e3
		limited := covers("miter with limit 1e3", miter, base)

		miter.MiterLimit = math.Inf(1)
		covers("miter without limit", miter, limited)
	}

	if opt.Cap == CapButt {
		round := opt
		round.Cap = CapRound
		covers("round cap", round, base)
	}
	if opt.Cap != CapRound {
		square := opt
		square.Cap = CapSquare
		covers("square cap", square, base)
	}
}
