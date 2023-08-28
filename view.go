package fwsui

import (
	"unicode/utf8"

	proto "github.com/Nekhaevalex/fwsprotocol"

	"github.com/nsf/termbox-go"
)

// View – interface for implementing UI elements that can be rendered in scene.
// Requires Render method that returns proto.Cell matrix of the element.
type View interface {
	getLogicalSize() (int, int)
	getActualSize() (int, int)
	getPos() (int, int)
	setPos(x, y int)
	render(width, height int) [][]proto.Cell
	hasGesture() bool
	getGesture() Gesture
}

type _Text struct {
	// Position and value
	x      int
	y      int
	width  int
	height int
	text   string
	// Attributes
	align      Align
	bold       bool
	blink      bool
	hidden     bool
	dim        bool
	underline  bool
	cursive    bool
	reverse    bool
	foreground proto.Color
	background proto.Color
	// Gesture part
	gestureFlag     bool
	gesture         Gesture
	awidth, aheight int // actual width & height
}

// getGesture implements View.
func (text *_Text) getGesture() Gesture {
	text.gesture.setParentViewSizes(text)
	return text.gesture
}

func (text *_Text) Align(a Align) *_Text {
	text.align = a
	return text
}

func (text *_Text) Bold(b bool) *_Text {
	text.bold = b
	return text
}

func (text *_Text) Blink(b bool) *_Text {
	text.blink = b
	return text
}

func (text *_Text) Hidden(b bool) *_Text {
	text.hidden = b
	return text
}

func (text *_Text) Dim(b bool) *_Text {
	text.dim = b
	return text
}

func (text *_Text) Underline(b bool) *_Text {
	text.underline = b
	return text
}

func (text *_Text) Cursive(b bool) *_Text {
	text.cursive = b
	return text
}

func (text *_Text) Reverse(b bool) *_Text {
	text.reverse = b
	return text
}

func (text *_Text) Foreground(c proto.Color) *_Text {
	text.foreground = c
	return text
}

func (text *_Text) Background(c proto.Color) *_Text {
	text.background = c
	return text
}

func (text *_Text) constructAttribute() proto.Attr {
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

func (text *_Text) SetSize(w, h int) *_Text {
	text.width = w
	text.height = h
	return text
}

func (text *_Text) SetText(s string) *_Text {
	text.text = s
	text.width = utf8.RuneCountInString(s)
	return text
}

func (text *_Text) getLogicalSize() (int, int) {
	return text.width, text.height
}

func (text *_Text) getActualSize() (int, int) {
	if text.width > 0 && text.height > 0 {
		return text.getLogicalSize()
	}
	return text.awidth, text.aheight
}

func (text *_Text) setPos(x, y int) {
	text.x = x
	text.y = y
}

func (text *_Text) getPos() (int, int) {
	return text.x, text.y
}

func (text *_Text) render(width, height int) [][]proto.Cell {
	text.awidth = width
	text.aheight = height
	canvas := allocateCanvas(width, height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
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
		start_x = width/2 - utf8.RuneCountInString(text.text)/2
	case Right:
		start_x = width - utf8.RuneCountInString(text.text)
	}

	start_y := height / 2
	for x := start_x; x < min(utf8.RuneCountInString(text.text)+start_x, width); x++ {
		if x >= 0 {
			canvas[x][start_y].Ch = []rune(text.text)[x-start_x]
		}
	}

	return canvas
}

func (text *_Text) hasGesture() bool {
	return text.gestureFlag
}

func (text *_Text) Gesture(gesture Gesture) *_Text {
	text.gestureFlag = true
	text.gesture = gesture
	return text
}

func Text(s string) *_Text {
	text := new(_Text)
	text.text = s
	text.align = Left
	text.x = 0
	text.y = 0
	text.width = utf8.RuneCountInString(s)
	text.height = 1
	text.gestureFlag = false
	return text
}

type _Spacer struct {
	x, y, width, height int
}

func (spacer *_Spacer) getLogicalSize() (int, int) {
	return spacer.width, spacer.height
}

func (spacer *_Spacer) getActualSize() (int, int) {
	return spacer.width, spacer.height
}

func (spacer *_Spacer) setPos(x, y int) {
	spacer.x = x
	spacer.y = y
}

func (spacer *_Spacer) getPos() (int, int) {
	return spacer.x, spacer.y
}

func (spacer *_Spacer) render(width, height int) [][]proto.Cell {
	return allocateCanvas(width, height)
}

func (spacer *_Spacer) SetSize(w, h int) *_Spacer {
	spacer.width = w
	spacer.height = h
	return spacer
}

func (spacer *_Spacer) getGesture() Gesture {
	return nil
}

func (spacer *_Spacer) hasGesture() bool {
	return false
}

func Spacer() *_Spacer {
	spacer := new(_Spacer)
	spacer.width = -1
	spacer.height = -1
	return spacer
}

type _Button struct {
	_Text
	action  func(outlet *_Button)
	pressed bool
}

func Button(s string, action func(outlet *_Button)) *_Button {
	button := new(_Button)
	button.text = s
	button.align = Center
	button.x = 0
	button.y = 0
	button.width = utf8.RuneCountInString(s) + 2
	button.height = 1
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

type _TextField struct {
	resultText  *string
	input       chan *proto.EventRequest
	prompt      string
	active      bool
	typeIndex   int
	selectIndex int
	onFinish    func()
	label       _Text
}

func (textfield *_TextField) enableInput() {
	AppInstance().setInput(&textfield.input)
}

func (textfield *_TextField) insertString(s string) {
	leftI := textfield.typeIndex
	rightI := textfield.selectIndex
	runeForm := []rune(*textfield.resultText)
	newRuneForm := append(append(runeForm[:leftI], []rune(s)...), runeForm[rightI:]...)
	*textfield.resultText = string(newRuneForm)
	textfield.typeIndex += utf8.RuneCountInString(s)
	textfield.selectIndex = textfield.typeIndex
}

func (textfield *_TextField) deletePartOfString() {
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

func (textfield *_TextField) handleEvent() {
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
				textfield.label.SetText(*textfield.resultText).SetSize(-1, 1)
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

func (textfield *_TextField) activate() {
	textfield.active = true
	if utf8.RuneCountInString(*textfield.resultText) == 0 {
		textfield.label.Foreground(Black).SetText("").SetSize(-1, 1)
	}
	textfield.typeIndex = 0
	textfield.selectIndex = 0
	go textfield.handleEvent()
}

func (textfield *_TextField) updateLabelView() {
	realWidth, _ := textfield.label.getActualSize()
	if utf8.RuneCountInString(*textfield.resultText) > realWidth {
		runeForm := []rune(*textfield.resultText)[realWidth:]
		textfield.label.SetText(string(runeForm)).SetSize(-1, 1)
	} else {
		textfield.label.SetText(*textfield.resultText).SetSize(-1, 1)
	}
}

func (textfield *_TextField) deactivate() {
	textfield.active = false
	if utf8.RuneCountInString(*textfield.resultText) == 0 {
		textfield.label.SetText(textfield.prompt).Foreground(Grey).SetSize(-1, 1)
	}
}

func (textfield *_TextField) OnFinish(action func()) *_TextField {
	textfield.onFinish = action
	return textfield
}

func TextField(text *string, prompt string) *_TextField {
	textfield := new(_TextField)
	textfield.input = make(chan *proto.EventRequest)
	textfield.label.Background(LightGrey)
	textfield.label.Foreground(Grey)
	textfield.label.SetText(prompt)
	textfield.prompt = prompt
	textfield.resultText = text
	textfield.label.align = Left
	textfield.label.x = 0
	textfield.label.y = 0
	textfield.label.width = -1
	textfield.label.height = 1
	textfield.label.gestureFlag = false
	textfield.active = false
	textfield.onFinish = func() {}

	selectGesture := DragGesture().OnChanged(func(value Value) {
		if !textfield.active {
			textfield.active = true
			textfield.activate()
			textfield.enableInput()
			go textfield.handleEvent()
		}
		sel1 := min(max(0, value.startLocationX-textfield.label.x), utf8.RuneCountInString(*textfield.resultText))
		sel2 := min(max(0, value.locationX-textfield.label.x), utf8.RuneCountInString(*textfield.resultText))
		textfield.typeIndex = min(sel1, sel2)
		textfield.selectIndex = max(sel1, sel2)
	}).OnEnded(func(value Value) {

	})
	textfield.label.Gesture(selectGesture)
	return textfield
}

func (textfield *_TextField) getLogicalSize() (int, int) {
	return textfield.label.getLogicalSize()
}
func (textfield *_TextField) getActualSize() (int, int) {
	return textfield.label.getActualSize()
}
func (textfield *_TextField) getPos() (int, int) {
	return textfield.label.getPos()
}
func (textfield *_TextField) setPos(x, y int) {
	textfield.label.setPos(x, y)
}
func (textfield *_TextField) render(width, height int) [][]proto.Cell {
	renderedView := textfield.label.render(width, height)
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
	return renderedView
}
func (textfield *_TextField) hasGesture() bool {
	return textfield.label.hasGesture()
}
func (textfield *_TextField) getGesture() Gesture {
	return textfield.label.getGesture()
}
