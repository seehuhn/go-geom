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

// TestPolylineStrokeBBoxSegment checks the box round a single segment: half
// the line width to either side of each end point.  That reaches half the
// line width past the ends, where a butt cap in fact draws nothing.
func TestPolylineStrokeBBoxSegment(t *testing.T) {
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 10, Y: 0}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, false, StrokeOptions{
		Width: 4, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	checkBBox(t, got, ok, rect.Rect{LLx: -2, LLy: -2, URx: 12, URy: 2})
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
	// reaches no further than a round one
	bevelled, _ := PolylineStrokeBBox([][]vec.Vec2{spikeTriangle}, true, StrokeOptions{
		Width: lw, Join: JoinMiter, MiterLimit: 1,
	})
	if bevelled != round {
		t.Errorf("bevelled bbox = %v, want the round one %v", bevelled, round)
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

// TestPolylineStrokeBBoxStraightVertex checks a vertex the stroke runs
// straight through: there is no corner there for a miter to reach beyond, so
// the stroke stays within half the line width of the path.
func TestPolylineStrokeBBoxStraightVertex(t *testing.T) {
	pts := []vec.Vec2{{X: 0, Y: 0}, {X: 3, Y: 4}, {X: 6, Y: 8}}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{pts}, false, StrokeOptions{
		Width: 2, Join: JoinMiter, MiterLimit: DefaultMiterLimit,
	})
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 7, URy: 9})
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
	if want := -lw / 2.0; math.Abs(open.LLx-want) > 1e-9 {
		t.Errorf("bbox starts at x = %g, want %g: an end point carries no join",
			open.LLx, want)
	}
}

// TestPolylineStrokeBBoxDegenerate checks the paths which have no direction
// to work from: no vertices at all, and a sub-path whose vertices coincide.
func TestPolylineStrokeBBoxDegenerate(t *testing.T) {
	opt := StrokeOptions{Width: 2, Cap: CapSquare, Join: JoinMiter, MiterLimit: DefaultMiterLimit}

	if _, ok := PolylineStrokeBBox(nil, false, opt); ok {
		t.Error("an empty path reported a bounding box")
	}
	if _, ok := PolylineStrokeBBox([][]vec.Vec2{{}}, false, opt); ok {
		t.Error("a sub-path with no vertices reported a bounding box")
	}

	// a single point, and a pair on top of each other: the cap can point
	// anywhere, so the box allows for the worst case
	r := math.Sqrt2
	want := rect.Rect{LLx: -r, LLy: -r, URx: r, URy: r}
	got, ok := PolylineStrokeBBox([][]vec.Vec2{{{X: 0, Y: 0}}}, false, opt)
	checkBBox(t, got, ok, want)
	got, ok = PolylineStrokeBBox([][]vec.Vec2{{{X: 0, Y: 0}, {X: 0, Y: 0}}}, false, opt)
	checkBBox(t, got, ok, want)

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
	checkBBox(t, got, ok, rect.Rect{LLx: -1, LLy: -1, URx: 31, URy: 21})
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
