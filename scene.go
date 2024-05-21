package fwsui

import (
	"log"

	proto "github.com/Nekhaevalex/fwsprotocol"
	"github.com/nsf/termbox-go"
)

// Scene – interface for implementing standalone objects that can be shown on
// screen and handle incomming events
type Scene interface {
	bindApp(app *_App)                         // Method for saving pointer of App instance
	requestLayerId() proto.ID                  // Method for requesting new layer ID from Window Server
	getEventChannel() chan *proto.EventRequest // Method for returning events incomming connection
	buildContent()                             // Method for building contained views
	eventHandler()                             // Handler for incomming events
}

type WindowObject struct {
	// Main values
	x, y, width, height int
	app                 *_App
	layerId             proto.ID
	title               string
	activeAreas         []GestureDescriptor
	events              chan *proto.EventRequest
	quit                chan int
	background          proto.Color
	body                View
	windowContainer     *AbstractStackObject
	staticCanvas        Canvas
	prevMouse           prevGesture
	lastX, lastY        int
	lastW, lastH        int
	onCloseFunc         func()
	titleText           *TextObject
}

func (window *WindowObject) Close() {
	delete(window.app.scenes, window.layerId)
	delete_request := &proto.DeleteRequest{Id: window.layerId}
	window.app.sendRequest(delete_request)
	window.onCloseFunc()
	window.quit <- 1
}

func (window *WindowObject) OnClose(closeFunc func()) *WindowObject {
	window.onCloseFunc = closeFunc
	return window
}

func (window *WindowObject) SetSize(width, height int) *WindowObject {
	window.width = width
	window.height = height
	return window
}

func (window *WindowObject) SetTitle(s string) *WindowObject {
	window.title = s
	if window.windowContainer != nil {
		window.titleText.SetText(s)
	}
	return window
}

func (window *WindowObject) bindApp(app *_App) {
	window.app = app
}

func (window *WindowObject) requestLayerId() proto.ID {
	// Construct initial window creations request
	new_window_request := &proto.NewWindowRequest{
		Pid:    window.app.pid,
		X:      window.x,
		Y:      window.y,
		Width:  window.width,
		Height: window.height,
	}
	window.layerId = window.app.sendRequest(new_window_request)
	return window.layerId
}

func (window *WindowObject) getEventChannel() chan *proto.EventRequest {
	return window.events
}

func (window *WindowObject) moveWindow(translation Vector) {
	moveRequest := &proto.MoveRequest{
		Id: window.layerId,
		X:  translation.X - window.lastX,
		Y:  translation.Y - window.lastY,
	}
	window.app.sendRequest(moveRequest)
	render := &proto.RenderRequest{Id: window.layerId}
	window.app.sendRequest(render)
	// window.lastX = translationX
	// window.lastY = translationY
	window.lastX = translation.X
	window.lastY = translation.Y
}

func (window *WindowObject) resizeWindow(translation Vector) {
	resulsW := window.width + translation.X - window.lastW
	resulsH := window.height + translation.Y - window.lastH
	if resulsW > 15 && resulsH > 5 {
		window.width = resulsW
		window.height = resulsH
		resizeRequest := &proto.ResizeRequest{
			Id:     window.layerId,
			Width:  window.width,
			Height: window.height,
		}
		window.app.sendRequest(resizeRequest)
	}
	window.lastW = translation.X
	window.lastH = translation.Y
}

func (window *WindowObject) buildContent() {
	// Move gesture
	windowMoveGesture := DragGesture().OnChanged(func(value Value) {
		window.moveWindow(value.translation)
	}).OnEnded(func(value Value) {
		window.lastX = 0
		window.lastY = 0
	})

	shadowColor := Black
	shadowColor.A = 127
	shadowRect := Text("").MaxSize(Size{Infinite, Infinite}).Background(shadowColor).Foreground(shadowColor)

	shadowLayer := VStack(
		Spacer().MaxSize(Size{Infinite, 1}),
		HStack(
			Spacer().MaxSize(Size{2, Infinite}),
			shadowRect,
		),
	)
	resizeGesture := DragGesture().OnChanged(func(value Value) {
		window.resizeWindow(value.translation)
	}).OnEnded(func(value Value) {
		window.lastW = 0
		window.lastH = 0
	})

	window.titleText = Text(window.title).Foreground(White).Background(Grey).Align(Center).MaxSize(Size{Infinite, Infinite}).Gesture(windowMoveGesture)

	windowFrame := VStack(
		HStack(
			Button("X", func(outlet *ButtonObject) {
				window.Close()
			}).Foreground(White).Background(Red),
			Button("-", func(outlet *ButtonObject) {
				// Todo
			}).Foreground(Grey).Background(Yellow),
			Button("+", func(outlet *ButtonObject) {
				// Todo
			}).Foreground(White).Background(Green),
			window.titleText,
		).MaxSize(Size{Infinite, 1}),
		ZStack(
			Text("").Background(White).Foreground(White).MaxSize(Size{Infinite, Infinite}),
			window.body,
			Box(Text("⇲").Background(White).Foreground(Black).Gesture(resizeGesture)).Gravity(Gravity{Right, Right}).MaxSize(Size{Infinite, Infinite}),
		))

	realLayer := VStack(
		HStack(
			windowFrame,
			Spacer().MaxSize(Size{2, Infinite}),
		),
		Spacer().MaxSize(Size{Infinite, 1}),
	)

	// Window view
	window.windowContainer = ZStack(shadowLayer, realLayer)
	window.redraw()
}

func (window *WindowObject) redraw() {
	var err error
	window.staticCanvas, err = window.Render()
	log.Fatal(err)
	draw_request := &proto.DrawFillRequest{
		Id:     window.layerId,
		Width:  window.width,
		Height: window.height,
		Img:    window.staticCanvas,
	}
	window.app.sendRequest(draw_request)
	render_request := &proto.RenderRequest{Id: window.layerId}
	window.app.sendRequest(render_request)
	window.activeAreas = make([]GestureDescriptor, 0)
	window.activeAreas = append(window.activeAreas, window.windowContainer.GetChildrenGestures()...)
}

func (window *WindowObject) getGestureInPoint(x, y int) Gesture {
	for i := len(window.activeAreas) - 1; i >= 0; i-- {
		area := window.activeAreas[i]
		if area.PointInArea(Point{x, y}) {
			return area.Pointer
		}
	}
	return nil
}

func (window *WindowObject) eventHandler() {
	window.activeAreas = append(window.activeAreas, window.windowContainer.GetChildrenGestures()...)
	for {
		select {
		case event := <-window.events:
			switch event.Type {
			case termbox.EventMouse:
				x := event.MouseX
				y := event.MouseY
				//Experimental!!!
				var actor Gesture
				if !window.prevMouse.isSameObject(event) {
					actor = window.getGestureInPoint(x, y)
				} else {
					actor = window.prevMouse.actor
				}
				window.prevMouse.save(event, actor)
				// [Experimental]
				if actor != nil {
					actor.updating(event)
					window.redraw()
				}
			case termbox.EventKey:
				if *window.app.keyInputChan != nil {
					*window.app.keyInputChan <- event
				}
				window.redraw()
			}
		case <-window.quit:
			return
		}
	}
}

func (window *WindowObject) getLogicalSize() (int, int) {
	return window.width, window.height
}

func (window *WindowObject) getActualSize() (int, int) {
	return window.width, window.height
}

func (window *WindowObject) getPos() (int, int) {
	return window.x, window.y
}

func (window *WindowObject) getGesture() Gesture {
	return nil
}

func (window *WindowObject) hasGesture() bool {
	return false
}

func (window *WindowObject) setPos(x, y int) {
	window.x = x
	window.y = y
}

func (window *WindowObject) Render() (Canvas, error) {
	window.windowContainer.SetPosition(Point{0, 0})
	return window.windowContainer.Render()
}

func Window(title string, body View) *WindowObject {
	window := new(WindowObject)
	window.x = 5
	window.y = 5
	window.width = 50
	window.height = 18
	window.title = title
	window.body = body
	window.events = make(chan *proto.EventRequest)
	window.quit = make(chan int)
	window.background = proto.Color{A: 255, R: 255, G: 255, B: 255}
	window.activeAreas = make([]GestureDescriptor, 0)
	window.onCloseFunc = func() {}
	return window
}
