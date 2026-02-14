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

package path

import (
	"testing"

	"seehuhn.de/go/geom/matrix"
	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

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

	bbox := Path(path).BBox()

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

	bbox := Path(path).BBox()

	expected := rect.Rect{LLx: 0, LLy: 0, URx: 20, URy: 10}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
	}
}

func TestBBoxEmpty(t *testing.T) {
	path := func(yield func(Command, []vec.Vec2) bool) {}

	bbox := Path(path).BBox()

	expected := rect.Rect{}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
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

		bbox := Path(path).BBox()

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

		bbox := Path(path).BBox()

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

		bbox := Path(path).BBox()

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
	bbox := d.Iter().Transform(transform).BBox()

	// Original: (0,0), (5,0), (5,5)
	// After transform: (10,10), (20,10), (20,20)
	expected := rect.Rect{LLx: 10, LLy: 10, URx: 20, URy: 20}
	if bbox != expected {
		t.Errorf("chained BBox() = %v, want %v", bbox, expected)
	}
}
