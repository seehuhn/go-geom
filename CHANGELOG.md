# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.7.5] (2026-10-10)

### Added
- `path.StrokeOptions`, with `CapStyle`, `JoinStyle` and
  `DefaultMiterLimit`, and `path.PolylineStrokeBBox` for the exact
  bounding box of a stroked set of polylines, allowing for dashes.
- New `polygon` package: perimeter, signed area and resampling of
  closed polygons.
- `linalg.SegmentIntersection` and `linalg.MiterLength`.
- `path.EvalCubic`, `path.Empty`, `rect.Rect.Grow` and
  `vec.Vec2.Frexp`.

### Changed
- `matrix.Matrix.Inv` returns an additional `ok` result, and `Zero`
  when the matrix is singular, instead of panicking.
- `path.Data.IsBlank` reports whether the path contains no commands at
  all.  A path of only move and close commands is no longer blank.
- `vec.Vec2.Normalize` and `vec.Vec2.Normal` no longer discard vectors
  shorter than 1e-9; only the zero vector has no direction.
- `linalg.Intersection` returns the zero vector when there is no
  answer.

### Fixed
- The geometry primitives in `vec`, `matrix` and `linalg` are accurate
  across the whole float64 range.  Tolerances are relative to the
  quantities involved, and intermediate overflow and underflow are
  avoided.  For example, `Inv` previously turned `Scale(1e200)` into
  `Zero`, and `linalg.Intersection` reported perpendicular lines as
  parallel outside coordinates from 1e-160 to 1e150.
- `path.Data.IsBlank` reported a path made of a move and a quadratic
  curve as blank.

## [v0.7.4] (2026-06-25)

### Added
- `rect.Rect.Transform` maps a rectangle through an affine
  `matrix.Matrix` and returns the axis-aligned bounding box of the
  result.

## [v0.7.3] (2026-05-19)

### Changed
- Minimum Go version raised to 1.25.

## [v0.7.2] (2026-05-11)

No source changes; re-tagged for the v0.7.2 Quire release.

## [v0.7.1] (2026-03-31)

### Changed
- `Matrix.Apply` uses `vec.Vec2` parameter type.
- `path.Transform` uses `matrix.Matrix` type.

## [v0.7.0] (2025-01-25)

### Added
- **New `vec` package** with `Vec2` type for 2D vector operations, including arithmetic, dot/cross products, length, and normalization
- **New `linalg` package** with line intersection utilities
- **New `path.Data` type** for mutable path construction with builder methods (`MoveTo`, `LineTo`, `QuadTo`, `CubeTo`, `Close`)
- **New `rect.IntRect` type** for integer rectangles
- `path.Data.Iter()` method to convert back to `Path` iterator
- `path.Data.IsBlank()` method to check for empty/move-only paths
- `path.DataFromPath()` constructor to create `Data` from a `Path` iterator
- `path.Command.NumPoints()` method returning number of points consumed by command
- `vec.Vec2.Neg()` method for vector negation
- `vec.Vec2.Rot90()` method for 90-degree counter-clockwise rotation
- `vec.Vec2.Normalize()` method returning unit vector (zero if length < 1e-9)
- Convenience methods for chaining matrix transformations

### Changed
- Simplified `path.BBox()` implementation
- Improved test coverage and documentation
