package fwsui

import (
	"log"

	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

type MouseEvent struct {
	Position Point
	Button   termbox.Key
	Happened bool
}

func FromEventRequest(event proto.EventRequest) MouseEvent {
	return MouseEvent{Point{event.MouseX, event.MouseY}, event.Key, true}
}

type ActiveArea struct {
	Position Point
	Size     Size
	Gesture  Gesture
}

func (aa ActiveArea) EventInArea(event MouseEvent) bool {
	if event.Position.X >= aa.Position.X && event.Position.Y >= aa.Position.Y {
		if event.Position.X < (aa.Position.X+int(aa.Size.Width)) && event.Position.Y < (aa.Position.Y+int(aa.Size.Height)) {
			return true
		}
	}
	return false
}

type Value struct {
	StartPosition Point
	Location      Point
	Translation   Vector
}

// Gesture – interface used for implementing interactive elements like
// buttons, drag areas, etc.
type Gesture interface {
	// Status of gesture getters
	Started() bool    // Returns true if gesture was started
	Ended() bool      // Returns true if gesture ended
	ForceEnd()        // Force end gesture
	InProgress() bool // Returns true if gesture started but not ended yet
	// Services
	GetGestureDescriptor(view View) ActiveArea // Returns ActiveArea description
	Update(event *proto.EventRequest)          // Called by owner every time new event happens
	// Callbacks
	onChanged() // Called when new event of gesture happens
	onEnded()   // Called when gesture ends
}

type AbstractGesture struct {
	Start, End, Current        MouseEvent
	actionChanged, actionEnded func()
}

func (ag AbstractGesture) Started() bool {
	return ag.Start.Happened
}

func (ag AbstractGesture) Ended() bool {
	return ag.End.Happened
}

func (ag *AbstractGesture) ForceEnd() {
	ag.End = ag.Current
}

func (ag AbstractGesture) InProgress() bool {
	return ag.Started() && !ag.Ended()
}

func (ag *AbstractGesture) GetGestureDescriptor(view View) ActiveArea {
	descriptor := ActiveArea{
		Position: view.GetPosition(),
		Size:     view.GetActualSize(),
		Gesture:  ag,
	}
	return descriptor
}

func (ag AbstractGesture) onChanged() {
	if ag.actionChanged != nil {
		ag.actionChanged()
	} else {
		log.Printf("actionChanged is nil")
	}
}

func (ag AbstractGesture) onEnded() {
	if ag.actionEnded != nil {
		ag.actionEnded()
	} else {
		log.Printf("actionEnded is nil")
	}
}

func (ag *AbstractGesture) Update(event *proto.EventRequest) {
	ag.Current = FromEventRequest(*event)
	if !ag.Started() || ag.Ended() {
		ag.Start = ag.Current
		ag.End = MouseEvent{Point{0, 0}, 0, false}
		ag.onChanged()
	} else if ag.Start.Button != ag.Current.Button && !ag.Ended() {
		ag.End = ag.Current
		ag.onEnded()
	} else {
		ag.onChanged()
	}
}

// Base gestures
type DragGestureObject struct {
	AbstractGesture
}

func DragGesture() *DragGestureObject {
	dragGesture := new(DragGestureObject)
	return dragGesture
}

// Builds action wrapper for drag gesture
func (dg *DragGestureObject) buildGestureFunction(action func(value Value)) func() {
	return func() {
		value := Value{
			StartPosition: dg.Start.Position,
			Location:      dg.Current.Position,
			Translation:   Vector(dg.Current.Position).Sub(Vector(dg.Start.Position)),
		}
		action(value)
	}
}

// Sets action to be done when gesture changes
func (dg *DragGestureObject) OnChanged(action func(value Value)) *DragGestureObject {
	dg.actionChanged = dg.buildGestureFunction(action)
	return dg
}

// Sets action to be done when gesture ends
func (dg *DragGestureObject) OnEnded(action func(value Value)) *DragGestureObject {
	dg.actionEnded = dg.buildGestureFunction(action)
	return dg
}

type AClickGestureObject struct {
	AbstractGesture
	ActiveArea
	Button       termbox.Key
	Count        uint
	clickCounter uint
}

func AClickGesture(button termbox.Key, count uint) *AClickGestureObject {
	gesture := new(AClickGestureObject)
	gesture.Button = button
	gesture.Count = count
	return gesture
}

func (acg *AClickGestureObject) GetGestureDescriptor(view View) ActiveArea {
	acg.Position = view.GetPosition()
	acg.Size = view.GetActualSize()
	acg.Gesture = acg
	return acg.ActiveArea
}

func (acg *AClickGestureObject) OnChanged(action func(inside bool)) *AClickGestureObject {
	acg.actionChanged = func() {
		if acg.EventInArea(acg.Current) && acg.Current.Button == acg.Button {
			action(true)
		} else {
			acg.clickCounter = 0
			action(false)
		}
	}
	return acg
}

func (acg *AClickGestureObject) OnEnded(action func(inside bool)) *AClickGestureObject {
	acg.actionEnded = func() {
		if acg.EventInArea(acg.Current) {
			acg.clickCounter++
			if acg.clickCounter == acg.Count {
				action(true)
				acg.clickCounter = 0
			}
		} else {
			action(false)
		}
	}
	return acg
}

func LClickGesture(count int) *AClickGestureObject {
	return AClickGesture(termbox.MouseLeft, uint(count))
}

func RClickGesture(count int) *AClickGestureObject {
	return AClickGesture(termbox.MouseRight, uint(count))
}

func MClickGesture(count int) *AClickGestureObject {
	return AClickGesture(termbox.MouseMiddle, uint(count))
}
