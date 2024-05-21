package fwsui

import (
	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

type GestureDescriptor struct {
	Position Point
	Size     Size
	Pointer  Gesture
}

func (gd GestureDescriptor) PointInArea(point Point) bool {
	if point.X >= gd.Position.X && point.Y >= gd.Position.Y {
		if point.X < (gd.Position.X+int(gd.Size.Width)) && point.Y < (gd.Position.Y+int(gd.Size.Height)) {
			return true
		}
	}
	return false
}

// Gesture – interface used for implementing interactive elements like
// buttons, drag areas, etc.
type Gesture interface {
	GetGestureDescriptor() GestureDescriptor
	setParentViewSizes(v View)
	updating(event *proto.EventRequest)
	onChanged()
	onEnded()
	setAltGesture(gesture Gesture)
}

// Clicks
type AClickGestureObject struct {
	Position                     Point
	Size                         Size
	mouseButton                  termbox.Key
	count, current_count         int
	descriptor                   GestureDescriptor
	inside                       bool
	buttonMatched                bool
	action_changed, action_ended func(inside bool)
	altGesture                   Gesture
}

func (click *AClickGestureObject) setParentViewSizes(v View) {
	click.Position = v.GetPosition()
	click.Size = v.GetActualSize()
}

func (click *AClickGestureObject) GetGestureDescriptor() GestureDescriptor {
	descriptor := GestureDescriptor{
		Position: click.Position,
		Size:     click.Size,
		Pointer:  click,
	}
	click.descriptor = descriptor
	return descriptor
}

func (click *AClickGestureObject) updating(event *proto.EventRequest) {
	switch event.Key {
	case click.mouseButton:
		click.buttonMatched = true
		if click.descriptor.PointInArea(Point{event.MouseX, event.MouseY}) {
			click.inside = true
		} else {
			click.inside = false
			click.current_count = 0
		}
		click.onChanged()
	case termbox.MouseRelease:
		if click.buttonMatched {
			click.buttonMatched = false
			if click.descriptor.PointInArea(Point{event.MouseX, event.MouseY}) {
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

func (click *AClickGestureObject) onChanged() {
	click.action_changed(click.inside)
}

func (click *AClickGestureObject) onEnded() {
	click.current_count++
	if click.current_count == click.count {
		click.action_ended(click.inside)
		click.current_count = 0
	}
}

func (click *AClickGestureObject) setAltGesture(gesture Gesture) {
	click.altGesture = gesture
}

func (click *AClickGestureObject) OnChanged(action func(inside bool)) *AClickGestureObject {
	click.action_changed = action
	return click
}

func (click *AClickGestureObject) OnEnded(action func(inside bool)) *AClickGestureObject {
	click.action_ended = action
	return click
}

func AClickGesture(button termbox.Key, count int) *AClickGestureObject {
	gesture := new(AClickGestureObject)
	gesture.mouseButton = button
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func LClickGesture(count int) *AClickGestureObject {
	gesture := new(AClickGestureObject)
	gesture.mouseButton = termbox.MouseLeft
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func RClickGesture(count int) *AClickGestureObject {
	gesture := new(AClickGestureObject)
	gesture.mouseButton = termbox.MouseRight
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

func MClickGesture(count int) *AClickGestureObject {
	gesture := new(AClickGestureObject)
	gesture.mouseButton = termbox.MouseMiddle
	gesture.count = count
	gesture.current_count = 0
	return gesture
}

type Value struct {
	startPosition Point
	location      Point
	translation   Vector
}

type DragGestureObject struct {
	Position                     Point
	Size                         Size
	descriptor                   GestureDescriptor
	value                        Value
	action_changed, action_ended func(value Value)
	in_process                   bool
	buttonMatched                bool
	altGesture                   Gesture
}

func (drag *DragGestureObject) GetGestureDescriptor() GestureDescriptor {
	descriptor := GestureDescriptor{
		Position: drag.Position,
		Size:     drag.Size,
		Pointer:  drag,
	}
	drag.descriptor = descriptor
	return descriptor
}
func (drag *DragGestureObject) setParentViewSizes(v View) {
	drag.Position = v.GetPosition()
	drag.Size = v.GetActualSize()
}
func (drag *DragGestureObject) updating(event *proto.EventRequest) {
	switch event.Key {
	case termbox.MouseLeft:
		drag.buttonMatched = true
		if !drag.in_process {
			drag.value.startPosition = Point{event.MouseX, event.MouseY}
			drag.in_process = true
		}
		drag.value.location = Point{event.MouseX, event.MouseY}
		drag.value.translation = Vector(drag.value.location).Sub(Vector(drag.value.startPosition))
		drag.onChanged()
	case termbox.MouseRelease:
		if drag.buttonMatched {
			drag.buttonMatched = false
			drag.in_process = false
			drag.value.location = Point{event.MouseX, event.MouseY}
			drag.value.translation = Vector(drag.value.location).Sub(Vector(drag.value.startPosition))
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
func (drag *DragGestureObject) onChanged() {
	if drag.action_changed != nil {
		drag.action_changed(drag.value)
	}
}
func (drag *DragGestureObject) onEnded() {
	if drag.action_ended != nil {
		drag.action_ended(drag.value)
	}
}
func (drag *DragGestureObject) setAltGesture(gesture Gesture) {
	drag.altGesture = gesture
}

func (drag *DragGestureObject) OnChanged(action func(value Value)) *DragGestureObject {
	drag.action_changed = action
	return drag
}

func (drag *DragGestureObject) OnEnded(action func(value Value)) *DragGestureObject {
	drag.action_ended = action
	return drag
}

func DragGesture() *DragGestureObject {
	drag := new(DragGestureObject)
	drag.in_process = false
	return drag
}
