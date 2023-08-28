package fwsui

import (
	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

// Some useful local functions

func allocateCanvas(width, height int) [][]proto.Cell {
	canvas := make([][]proto.Cell, width)
	for i := 0; i < width; i++ {
		canvas[i] = make([]proto.Cell, height)
	}
	return canvas
}

func viewSizeFloating(v View) (bool, bool) {
	size_x, size_y := v.getLogicalSize()
	can_x, can_y := false, false
	if size_x < 0 {
		can_x = true
	}
	if size_y < 0 {
		can_y = true
	}
	return can_x, can_y
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func pointInArea(x, y int, area GestureDescriptor) bool {
	if x >= area.x && y >= area.y {
		if x < (area.x+area.width) && y < (area.y+area.height) {
			return true
		}
	}
	return false
}

// Some popular colors

var White = proto.Color{A: 255, R: 255, G: 255, B: 255}
var Black = proto.Color{A: 255, R: 0, G: 0, B: 0}
var Red = proto.Color{A: 255, R: 255, G: 0, B: 0}
var Grey = proto.Color{A: 255, R: 127, G: 127, B: 127}
var LightGrey = proto.Color{A: 255, R: 192, G: 192, B: 192}
var Yellow = proto.Color{A: 255, R: 255, G: 255, B: 0}
var Green = proto.Color{A: 255, R: 0, G: 255, B: 0}
var Blue = proto.Color{A: 255, R: 0, G: 0, B: 255}

type Align uint8

const (
	Left Align = iota
	Center
	Right
)

type prevGesture struct {
	key   termbox.Key
	x, y  int
	actor Gesture
}

func (prev *prevGesture) isSameObject(event *proto.EventRequest) bool {
	prevCont := (prev.key == termbox.MouseLeft) || (prev.key == termbox.MouseMiddle) || (prev.key == termbox.MouseRight)
	newCont := (event.Key == termbox.MouseLeft) || (event.Key == termbox.MouseMiddle) || (event.Key == termbox.MouseRight) || (event.Key == termbox.MouseRelease)
	return prevCont && newCont
}

func (prev *prevGesture) save(event *proto.EventRequest, actor Gesture) {
	prev.key = event.Key
	prev.x = event.MouseX
	prev.y = event.MouseY
	prev.actor = actor
}
