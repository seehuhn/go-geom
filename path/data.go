// Copyright 2026 Jochen Voss <voss@seehuhn.de>
// SPDX-License-Identifier: BSD-3-Clause

package path

import "seehuhn.de/go/geom/vec"

// Data is a compact, mutable representation of a path.
type Data struct {
	Cmds   []Command
	Coords []vec.Vec2
}

// DataFromPath creates a new Data object from the given Path.
func DataFromPath(p Path) *Data {
	data := &Data{}
	for cmd, args := range p {
		data.Cmds = append(data.Cmds, cmd)
		data.Coords = append(data.Coords, args...)
	}
	return data
}

// IsBlank reports whether the path is empty, i.e. whether it contains no
// commands at all.
func (d *Data) IsBlank() bool {
	return len(d.Cmds) == 0
}

// MoveTo starts a new sub-path at the given point.
func (d *Data) MoveTo(p vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdMoveTo)
	d.Coords = append(d.Coords, p)
	return d
}

// LineTo adds a line segment to the given point.
func (d *Data) LineTo(p vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdLineTo)
	d.Coords = append(d.Coords, p)
	return d
}

// QuadTo adds a quadratic Bezier curve.
func (d *Data) QuadTo(ctrl, end vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdQuadTo)
	d.Coords = append(d.Coords, ctrl, end)
	return d
}

// CubeTo adds a cubic Bezier curve.
func (d *Data) CubeTo(ctrl1, ctrl2, end vec.Vec2) *Data {
	d.Cmds = append(d.Cmds, CmdCubeTo)
	d.Coords = append(d.Coords, ctrl1, ctrl2, end)
	return d
}

// Close closes the current sub-path.
func (d *Data) Close() *Data {
	d.Cmds = append(d.Cmds, CmdClose)
	return d
}

// Iter returns a Path iterator over the path data.
func (d *Data) Iter() Path {
	return func(yield func(Command, []vec.Vec2) bool) {
		i := 0
		for _, cmd := range d.Cmds {
			n := cmd.NumPoints()
			if !yield(cmd, d.Coords[i:i+n]) {
				return
			}
			i += n
		}
	}
}
