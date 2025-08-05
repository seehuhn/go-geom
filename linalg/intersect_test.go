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
			if ok && intersection != tc.expected {
				t.Errorf("expected intersection=%v, got %v", tc.expected, intersection)
			}
		})
	}
}
