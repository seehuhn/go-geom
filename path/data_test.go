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
	"testing"

	"seehuhn.de/go/geom/vec"
)

func TestIsBlank(t *testing.T) {
	a := vec.Vec2{X: 1, Y: 2}
	b := vec.Vec2{X: 3, Y: 4}
	c := vec.Vec2{X: 5, Y: 6}

	cases := []struct {
		name     string
		data     *Data
		expected bool
	}{
		{"empty", &Data{}, true},
		{"move only", (&Data{}).MoveTo(a), false},
		{"degenerate sub-path", (&Data{}).MoveTo(a).Close(), false},
		{"line", (&Data{}).MoveTo(a).LineTo(b), false},
		{"quadratic curve", (&Data{}).MoveTo(a).QuadTo(b, c), false},
		{"cubic curve", (&Data{}).MoveTo(a).CubeTo(a, b, c), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.data.IsBlank(); got != tc.expected {
				t.Errorf("IsBlank() = %t, want %t", got, tc.expected)
			}
		})
	}
}
