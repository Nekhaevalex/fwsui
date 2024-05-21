package fwsui

import (
	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

// Some useful local functions

func max[T int | uint | float32 | float64](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func min[T int | uint | float32 | float64](x, y T) T {
	if x > y {
		return y
	}
	return x
}

func abs[T int | float32 | float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
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
