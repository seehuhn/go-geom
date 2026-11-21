// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import (
	"math"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

// TestEmpty checks that [Empty] is usable as a path: unlike the zero Path it
// can be ranged over, and it yields no segments.
func TestEmpty(t *testing.T) {
	n := 0
	for range Empty {
		n++
	}
	if n != 0 {
		t.Errorf("the empty path yielded %d segments", n)
	}
}

func TestTightBBox(t *testing.T) {
	big := math.Ldexp(1, 600)
	cases := []struct {
		name string
		p    *Data
		want rect.Rect
	}{
		{
			"quadratic",
			(&Data{}).MoveTo(vec.Vec2{}).QuadTo(vec.Vec2{X: 50, Y: 100}, vec.Vec2{X: 100}),
			rect.Rect{URx: 100, URy: 50},
		},
		{
			"cubic",
			(&Data{}).MoveTo(vec.Vec2{}).CubeTo(vec.Vec2{X: 10, Y: 40}, vec.Vec2{X: 40, Y: 40}, vec.Vec2{X: 50}),
			rect.Rect{URx: 50, URy: 30},
		},
		{
			"lines and a lone move",
			(&Data{}).MoveTo(vec.Vec2{X: 1, Y: 2}).LineTo(vec.Vec2{X: 3, Y: -1}).Close().MoveTo(vec.Vec2{X: -5, Y: 7}),
			rect.Rect{LLx: -5, LLy: -1, URx: 3, URy: 7},
		},
		{
			// the curve starts where the closed sub-path started
			"curve after close",
			(&Data{}).MoveTo(vec.Vec2{}).LineTo(vec.Vec2{X: 10, Y: 10}).Close().
				CubeTo(vec.Vec2{X: 20}, vec.Vec2{X: 20}, vec.Vec2{}),
			rect.Rect{URx: 15, URy: 10},
		},
		{
			// the current point starts at the origin
			"line without a move",
			(&Data{}).LineTo(vec.Vec2{X: 5, Y: 5}),
			rect.Rect{URx: 5, URy: 5},
		},
		{
			"curve without a move",
			(&Data{}).QuadTo(vec.Vec2{X: 2, Y: 4}, vec.Vec2{X: 4}),
			rect.Rect{URx: 4, URy: 2},
		},
		{
			"close without a move",
			(&Data{}).Close().LineTo(vec.Vec2{X: -3, Y: 1}),
			rect.Rect{LLx: -3, URy: 1},
		},
		{
			"huge",
			(&Data{}).MoveTo(vec.Vec2{}).QuadTo(vec.Vec2{X: 50 * big, Y: 100 * big}, vec.Vec2{X: 100 * big}),
			rect.Rect{URx: 100 * big, URy: 50 * big},
		},
		{
			"at the float64 limit",
			(&Data{}).MoveTo(vec.Vec2{X: -math.MaxFloat64}).
				QuadTo(vec.Vec2{X: math.MaxFloat64}, vec.Vec2{X: -math.MaxFloat64}),
			rect.Rect{LLx: -math.MaxFloat64},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := c.p.Iter().TightBBox()
			if !ok || got != c.want {
				t.Errorf("got %v, %t, want %v, true", got, ok, c.want)
			}
		})
	}
}

// TestToCubicWithoutMove checks that a curve at the start of a path begins at
// the origin.
func TestToCubicWithoutMove(t *testing.T) {
	p := (&Data{}).QuadTo(vec.Vec2{X: 3, Y: 6}, vec.Vec2{X: 6}).Iter().ToCubic()
	want := []vec.Vec2{{X: 2, Y: 4}, {X: 4, Y: 4}, {X: 6}}
	n := 0
	for cmd, pts := range p {
		n++
		if cmd != CmdCubeTo || !slices.Equal(pts, want) {
			t.Errorf("got %v %v, want %v %v", cmd, pts, CmdCubeTo, want)
		}
	}
	if n != 1 {
		t.Errorf("got %d segments, want 1", n)
	}
}

func TestTightBBoxEmpty(t *testing.T) {
	for _, p := range []Path{Empty, (&Data{}).Close().Iter()} {
		if _, ok := p.TightBBox(); ok {
			t.Error("a path without points has a box")
		}
	}
}

// TestTransformWithoutMove checks that the implicit start of a path at the
// origin is transformed like every other point.
func TestTransformWithoutMove(t *testing.T) {
	type segment struct {
		cmd Command
		pts []vec.Vec2
	}
	M := matrix.Translate(10, 20)
	cases := []struct {
		name string
		p    *Data
		want []segment
	}{
		{
			"line",
			(&Data{}).LineTo(vec.Vec2{X: 5, Y: 5}),
			[]segment{
				{CmdMoveTo, []vec.Vec2{{X: 10, Y: 20}}},
				{CmdLineTo, []vec.Vec2{{X: 15, Y: 25}}},
			},
		},
		{
			"close first",
			(&Data{}).Close().LineTo(vec.Vec2{X: 5, Y: 5}),
			[]segment{
				{CmdClose, nil},
				{CmdMoveTo, []vec.Vec2{{X: 10, Y: 20}}},
				{CmdLineTo, []vec.Vec2{{X: 15, Y: 25}}},
			},
		},
		{
			"move first",
			(&Data{}).MoveTo(vec.Vec2{X: 1, Y: 2}).LineTo(vec.Vec2{X: 5, Y: 5}),
			[]segment{
				{CmdMoveTo, []vec.Vec2{{X: 11, Y: 22}}},
				{CmdLineTo, []vec.Vec2{{X: 15, Y: 25}}},
			},
		},
		{
			"close only",
			(&Data{}).Close(),
			[]segment{{CmdClose, nil}},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got []segment
			for cmd, pts := range c.p.Iter().Transform(M) {
				got = append(got, segment{cmd, slices.Clone(pts)})
			}
			if diff := cmp.Diff(c.want, got, cmp.AllowUnexported(segment{}), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("transform failed (-want +got):\n%s", diff)
			}
		})
	}
}

// TestBBoxWithoutMove checks that the bounding box of a path which begins
// with a drawing command contains the origin, where the path starts.
func TestBBoxWithoutMove(t *testing.T) {
	cases := []struct {
		name string
		p    *Data
		want rect.Rect
	}{
		{"line", (&Data{}).LineTo(vec.Vec2{X: 5, Y: 5}), rect.Rect{URx: 5, URy: 5}},
		{"close first", (&Data{}).Close().LineTo(vec.Vec2{X: -3, Y: 1}), rect.Rect{LLx: -3, URy: 1}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, ok := c.p.Iter().BBox(); !ok || got != c.want {
				t.Errorf("got %v, %t, want %v, true", got, ok, c.want)
			}
		})
	}
}

func TestBBox(t *testing.T) {
	path := func(yield func(Command, []vec.Vec2) bool) {
		if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
			return
		}
		if !yield(CmdLineTo, []vec.Vec2{{X: 10, Y: 0}}) {
			return
		}
		if !yield(CmdLineTo, []vec.Vec2{{X: 10, Y: 5}}) {
			return
		}
		if !yield(CmdLineTo, []vec.Vec2{{X: 0, Y: 5}}) {
			return
		}
		if !yield(CmdClose, nil) {
			return
		}
	}

	bbox, _ := Path(path).BBox()

	expected := rect.Rect{LLx: 0, LLy: 0, URx: 10, URy: 5}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
	}
}

func TestBBoxWithCurves(t *testing.T) {
	path := func(yield func(Command, []vec.Vec2) bool) {
		if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
			return
		}
		if !yield(CmdCubeTo, []vec.Vec2{{X: 5, Y: 10}, {X: 15, Y: 10}, {X: 20, Y: 0}}) {
			return
		}
	}

	bbox, _ := Path(path).BBox()

	expected := rect.Rect{LLx: 0, LLy: 0, URx: 20, URy: 10}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
	}
}

func TestBBoxEmpty(t *testing.T) {
	for _, p := range []Path{Empty, (&Data{}).Close().Iter()} {
		if _, ok := p.BBox(); ok {
			t.Error("a path without points has a box")
		}
	}
}

func TestTransform(t *testing.T) {
	originalPath := func(yield func(Command, []vec.Vec2) bool) {
		if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
			return
		}
		if !yield(CmdLineTo, []vec.Vec2{{X: 1, Y: 0}}) {
			return
		}
		if !yield(CmdLineTo, []vec.Vec2{{X: 1, Y: 1}}) {
			return
		}
	}

	tests := []struct {
		name      string
		transform matrix.Matrix
		expected  []struct {
			cmd Command
			pts []vec.Vec2
		}
	}{
		{
			name:      "identity transform",
			transform: matrix.Matrix{1, 0, 0, 1, 0, 0},
			expected: []struct {
				cmd Command
				pts []vec.Vec2
			}{
				{CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}},
				{CmdLineTo, []vec.Vec2{{X: 1, Y: 0}}},
				{CmdLineTo, []vec.Vec2{{X: 1, Y: 1}}},
			},
		},
		{
			name:      "translate by (2,3)",
			transform: matrix.Matrix{1, 0, 0, 1, 2, 3},
			expected: []struct {
				cmd Command
				pts []vec.Vec2
			}{
				{CmdMoveTo, []vec.Vec2{{X: 2, Y: 3}}},
				{CmdLineTo, []vec.Vec2{{X: 3, Y: 3}}},
				{CmdLineTo, []vec.Vec2{{X: 3, Y: 4}}},
			},
		},
		{
			name:      "scale by 2",
			transform: matrix.Matrix{2, 0, 0, 2, 0, 0},
			expected: []struct {
				cmd Command
				pts []vec.Vec2
			}{
				{CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}},
				{CmdLineTo, []vec.Vec2{{X: 2, Y: 0}}},
				{CmdLineTo, []vec.Vec2{{X: 2, Y: 2}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transformedPath := Path(originalPath).Transform(tt.transform)

			var results []struct {
				cmd Command
				pts []vec.Vec2
			}

			for cmd, pts := range transformedPath {
				ptsCopy := make([]vec.Vec2, len(pts))
				copy(ptsCopy, pts)
				results = append(results, struct {
					cmd Command
					pts []vec.Vec2
				}{cmd, ptsCopy})
			}

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d commands, got %d", len(tt.expected), len(results))
			}

			for i, expected := range tt.expected {
				if results[i].cmd != expected.cmd {
					t.Errorf("command %d: expected %v, got %v", i, expected.cmd, results[i].cmd)
				}
				if len(results[i].pts) != len(expected.pts) {
					t.Errorf("command %d: expected %d points, got %d", i, len(expected.pts), len(results[i].pts))
					continue
				}
				for j, expectedPt := range expected.pts {
					if results[i].pts[j] != expectedPt {
						t.Errorf("command %d, point %d: expected %v, got %v", i, j, expectedPt, results[i].pts[j])
					}
				}
			}
		})
	}
}

func TestToCubic(t *testing.T) {
	t.Run("convert quadratic to cubic", func(t *testing.T) {
		quadPath := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
				return
			}
			if !yield(CmdQuadTo, []vec.Vec2{{X: 1, Y: 1}, {X: 2, Y: 0}}) {
				return
			}
		}

		cubicPath := Path(quadPath).ToCubic()

		var results []struct {
			cmd Command
			pts []vec.Vec2
		}

		for cmd, pts := range cubicPath {
			results = append(results, struct {
				cmd Command
				pts []vec.Vec2
			}{cmd, pts})
		}

		if len(results) != 2 {
			t.Fatalf("expected 2 commands, got %d", len(results))
		}

		if results[0].cmd != CmdMoveTo {
			t.Errorf("expected MoveTo, got %v", results[0].cmd)
		}

		if results[1].cmd != CmdCubeTo {
			t.Errorf("expected CubeTo, got %v", results[1].cmd)
		}

		if len(results[1].pts) != 3 {
			t.Errorf("expected 3 points for CubeTo, got %d", len(results[1].pts))
		}
	})

	t.Run("leave cubic curves unchanged", func(t *testing.T) {
		cubicPath := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
				return
			}
			if !yield(CmdCubeTo, []vec.Vec2{{X: 1, Y: 1}, {X: 2, Y: 1}, {X: 3, Y: 0}}) {
				return
			}
		}

		convertedPath := Path(cubicPath).ToCubic()

		var originalResults, convertedResults []struct {
			cmd Command
			pts []vec.Vec2
		}

		for cmd, pts := range cubicPath {
			originalResults = append(originalResults, struct {
				cmd Command
				pts []vec.Vec2
			}{cmd, pts})
		}

		for cmd, pts := range convertedPath {
			convertedResults = append(convertedResults, struct {
				cmd Command
				pts []vec.Vec2
			}{cmd, pts})
		}

		if len(originalResults) != len(convertedResults) {
			t.Fatalf("expected same number of commands: original=%d, converted=%d", len(originalResults), len(convertedResults))
		}

		for i := range originalResults {
			if originalResults[i].cmd != convertedResults[i].cmd {
				t.Errorf("command %d differs: original=%v, converted=%v", i, originalResults[i].cmd, convertedResults[i].cmd)
			}

			if len(originalResults[i].pts) != len(convertedResults[i].pts) {
				t.Errorf("command %d point count differs: original=%d, converted=%d", i, len(originalResults[i].pts), len(convertedResults[i].pts))
				continue
			}

			for j := range originalResults[i].pts {
				if originalResults[i].pts[j] != convertedResults[i].pts[j] {
					t.Errorf("command %d, point %d differs: original=%v, converted=%v", i, j, originalResults[i].pts[j], convertedResults[i].pts[j])
				}
			}
		}
	})

	t.Run("handle close command", func(t *testing.T) {
		pathWithClose := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
				return
			}
			if !yield(CmdLineTo, []vec.Vec2{{X: 1, Y: 0}}) {
				return
			}
			if !yield(CmdClose, nil) {
				return
			}
		}

		cubicPath := Path(pathWithClose).ToCubic()

		var results []struct {
			cmd Command
			pts []vec.Vec2
		}

		for cmd, pts := range cubicPath {
			results = append(results, struct {
				cmd Command
				pts []vec.Vec2
			}{cmd, pts})
		}

		if len(results) != 3 {
			t.Fatalf("expected 3 commands, got %d", len(results))
		}

		if results[2].cmd != CmdClose {
			t.Errorf("expected Close, got %v", results[2].cmd)
		}

		if len(results[2].pts) != 0 {
			t.Errorf("expected no points for Close, got %v", results[2].pts)
		}
	})
}

func TestBBoxMoreCases(t *testing.T) {
	t.Run("path with quadratic curve", func(t *testing.T) {
		path := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
				return
			}
			if !yield(CmdQuadTo, []vec.Vec2{{X: 5, Y: 10}, {X: 10, Y: 0}}) {
				return
			}
		}

		bbox, _ := Path(path).BBox()

		expected := rect.Rect{LLx: 0, LLy: 0, URx: 10, URy: 10}
		if bbox != expected {
			t.Errorf("BBox() = %v, want %v", bbox, expected)
		}
	})

	t.Run("path with mixed curve types", func(t *testing.T) {
		path := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 1, Y: 1}}) {
				return
			}
			if !yield(CmdLineTo, []vec.Vec2{{X: 3, Y: 1}}) {
				return
			}
			if !yield(CmdQuadTo, []vec.Vec2{{X: 5, Y: 3}, {X: 7, Y: 1}}) {
				return
			}
			if !yield(CmdCubeTo, []vec.Vec2{{X: 8, Y: 0}, {X: 9, Y: 4}, {X: 10, Y: 2}}) {
				return
			}
		}

		bbox, _ := Path(path).BBox()

		expected := rect.Rect{LLx: 1, LLy: 0, URx: 10, URy: 4}
		if bbox != expected {
			t.Errorf("BBox() = %v, want %v", bbox, expected)
		}
	})

	t.Run("path with close command", func(t *testing.T) {
		path := func(yield func(Command, []vec.Vec2) bool) {
			if !yield(CmdMoveTo, []vec.Vec2{{X: 2, Y: 3}}) {
				return
			}
			if !yield(CmdLineTo, []vec.Vec2{{X: 5, Y: 3}}) {
				return
			}
			if !yield(CmdLineTo, []vec.Vec2{{X: 5, Y: 6}}) {
				return
			}
			if !yield(CmdClose, nil) {
				return
			}
		}

		bbox, _ := Path(path).BBox()

		expected := rect.Rect{LLx: 2, LLy: 3, URx: 5, URy: 6}
		if bbox != expected {
			t.Errorf("BBox() = %v, want %v", bbox, expected)
		}
	})
}

func TestNumPoints(t *testing.T) {
	tests := []struct {
		cmd      Command
		expected int
	}{
		{CmdMoveTo, 1},
		{CmdLineTo, 1},
		{CmdQuadTo, 2},
		{CmdCubeTo, 3},
		{CmdClose, 0},
		{Command(99), 0}, // unknown command
	}

	for _, tt := range tests {
		if got := tt.cmd.NumPoints(); got != tt.expected {
			t.Errorf("Command(%d).NumPoints() = %d, want %d", tt.cmd, got, tt.expected)
		}
	}
}

func TestDataBuilder(t *testing.T) {
	d := &Data{}
	d.MoveTo(vec.Vec2{X: 0, Y: 0}).
		LineTo(vec.Vec2{X: 10, Y: 0}).
		QuadTo(vec.Vec2{X: 15, Y: 5}, vec.Vec2{X: 10, Y: 10}).
		CubeTo(vec.Vec2{X: 5, Y: 15}, vec.Vec2{X: 0, Y: 15}, vec.Vec2{X: 0, Y: 10}).
		Close()

	expectedCmds := []Command{CmdMoveTo, CmdLineTo, CmdQuadTo, CmdCubeTo, CmdClose}
	if len(d.Cmds) != len(expectedCmds) {
		t.Fatalf("expected %d commands, got %d", len(expectedCmds), len(d.Cmds))
	}
	for i, cmd := range expectedCmds {
		if d.Cmds[i] != cmd {
			t.Errorf("Cmds[%d] = %v, want %v", i, d.Cmds[i], cmd)
		}
	}

	// 1 + 1 + 2 + 3 = 7 coordinates
	expectedCoordCount := 7
	if len(d.Coords) != expectedCoordCount {
		t.Errorf("expected %d coordinates, got %d", expectedCoordCount, len(d.Coords))
	}
}

func TestDataIter(t *testing.T) {
	d := &Data{}
	d.MoveTo(vec.Vec2{X: 0, Y: 0}).
		LineTo(vec.Vec2{X: 10, Y: 0}).
		LineTo(vec.Vec2{X: 10, Y: 10}).
		Close()

	var results []struct {
		cmd Command
		pts []vec.Vec2
	}

	for cmd, pts := range d.Iter() {
		ptsCopy := make([]vec.Vec2, len(pts))
		copy(ptsCopy, pts)
		results = append(results, struct {
			cmd Command
			pts []vec.Vec2
		}{cmd, ptsCopy})
	}

	expected := []struct {
		cmd Command
		pts []vec.Vec2
	}{
		{CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}},
		{CmdLineTo, []vec.Vec2{{X: 10, Y: 0}}},
		{CmdLineTo, []vec.Vec2{{X: 10, Y: 10}}},
		{CmdClose, []vec.Vec2{}},
	}

	if len(results) != len(expected) {
		t.Fatalf("expected %d commands, got %d", len(expected), len(results))
	}

	for i, exp := range expected {
		if results[i].cmd != exp.cmd {
			t.Errorf("command %d: expected %v, got %v", i, exp.cmd, results[i].cmd)
		}
		if len(results[i].pts) != len(exp.pts) {
			t.Errorf("command %d: expected %d points, got %d", i, len(exp.pts), len(results[i].pts))
			continue
		}
		for j, pt := range exp.pts {
			if results[i].pts[j] != pt {
				t.Errorf("command %d, point %d: expected %v, got %v", i, j, pt, results[i].pts[j])
			}
		}
	}
}

func TestDataIterChaining(t *testing.T) {
	d := &Data{}
	d.MoveTo(vec.Vec2{X: 0, Y: 0}).
		LineTo(vec.Vec2{X: 5, Y: 0}).
		LineTo(vec.Vec2{X: 5, Y: 5}).
		Close()

	// Test chaining: Data.Iter().Transform().BBox()
	transform := matrix.Matrix{2, 0, 0, 2, 10, 10} // scale by 2, translate by (10,10)
	bbox, _ := d.Iter().Transform(transform).BBox()

	// Original: (0,0), (5,0), (5,5)
	// After transform: (10,10), (20,10), (20,20)
	expected := rect.Rect{LLx: 10, LLy: 10, URx: 20, URy: 20}
	if bbox != expected {
		t.Errorf("chained BBox() = %v, want %v", bbox, expected)
	}
}
