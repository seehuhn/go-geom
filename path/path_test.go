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

import "testing"

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
