// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

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
