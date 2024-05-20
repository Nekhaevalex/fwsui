package fwsui

import (
	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

// Some useful local functions

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
