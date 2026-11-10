// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

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
