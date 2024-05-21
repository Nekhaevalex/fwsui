/*
Describes Canvas abstraction and it's core functions
*/
package fwsui

import (
	"errors"
	"fmt"

	proto "github.com/Nekhaevalex/fwsprotocol"
)

// Represents canvas - 2D matrix of cells.
// Cell is defined in fwsprotocol package.
type Canvas [][]proto.Cell

// Sets cell at point
func (c Canvas) Set(point Point, cell proto.Cell) Canvas {
	c[point.X][point.Y] = cell.Over(c[point.X][point.Y])
	return c
}

// Gets point as cell
func (c Canvas) Get(point Point) proto.Cell {
	return c[point.X][point.Y]
}

// Returns canvas size
func (c Canvas) Size() Size {
	if len(c) != 0 {
		w := uint(len(c))
		h := uint(len(c[0]))
		return Size{w, h}
	}
	return Size{0, 0}
}

// Copies content of other canvas with translation.
func (c Canvas) Inpaint(child Canvas, childStart Vector) Canvas {
	// Screen start
	screenStart := Vector{0, 0}
	// Vector to end of canvas
	screenEnd := c.Size().ToVector()
	// Vector from child's zero to it's end
	lChildEnd := child.Size().ToVector()
	// Vector from global zero to child's end
	childEnd := childStart.Add(lChildEnd)

	// Vector to begin of visible area from zero
	visibleStart := Vector{min(max(childStart.X, screenStart.X), screenEnd.X), min(max(childStart.Y, screenStart.Y), screenEnd.Y)}
	// Vector to end of visible area from zero
	visibleEnd := Vector{max(min(childEnd.X, screenEnd.X), screenStart.X), max(min(childEnd.Y, screenEnd.Y), screenStart.Y)}

	for x := visibleStart.X; x < visibleEnd.X; x++ {
		for y := visibleStart.Y; y < visibleEnd.Y; y++ {
			currentPoint := Point{x, y}
			localCurrentPoint := Point(Vector(currentPoint).Sub(childStart))
			c.Set(currentPoint, child.Get(localCurrentPoint))
		}
	}
	return c
}

// Resizes existing canvas without cleaning contents
// If new size is bigger, canvas will be filled with empty cells
// If new size is smaller, image will be cropped
func (c Canvas) Resize(newSize Size) (Canvas, error) {
	newCanvas, err := AllocateCanvas(newSize)
	if err != nil {
		return nil, errors.Join(errors.New("resize failed on allocation"), err)
	}
	newCanvas.Inpaint(c, Vector{0, 0})
	return newCanvas, err
}

func AllocateCanvas(size Size) (Canvas, error) {
	if size.Width == 0 || size.Height == 0 {
		return nil, fmt.Errorf(
			"AllocateCanvas: at least one dimension of canvas is zero: (W: %d, H: %d)\ncheck if ActualSize was assigned",
			size.Width, size.Height)
	}
	canvas := make([][]proto.Cell, size.Width)
	for i := 0; i < int(size.Width); i++ {
		canvas[i] = make([]proto.Cell, size.Height)
	}
	return canvas, nil
}
