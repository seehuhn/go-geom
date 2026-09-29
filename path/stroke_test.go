// seehuhn.de/go/geom - two-dimensional geometry
// Copyright (C) 2026  Jochen Voss <voss@seehuhn.de>
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

package path

import (
	"math"
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

// TestDistinctVertices checks the collapsing of coincident vertices on its
// own, including that an input with nothing to collapse is handed back as
// it is.
func TestDistinctVertices(t *testing.T) {
	a, b, c := vec.Vec2{X: 0, Y: 0}, vec.Vec2{X: 1, Y: 0}, vec.Vec2{X: 1, Y: 1}
	for _, tc := range []struct {
		in     []vec.Vec2
		closed bool
		want   []vec.Vec2
	}{
		{nil, false, nil},
		{[]vec.Vec2{}, true, []vec.Vec2{}},
		{[]vec.Vec2{a}, true, []vec.Vec2{a}},
		{[]vec.Vec2{a, a}, false, []vec.Vec2{a}},
		{[]vec.Vec2{a, a}, true, []vec.Vec2{a}},
		{[]vec.Vec2{a, b, c}, true, []vec.Vec2{a, b, c}},
		{[]vec.Vec2{a, a, b, b, c, c}, false, []vec.Vec2{a, b, c}},
		{[]vec.Vec2{a, b, c, a}, false, []vec.Vec2{a, b, c, a}},
		{[]vec.Vec2{a, b, c, a}, true, []vec.Vec2{a, b, c}},
		{[]vec.Vec2{a, b, c, c, a, a}, true, []vec.Vec2{a, b, c}},
		{[]vec.Vec2{a, b, a}, true, []vec.Vec2{a, b}},
	} {
		got := distinctVertices(tc.in, tc.closed)
		if len(got) != len(tc.want) {
			t.Errorf("distinctVertices(%v, %v) = %v, want %v", tc.in, tc.closed, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("distinctVertices(%v, %v) = %v, want %v", tc.in, tc.closed, got, tc.want)
				break
			}
		}
	}

	clean := []vec.Vec2{a, b, c}
	if got := distinctVertices(clean, false); len(got) != 3 || &got[0] != &clean[0] {
		t.Error("a sub-path with nothing to collapse was copied")
	}
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
