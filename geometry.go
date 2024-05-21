/*
	Geometry provides basic abstractions used in UI library such as
	points, sizes, vectors, axis
*/

package fwsui

import "math"

/******************************************************************************/

// Infinite - constatnt representing infinity as maximal possible integer
const Infinite = math.MaxInt

// Axis - abstraction for representing 3D axis
type Axis uint

const (
	X Axis = iota
	Y
	Z
)

// Returns orthogonal axis (X -> Y, Y -> X, Z -> Z)
func (ax Axis) PlaneOrthogonal() Axis {
	switch ax {
	case X:
		return Y
	case Y:
		return X
	default:
		return Z
	}
}

/******************************************************************************/

// Point - abstract primitive that represents 2D point
type Point struct {
	X, Y int // Cartesian coordinates
}

// Equal - returns true if points' components are equal
func (point Point) Equal(other Point) bool {
	return point.X == other.X && point.Y == other.Y
}

// Translate - translates point with vector and returns it's pointer
func (point *Point) Translate(vector Vector) *Point {
	point.X += vector.X
	point.Y += vector.Y
	return point
}

// Unpack - returns Point value as tuple
func (p Point) Unpack() (int, int) {
	return p.X, p.Y
}

// Gets point component by axis
func (p Point) GetComponent(axis Axis) int {
	switch axis {
	case X:
		return p.X
	case Y:
		return p.Y
	case Z:
		return 0
	}
	return 0
}

// Sets point component by axis
func (p *Point) SetComponent(value int, axis Axis) *Point {
	switch axis {
	case X:
		p.X = value
	case Y:
		p.Y = value
	}
	return p
}

/******************************************************************************/

// Vector - abstract primitive that represents 2D vector
type Vector Point

// Gets vector component by axis
func (v Vector) GetComponent(axis Axis) int {
	switch axis {
	case X:
		return v.X
	case Y:
		return v.Y
	case Z:
		return 0
	}
	return 0
}

// Sets vector component by axis
func (v *Vector) SetComponent(value int, axis Axis) *Vector {
	switch axis {
	case X:
		v.X = value
	case Y:
		v.Y = value
	}
	return v
}

func (v Vector) Equal(vector Vector) bool {
	return v.ToPoint().Equal(vector.ToPoint())
}

// Add - returns sum of two vectors
func (vec Vector) Add(other Vector) Vector {
	return Vector{vec.X + other.X, vec.Y + other.Y}
}

// Add - returns substraction of two vectors
func (vec Vector) Sub(other Vector) Vector {
	return Vector{vec.X - other.X, vec.Y - other.Y}
}

// Mul - returns result of multiplication of integer number and vector
func (vec Vector) Mul(a int) Vector {
	return Vector{vec.X * a, vec.Y * a}
}

// VecMul - returns result of vector multiplication
func (vec Vector) VecMul(other Vector) Vector {
	return Vector{vec.X * other.X, vec.Y * other.Y}
}

// Reverse - returns reversed vector
func (vec Vector) Reverse() Vector {
	return vec.Mul(-1)
}

// Dot - returns dot product of two vectors
func (vec Vector) Dot(other Vector) int {
	return vec.X*other.X + vec.Y*other.Y
}

// Norm - returns Euclidian norm of vector
func (vec Vector) Norm() float64 {
	return math.Sqrt(float64(vec.X*vec.X + vec.Y*vec.Y))
}

// Project - returns vector projected on axis
// If axis == Z returns (0, 0)
func (vec Vector) Project(axis Axis) Vector {
	switch axis {
	case X, Y:
		return *vec.SetComponent(0, axis.PlaneOrthogonal())
	default:
		return Vector{0, 0}
	}
}

// ToSize - converts vector to Size object
// Note: size components cannot be negative. Negative components will be converted to positive.
func (vec Vector) ToSize() Size {
	return Size{uint(abs(vec.X)), uint(abs(vec.Y))}
}

// ToPoint - converts vector to Point object
func (vec Vector) ToPoint() Point {
	return Point(vec)
}

/******************************************************************************/

// Represents 2D size attribute
type Size struct {
	Width, Height uint
}

// Returns true if width is infinite
func (s Size) InfiniteWidth() bool {
	return s.Width == Infinite
}

// Returns true if height is infinite
func (s Size) InfiniteHeight() bool {
	return s.Height == Infinite
}

// Returns true if width and height are infinite
func (s Size) IsInfinite() (bool, bool) {
	return s.InfiniteWidth(), s.InfiniteHeight()
}

func (s Size) Equal(other Size) bool {
	return s.Width == other.Width && s.Height == other.Height
}

// Gets size component by axis
func (s Size) GetComponent(axis Axis) uint {
	switch axis {
	case X:
		return s.Width
	case Y:
		return s.Height
	case Z:
		return 0
	}
	return 0
}

// Sets size component by axis
func (s *Size) SetComponent(value uint, axis Axis) *Size {
	switch axis {
	case X:
		s.Width = value
	case Y:
		s.Height = value
	}
	return s
}

// Converts size to Vector object
func (s Size) ToVector() Vector {
	return Vector{int(s.Width), int(s.Height)}
}

// Returns Size value as tuple of (width, height)
func (s Size) Unpack() (uint, uint) {
	return s.Width, s.Height
}

/******************************************************************************/

// Frame - represents object for calculations on rectangles
type Frame struct {
	UpperLeft  Point
	LowerRight Point
}

// Returns size
func (f Frame) GetSize() Size {
	ul := Vector(f.UpperLeft)
	lr := Vector(f.LowerRight)
	return lr.Sub(ul).ToSize()
}

// Sets size and returns object pointer
func (f *Frame) SetSize(size Size) *Frame {
	f.LowerRight = Point(Vector(f.UpperLeft).Add(size.ToVector()))
	return f
}

// Moves frame and retuns object pointer
func (f *Frame) Move(translationVector Vector) *Frame {
	f.UpperLeft.Translate(translationVector)
	f.LowerRight.Translate(translationVector)
	return f
}

// Returns new frame cutted with other frame
func (f Frame) Cut(with Frame) Frame {
	return Frame{
		Point{max(f.UpperLeft.X, with.UpperLeft.X), max(f.UpperLeft.Y, with.UpperLeft.Y)},
		Point{max(f.LowerRight.X, with.LowerRight.X), max(f.LowerRight.Y, with.LowerRight.Y)},
	}
}

// Returns pair of 1D coordinates of start & end of frame
func (f Frame) GetStartEnd(axis Axis) (int, int) {
	return f.UpperLeft.GetComponent(axis), f.LowerRight.GetComponent(axis)
}
