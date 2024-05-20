/*
Describes Canvas abstraction and it's core functions
*/
package fwsui

import (
	"errors"

	proto "github.com/Nekhaevalex/fwsprotocol"
)

// Represents canvas - 2D matrix of cells.
// Cell is defined in fwsprotocol package.
type Canvas [][]proto.Cell

// Returns canvas size
func (c Canvas) Size() Size {
	if len(c) != 0 {
		w := len(c)
		h := len(c[0])
		return Size{w, h}
	}
	return Size{0, 0}
}

func AllocateCanvas(size Size) (Canvas, error) {
	if size.Width == 0 || size.Height == 0 {
		return nil, errors.New("at least one dimension of canvas is zero")
	}
	canvas := make([][]proto.Cell, size.Width)
	for i := 0; i < size.Width; i++ {
		canvas[i] = make([]proto.Cell, size.Height)
	}
	return canvas, nil
}
