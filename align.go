/*
Provides Align - universal type for contrainting alignment
*/
package fwsui

// Represents one of 5 possible 2D alignments: left, center, right, top, bottom.
type Align uint8

const (
	Left Align = iota
	Center
	Right
	Top
	Bottom
)

// Transfroms any alignment to horizontal
func (alignment Align) TransformAlignment() Align {
	switch alignment {
	case Top:
		return Left
	case Bottom:
		return Right
	default:
		return alignment
	}
}

// Represents 2D alignment of child object in container
type Gravity struct {
	Horizontal Align // Represents horizontal alignment
	Vertical   Align // Represents vertical alignment
}

// Returns component by axis
func (g Gravity) GetComponent(axis Axis) Align {
	switch axis {
	case X:
		return g.Horizontal
	case Y:
		return g.Vertical
	default:
		return 0
	}
}

// Sets component by axis and returns pointer to gravity object
func (g *Gravity) SetComponent(value Align, axis Axis) *Gravity {
	switch axis {
	case X:
		g.Horizontal = value
	case Y:
		g.Vertical = value
	}
	return g
}
