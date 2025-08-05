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

	"seehuhn.de/go/geom/rect"
	"seehuhn.de/go/geom/vec"
)

func TestBBox(t *testing.T) {
	// Create a simple rectangular path
	path := func(yield func(Command, []vec.Vec2) bool) {
		// Move to origin
		if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
			return
		}
		// Line to (10, 0)
		if !yield(CmdLineTo, []vec.Vec2{{X: 10, Y: 0}}) {
			return
		}
		// Line to (10, 5)
		if !yield(CmdLineTo, []vec.Vec2{{X: 10, Y: 5}}) {
			return
		}
		// Line to (0, 5)
		if !yield(CmdLineTo, []vec.Vec2{{X: 0, Y: 5}}) {
			return
		}
		// Close path
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
	// Create a path with a cubic curve
	path := func(yield func(Command, []vec.Vec2) bool) {
		// Move to origin
		if !yield(CmdMoveTo, []vec.Vec2{{X: 0, Y: 0}}) {
			return
		}
		// Cubic curve with control points at (5, 10) and (15, 10), ending at (20, 0)
		if !yield(CmdCubeTo, []vec.Vec2{{X: 5, Y: 10}, {X: 15, Y: 10}, {X: 20, Y: 0}}) {
			return
		}
	}

	bbox := Path(path).BBox()

	// The bounding box should include all control points and endpoints
	expected := rect.Rect{LLx: 0, LLy: 0, URx: 20, URy: 10}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
	}
}

func TestBBoxEmpty(t *testing.T) {
	// Empty path
	path := func(yield func(Command, []vec.Vec2) bool) {}

	bbox := Path(path).BBox()

	// Empty path should return zero rectangle
	expected := rect.Rect{}
	if bbox != expected {
		t.Errorf("BBox() = %v, want %v", bbox, expected)
	}
}
