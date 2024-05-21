// Provides core abstract buidling blocks for views:
// AbstractView, View
//
// Also provides basic views:
// Text, Spacer, Button, TextField
package fwsui

import (
	"unicode/utf8"

	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

// Represents abstract view and it's key attributes such as poisition, minimal
// possible size, maximal possible size, actual size.
// FWS suggests universal structure of all views containing key attributes for
// later constraint solving.
// Each view structure must contain AbstractView via composition for providing
// universal access to view attributes.
//
// MinSize and MaxSize will be used for constraing solving stage during which
// inequalities like MinSize.Width <= width <= MaxSize.Width and
// MinSize.Height <= height <= MaxSize.Height will be solves.
// Results will be stored at ActualSize variable
//
// View naming convention: each new view must be names as <View>Object e.g.
// Text will be called TextObject.
// Each view must have constructor function names as <View> e.g.
// Text constructor will be called Text()
//
// All getters and setters of AbstractView are public and named with upper case letter.
// However setters do not return any value and though cannot be chained.
// This is done due to Golang does not provide true inheritance.
// It uses composition, so AbstractView instance will be returned.
// If chaining needed please define new chaining method.
type AbstractView struct {
	position   Point   // Represents abstract view position in parent's coordinates. Assigned by parent object. Default always (0, 0)
	minSize    Size    // Represents abstract view minimal size
	maxSize    Size    // Represents abstract view maximal size
	actualSize Size    // Represents abstract view actual size that will be assigned at constraint solving stage
	gesture    Gesture // Gesture assigned to view
}

// Returns AbstractView position
func (av AbstractView) GetPosition() Point {
	return av.position
}

// Sets AbstractView position
func (av *AbstractView) SetPosition(position Point) {
	av.position = position
}

// Returns AbstractView minimal size
func (av AbstractView) GetMinSize() Size {
	return av.minSize
}

// Returns AbstractView maxinal size
func (av AbstractView) GetMaxSize() Size {
	return av.maxSize
}

// Returns AbstractView actual size
func (av AbstractView) GetActualSize() Size {
	return av.actualSize
}

func (av AbstractView) IsFixedSize() bool {
	return av.minSize.IsEqual(av.maxSize)
}

// Sets AbstractView exact size (MinSize = MaxSize = size)
func (av *AbstractView) SetSize(size Size) {
	av.minSize = size
	av.maxSize = size
}

// Sets AbstractView minimal size
func (av *AbstractView) SetMinSize(size Size) {
	// Increase max size if it is low
	if size.Height > av.maxSize.Height {
		av.maxSize.Height = size.Height
	}
	if size.Width > av.maxSize.Width {
		av.maxSize.Width = size.Width
	}
	av.minSize = size
}

// Sets AbstractView maximal size
func (av *AbstractView) SetMaxSize(size Size) {
	// Reduce min size if it is low
	if size.Height < av.minSize.Height {
		av.minSize.Height = size.Height
	}
	if size.Width < av.minSize.Width {
		av.minSize.Width = size.Width
	}
	av.maxSize = size
}

// Sets AbstractView maximal size
func (av *AbstractView) SetActualSize(size Size) {
	av.actualSize = size
}

// Returns true if View contains binded Gesture
func (av AbstractView) HasGesture() bool {
	return !(av.gesture == nil)
}

// Returns AV's Gesture
func (av AbstractView) GetGesture() Gesture {
	return av.gesture
}

// Sets AV's Gesture
func (av *AbstractView) SetGesture(gesture Gesture) {
	av.gesture = gesture
}

// Returns AV's Frame representation
func (av AbstractView) GetFrame() *Frame {
	frame := new(Frame)
	frame.UpperLeft = av.position
	frame.SetSize(av.actualSize)
	return frame
}

/******************************************************************************/

// View – interface for implementing UI elements that can be rendered in scene.
// Requires Render method that returns proto.Cell matrix of the element.
// Each view must implement those methods
type View interface {
	// Position methods
	GetPosition() Point         // Returns AbstractView position
	SetPosition(position Point) // Sets AbstractView position

	// Size methods
	GetMinSize() Size    // Returns AbstractView minimal size
	GetMaxSize() Size    // Returns AbstractView maxinal size
	GetActualSize() Size // Returns AbstractView actual size
	GetFrame() *Frame    // Returns Frame of AbstractView
	IsFixedSize() bool   // Returns true if min size equal to max size (size is fixed)

	SetSize(size Size)       // Sets AbstractView exact size (MinSize = MaxSize = size)
	SetMinSize(size Size)    // Sets AbstractView minimal size
	SetMaxSize(size Size)    // Sets AbstractView maximal size
	SetActualSize(size Size) // Sets AbstractView actual size

	// Gesture methods
	HasGesture() bool    // Returns true if view have gesture
	GetGesture() Gesture // Returns view's gesture

	SetGesture(gesture Gesture) // Sets view's gesture

	Render() (Canvas, error) // Renders view to canvas
}

/******************************************************************************/

// Provides simple rectangle abstraction with specified size and color
type RectangleObject struct {
	AbstractView
	color proto.Color
}

// Creates simple rectangle abstraction with specified size and color
func Rectangle(size Size) *RectangleObject {
	rect := new(RectangleObject)
	rect.SetPosition(Point{0, 0})
	rect.SetMinSize(size)
	rect.SetGesture(nil)
	return rect
}

// Sets rectangle color
func (r *RectangleObject) Color(color proto.Color) *RectangleObject {
	r.color = color
	return r
}

// Sets rectangle position
func (r *RectangleObject) Position(position Point) *RectangleObject {
	r.SetPosition(position)
	return r
}

// Sets rectangle size
func (r *RectangleObject) Size(size Size) *RectangleObject {
	r.SetSize(size)
	return r
}

func (r *RectangleObject) MinSize(size Size) *RectangleObject {
	r.SetMinSize(size)
	return r
}

func (r *RectangleObject) MaxSize(size Size) *RectangleObject {
	r.SetMaxSize(size)
	return r
}

func (r RectangleObject) Render() (Canvas, error) {
	canvas, error := AllocateCanvas(r.GetActualSize())
	if error != nil {
		return nil, error
	}
	for x := 0; x < int(r.GetActualSize().Width); x++ {
		for y := 0; y < int(r.GetActualSize().Height); y++ {
			canvas[x][y].Bg = r.color
			canvas[x][y].Fg = r.color
			canvas[x][y].Ch = ' '
		}
	}
	return canvas, nil
}

/******************************************************************************/

// Represents text object. Text object is single string with attributes applied
// to all string and ability to bind gesture.
type TextObject struct {
	// AbstractView composition
	AbstractView

	// Text value
	text string

	// Attributes
	align Align

	bold      bool
	blink     bool
	hidden    bool
	dim       bool
	underline bool
	cursive   bool
	reverse   bool

	foreground proto.Color
	background proto.Color
}

// Sets alignment of text
func (text *TextObject) Align(a Align) *TextObject {
	text.align = a
	return text
}

// Sets if text is bold
func (text *TextObject) Bold(b bool) *TextObject {
	text.bold = b
	return text
}

// Sets if text is blinking
func (text *TextObject) Blink(b bool) *TextObject {
	text.blink = b
	return text
}

// Sets if text is hidden
func (text *TextObject) Hidden(b bool) *TextObject {
	text.hidden = b
	return text
}

// Sets if text is dimmed
func (text *TextObject) Dim(b bool) *TextObject {
	text.dim = b
	return text
}

// Sets if text is underlined
func (text *TextObject) Underline(b bool) *TextObject {
	text.underline = b
	return text
}

// Sets if text is cursive
func (text *TextObject) Cursive(b bool) *TextObject {
	text.cursive = b
	return text
}

// Sets if text is reversed
func (text *TextObject) Reverse(b bool) *TextObject {
	text.reverse = b
	return text
}

// Sets foreground color
func (text *TextObject) Foreground(c proto.Color) *TextObject {
	text.foreground = c
	return text
}

// Sets background color
func (text *TextObject) Background(c proto.Color) *TextObject {
	text.background = c
	return text
}

// Constructs attributes superposition
func (text *TextObject) constructAttribute() proto.Attr {
	var attr proto.Attr = 0
	if text.bold {
		attr = attr | proto.Attr(termbox.AttrBold)
	}
	if text.blink {
		attr = attr | proto.Attr(termbox.AttrBlink)
	}
	if text.hidden {
		attr = attr | proto.Attr(termbox.AttrHidden)
	}
	if text.dim {
		attr = attr | proto.Attr(termbox.AttrDim)
	}
	if text.underline {
		attr = attr | proto.Attr(termbox.AttrUnderline)
	}
	if text.cursive {
		attr = attr | proto.Attr(termbox.AttrCursive)
	}
	if text.reverse {
		attr = attr | proto.Attr(termbox.AttrReverse)
	}

	return attr
}

// Renders text view to canvas
func (text *TextObject) Render() (Canvas, error) {
	canvas, error := AllocateCanvas(text.actualSize)
	if error != nil {
		return nil, error
	}
	width, height := text.actualSize.Unpack()
	for x := 0; x < int(width); x++ {
		for y := 0; y < int(height); y++ {
			canvas[x][y].Ch = rune(" "[0])
			canvas[x][y].Fg = text.foreground
			canvas[x][y].Bg = text.background
			canvas[x][y].Attribute = text.constructAttribute()
		}
	}

	var start_x int
	switch text.align {
	case Left:
		start_x = 0
	case Center:
		start_x = int(width)/2 - utf8.RuneCountInString(text.text)/2
	case Right:
		start_x = int(width) - utf8.RuneCountInString(text.text)
	}

	start_y := height / 2
	for x := start_x; x < min(utf8.RuneCountInString(text.text)+start_x, int(width)); x++ {
		if x >= 0 {
			canvas[x][start_y].Ch = []rune(text.text)[x-start_x]
		}
	}

	return canvas, nil
}

// Chained wrapper of SetPosition
func (to *TextObject) Position(position Point) *TextObject {
	to.SetPosition(position)
	return to
}

// Chained wrapper of SetSize
func (to *TextObject) Size(size Size) *TextObject {
	to.SetSize(size)
	return to
}

// Chained wrapper of SetMinSize
func (to *TextObject) MinSize(size Size) *TextObject {
	to.SetMinSize(size)
	return to
}

// Chained wrapper of SetMaxSize
func (to *TextObject) MaxSize(size Size) *TextObject {
	to.SetMaxSize(size)
	return to
}

// Chained wrapper of SetGesture
func (to *TextObject) Gesture(gesture Gesture) *TextObject {
	to.SetGesture(gesture)
	return to
}

// Sets text value of existing TextObject.
// Note: this is chained public method.
func (to *TextObject) SetText(s string) *TextObject {
	to.text = s
	return to
}

// Creates text object. Text object is single string with attributes applied
// to all string and ability to bind gesture.
// Resulted object will contain provided string s with left alignment and empty
// attributes.
// Default position is (0, 0).
// Default size is (length(s), 1)
// No default gesture provided.
func Text(s string) *TextObject {
	text := new(TextObject)
	text.text = s
	text.align = Left
	text.SetPosition(Point{0, 0})
	text.SetSize(Size{uint(utf8.RuneCountInString(s)), 1})
	text.SetGesture(nil)
	return text
}

/******************************************************************************/

// Represents Spacer - transparent area for filling space between other views
type SpacerObject struct {
	AbstractView
}

func (spacer *SpacerObject) Render() (Canvas, error) {
	return AllocateCanvas(spacer.actualSize)
}

// Chained wrapper of SetPosition
func (so *SpacerObject) Position(position Point) *SpacerObject {
	so.SetPosition(position)
	return so
}

// Chained wrapper of SetSize
func (so *SpacerObject) Size(size Size) *SpacerObject {
	so.SetSize(size)
	return so
}

// Chained wrapper of SetMinSize
func (so *SpacerObject) MinSize(size Size) *SpacerObject {
	so.SetMinSize(size)
	return so
}

// Chained wrapper of SetMaxSize
func (so *SpacerObject) MaxSize(size Size) *SpacerObject {
	so.SetMaxSize(size)
	return so
}

// Creates Spacer.
// By default can grow as big as possible so default max width/height is infinite,
// default min width/height is 0. No default Gesture provided.
func Spacer() *SpacerObject {
	spacer := new(SpacerObject)
	spacer.SetMinSize(Size{0, 0})
	spacer.SetMaxSize(Size{Infinite, Infinite})
	return spacer
}

/******************************************************************************/

// Represents Button object. Button is very similar to Text (even based on it)
// with some predifined outlet and customizable action which will be executed on
// click.
type ButtonObject struct {
	TextObject
	action  func(outlet *ButtonObject)
	pressed bool
}

// Creates Button object. Button is very similar to Text (even based on it)
// with some predifined outlet and customizable action which will be executed on
// click.
// It requires function 'action' with signature `func(outlet *ButtonObject)`
// which can perform arbitrary actions and have access to button object as outlet.
//
// Default size is predifined as (length(s) + 2, 1).
// Defualt font color is white and background is grey.
//
// Button has predefined Gesture that is not recomended to change.
func Button(s string, action func(outlet *ButtonObject)) *ButtonObject {
	button := new(ButtonObject)
	button.text = s
	button.align = Center
	button.SetPosition(Point{0, 0})
	button.SetSize(Size{uint(utf8.RuneCountInString(s) + 2), 1})
	button.action = action
	button.foreground = White
	button.background = Grey

	buttonPressed := func() {
		if button.pressed {
			return
		}
		button.foreground.R /= 2
		button.foreground.G /= 2
		button.foreground.B /= 2

		button.background.R /= 2
		button.background.G /= 2
		button.background.B /= 2
	}

	buttonUnpressed := func() {
		if !button.pressed {
			return
		}
		button.foreground.R *= 2
		button.foreground.G *= 2
		button.foreground.B *= 2

		button.background.R *= 2
		button.background.G *= 2
		button.background.B *= 2
	}

	buttonClickGesture := LClickGesture(1).OnChanged(func(inside bool) {
		if inside {
			buttonPressed()
			button.pressed = true
		} else {
			buttonUnpressed()
			button.pressed = false
		}
	}).OnEnded(func(inside bool) {
		if inside {
			buttonUnpressed()
			button.pressed = false
			action(button)
		} else {
			buttonUnpressed()
			button.pressed = false
		}
	})

	button.Gesture(buttonClickGesture)

	return button
}

/******************************************************************************/

// Provides TextField object. This is basic editable text field based on TextObject.
type TextFieldObject struct {
	AbstractView
	resultText  *string
	input       chan *proto.EventRequest
	prompt      string
	active      bool
	typeIndex   int
	selectIndex int
	onFinish    func()
	label       TextObject
}

func (textfield *TextFieldObject) enableInput() {
	AppInstance().setInput(&textfield.input)
}

func (textfield *TextFieldObject) insertString(s string) {
	leftI := textfield.typeIndex
	rightI := textfield.selectIndex
	runeForm := []rune(*textfield.resultText)
	newRuneForm := append(append(runeForm[:leftI], []rune(s)...), runeForm[rightI:]...)
	*textfield.resultText = string(newRuneForm)
	textfield.typeIndex += utf8.RuneCountInString(s)
	textfield.selectIndex = textfield.typeIndex
}

func (textfield *TextFieldObject) deletePartOfString() {
	leftI := textfield.typeIndex
	rightI := textfield.selectIndex
	if leftI != rightI {
		runeForm := []rune(*textfield.resultText)
		newRuneForm := append(runeForm[:leftI], runeForm[rightI:]...)
		*textfield.resultText = string(newRuneForm)
		if textfield.typeIndex > 0 {
			textfield.typeIndex -= 1
		} else {
			textfield.typeIndex = 0
		}
		textfield.selectIndex = textfield.typeIndex
	} else {
		if leftI == 0 {
			return
		}
		runeForm := []rune(*textfield.resultText)
		newRuneForm := append(runeForm[:leftI-1], runeForm[leftI:]...)
		*textfield.resultText = string(newRuneForm)
		textfield.typeIndex -= 1
		textfield.selectIndex = textfield.typeIndex
	}
}

func (textfield *TextFieldObject) handleEvent() {
	for textfield.active {
		event := <-textfield.input
		if event.Ch == 0 {
			switch event.Key {
			case termbox.KeyEnter:
				textfield.active = false
				textfield.deactivate()
				textfield.onFinish()
			case termbox.KeyEsc:
				textfield.active = false
				textfield.deactivate()
			case termbox.KeySpace:
				textfield.insertString(" ")
				textfield.label.SetText(*textfield.resultText).SetSize(Size{Infinite, 1})
			case termbox.KeyArrowLeft:
				if event.Mod != termbox.ModAlt {
					if textfield.selectIndex > 0 {
						textfield.selectIndex -= 1
					}
				} else {
					if textfield.typeIndex > 0 {
						textfield.typeIndex -= 1
					}
				}
			case termbox.KeyArrowRight:
				if event.Mod != termbox.ModAlt {
					if textfield.selectIndex < utf8.RuneCountInString(*textfield.resultText) {
						textfield.selectIndex += 1
					}
				} else {
					if textfield.typeIndex < utf8.RuneCountInString(*textfield.resultText) {
						textfield.typeIndex += 1
					}
				}
			case termbox.KeyBackspace, termbox.KeyBackspace2:
				textfield.deletePartOfString()
				textfield.updateLabelView()
			}
		} else {
			textfield.insertString(string(event.Ch))
			textfield.updateLabelView()
		}
	}
}

func (textfield *TextFieldObject) activate() {
	textfield.active = true
	if utf8.RuneCountInString(*textfield.resultText) == 0 {
		textfield.label.Foreground(Black).SetText("").SetSize(Size{Infinite, 1})
	}
	textfield.typeIndex = 0
	textfield.selectIndex = 0
	go textfield.handleEvent()
}

func (textfield *TextFieldObject) updateLabelView() {
	realWidth, _ := textfield.label.GetActualSize().Unpack()
	if utf8.RuneCountInString(*textfield.resultText) > int(realWidth) {
		runeForm := []rune(*textfield.resultText)[realWidth:]
		textfield.label.SetText(string(runeForm)).SetSize(Size{Infinite, 1})
	} else {
		textfield.label.SetText(*textfield.resultText).SetSize(Size{Infinite, 1})
	}
}

func (textfield *TextFieldObject) deactivate() {
	textfield.active = false
	if utf8.RuneCountInString(*textfield.resultText) == 0 {
		textfield.label.SetText(textfield.prompt).Foreground(Grey).SetSize(Size{Infinite, 1})
	}
}

func (textfield *TextFieldObject) OnFinish(action func()) *TextFieldObject {
	textfield.onFinish = action
	return textfield
}

func TextField(text *string, prompt string) *TextFieldObject {
	textfield := new(TextFieldObject)
	textfield.input = make(chan *proto.EventRequest)
	textfield.label.Background(LightGrey)
	textfield.label.Foreground(Grey)
	textfield.label.SetText(prompt)
	textfield.prompt = prompt
	textfield.resultText = text
	textfield.label.align = Left
	textfield.label.SetPosition(Point{0, 0})
	textfield.label.SetSize(Size{Infinite, 1})
	textfield.label.SetGesture(nil)
	textfield.active = false
	textfield.onFinish = func() {}

	selectGesture := DragGesture().OnChanged(func(value Value) {
		if !textfield.active {
			textfield.active = true
			textfield.activate()
			textfield.enableInput()
			go textfield.handleEvent()
		}
		sel1 := min(max(0, value.startLocationX-textfield.label.GetPosition().X), utf8.RuneCountInString(*textfield.resultText))
		sel2 := min(max(0, value.locationX-textfield.label.GetPosition().X), utf8.RuneCountInString(*textfield.resultText))
		textfield.typeIndex = min(sel1, sel2)
		textfield.selectIndex = max(sel1, sel2)
	}).OnEnded(func(value Value) {

	})
	textfield.label.Gesture(selectGesture)
	return textfield
}

func (textfield *TextFieldObject) Render() (Canvas, error) {
	renderedView, error := textfield.label.Render()
	if error != nil {
		return nil, error
	}
	if textfield.active {
		if textfield.typeIndex != textfield.selectIndex {
			for i := textfield.typeIndex; i < textfield.selectIndex; i++ {
				renderedView[i][0].Bg = Blue
				renderedView[i][0].Fg = White
			}
		} else {
			renderedView[textfield.typeIndex][0].Ch = []rune("|")[0]
		}
	}
	return renderedView, nil
}

// Chained wrapper of SetPosition
func (tfo *TextFieldObject) Position(position Point) *TextFieldObject {
	tfo.SetPosition(position)
	return tfo
}

// Chained wrapper of SetSize
func (tfo *TextFieldObject) Size(size Size) *TextFieldObject {
	tfo.SetSize(size)
	return tfo
}

// Chained wrapper of SetMinSize
func (tfo *TextFieldObject) MinSize(size Size) *TextFieldObject {
	tfo.SetMinSize(size)
	return tfo
}

// Chained wrapper of SetMaxSize
func (tfo *TextFieldObject) MaxSize(size Size) *TextFieldObject {
	tfo.SetMaxSize(size)
	return tfo
}
