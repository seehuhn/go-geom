# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
