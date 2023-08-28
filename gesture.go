package fwsui

import (
	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

type GestureDescriptor struct {
	x, y, width, height int
	pointer             Gesture
}

// Gesture – interface used for implementing interactive elements like
// buttons, drag areas, etc.
type Gesture interface {
	getGestureDescriptor(x, y int) GestureDescriptor
	setParentViewSizes(v View)
	updating(event *proto.EventRequest)
	onChanged()
	onEnded()
	setAltGesture(gesture Gesture)
}

// Clicks
type _AClickGesture struct {
	x, y, width, height          int
	mouseButton                  termbox.Key
	count, current_count         int
	descriptor                   GestureDescriptor
	inside                       bool
	buttonMatched                bool
	action_changed, action_ended func(inside bool)
	altGesture                   Gesture
}

func (click *_AClickGesture) setParentViewSizes(v View) {
	click.x, click.y = v.getPos()
	click.width, click.height = v.getActualSize()
}

func (click *_AClickGesture) getGestureDescriptor(x, y int) GestureDescriptor {
	descriptor := GestureDescriptor{
		x:       click.x + x,
		y:       click.y + y,
		width:   click.width,
		height:  click.height,
		pointer: click,
	}
	click.descriptor = descriptor
	return descriptor
}

func (click *_AClickGesture) updating(event *proto.EventRequest) {
	switch event.Key {
	case click.mouseButton:
		click.buttonMatched = true
		if pointInArea(event.MouseX, event.MouseY, click.descriptor) {
			click.inside = true
		} else {
			click.inside = false
			click.current_count = 0
		}
		click.onChanged()
	case termbox.MouseRelease:
		if click.buttonMatched {
			click.buttonMatched = false
			if pointInArea(event.MouseX, event.MouseY, click.descriptor) {
				click.inside = true
			} else {
				click.inside = false
			}
			click.onEnded()
		} else {
			if click.altGesture != nil {
				click.altGesture.updating(event)
			}
		}
	default:
		if click.altGesture != nil {
			click.altGesture.updating(event)
		}
	}
}

func (click *_AClickGesture) onChanged() {
	click.action_changed(click.inside)
}

func (click *_AClickGesture) onEnded() {
	click.current_count++
	if click.current_count == click.count {
		click.action_ended(click.inside)
		click.current_count = 0
	}
}

func (click *_AClickGesture) setAltGesture(gesture Gesture) {
	click.altGesture = gesture
}

func (click *_AClickGesture) OnChanged(action func(inside bool)) *_AClickGesture {
	click.action_changed = action
	return click
}

func (click *_AClickGesture) OnEnded(action func(inside bool)) *_AClickGesture {
	click.action_ended = action
	return click
}

func AClickGesture(button termbox.Key, count int) *_AClickGesture {
	gesture := new(_AClickGesture)
	gesture.mouseButton = button
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func LClickGesture(count int) *_AClickGesture {
	gesture := new(_AClickGesture)
	gesture.mouseButton = termbox.MouseLeft
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func RClickGesture(count int) *_AClickGesture {
	gesture := new(_AClickGesture)
	gesture.mouseButton = termbox.MouseRight
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func MClickGesture(count int) *_AClickGesture {
	gesture := new(_AClickGesture)
	gesture.mouseButton = termbox.MouseMiddle
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

type Value struct {
	startLocationX, startLocationY int
	locationX, locationY           int
	translationX, translationY     int
}

type _DragGesture struct {
	x, y, width, height          int
	descriptor                   GestureDescriptor
	value                        Value
	action_changed, action_ended func(value Value)
	in_process                   bool
	buttonMatched                bool
	altGesture                   Gesture
}

func (drag *_DragGesture) getGestureDescriptor(x, y int) GestureDescriptor {
	descriptor := GestureDescriptor{
		x:       drag.x + x,
		y:       drag.y + y,
		width:   drag.width,
		height:  drag.height,
		pointer: drag,
	}
	drag.descriptor = descriptor
	return descriptor
}
func (drag *_DragGesture) setParentViewSizes(v View) {
	drag.x, drag.y = v.getPos()
	drag.width, drag.height = v.getActualSize()
}
func (drag *_DragGesture) updating(event *proto.EventRequest) {
	switch event.Key {
	case termbox.MouseLeft:
		drag.buttonMatched = true
		if !drag.in_process {
			drag.value.startLocationX = event.MouseX
			drag.value.startLocationY = event.MouseY
			drag.in_process = true
		}
		drag.value.locationX = event.MouseX
		drag.value.locationY = event.MouseY
		drag.value.translationX = drag.value.locationX - drag.value.startLocationX
		drag.value.translationY = drag.value.locationY - drag.value.startLocationY
		drag.onChanged()
	case termbox.MouseRelease:
		if drag.buttonMatched {
			drag.buttonMatched = false
			drag.in_process = false
			drag.value.locationX = event.MouseX
			drag.value.locationY = event.MouseY
			drag.value.translationX = drag.value.locationX - drag.value.startLocationX
			drag.value.translationY = drag.value.locationY - drag.value.startLocationY
			drag.onEnded()
		} else {
			if drag.altGesture != nil {
				drag.altGesture.updating(event)
			}
		}
	default:
		if drag.altGesture != nil {
			drag.altGesture.updating(event)
		}
	}
}
func (drag *_DragGesture) onChanged() {
	if drag.action_changed != nil {
		drag.action_changed(drag.value)
	}
}
func (drag *_DragGesture) onEnded() {
	if drag.action_ended != nil {
		drag.action_ended(drag.value)
	}
}
func (drag *_DragGesture) setAltGesture(gesture Gesture) {
	drag.altGesture = gesture
}

func (drag *_DragGesture) OnChanged(action func(value Value)) *_DragGesture {
	drag.action_changed = action
	return drag
}

func (drag *_DragGesture) OnEnded(action func(value Value)) *_DragGesture {
	drag.action_ended = action
	return drag
}

func DragGesture() *_DragGesture {
	drag := new(_DragGesture)
	drag.in_process = false
	return drag
}
