package fwsui

import (
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

type _Window struct {
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
	windowContainer     *_ZStack
	staticCanvas        [][]proto.Cell
	prevMouse           prevGesture
	lastX, lastY        int
	lastW, lastH        int
	onCloseFunc         func()
	titleText           *_Text
}

func (window *_Window) Close() {
	delete(window.app.scenes, window.layerId)
	delete_request := &proto.DeleteRequest{Id: window.layerId}
	window.app.sendRequest(delete_request)
	window.onCloseFunc()
	window.quit <- 1
}

func (window *_Window) OnClose(closeFunc func()) *_Window {
	window.onCloseFunc = closeFunc
	return window
}

func (window *_Window) SetSize(width, height int) *_Window {
	window.width = width
	window.height = height
	return window
}

func (window *_Window) SetTitle(s string) *_Window {
	window.title = s
	if window.windowContainer != nil {
		window.titleText.SetText(s)
	}
	return window
}

func (window *_Window) bindApp(app *_App) {
	window.app = app
}

func (window *_Window) requestLayerId() proto.ID {
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

func (window *_Window) getEventChannel() chan *proto.EventRequest {
	return window.events
}

func (window *_Window) moveWindow(translationX, translationY int) {
	moveRequest := &proto.MoveRequest{
		Id: window.layerId,
		X:  translationX - window.lastX,
		Y:  translationY - window.lastY,
	}
	window.app.sendRequest(moveRequest)
	render := &proto.RenderRequest{Id: window.layerId}
	window.app.sendRequest(render)
	// window.lastX = translationX
	// window.lastY = translationY
	window.lastX = translationX
	window.lastY = translationY
}

func (window *_Window) resizeWindow(translationX, translationY int) {
	resulsW := window.width + translationX - window.lastW
	resulsH := window.height + translationY - window.lastH
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
	window.lastW = translationX
	window.lastH = translationY
}

func (window *_Window) buildContent() {
	// Move gesture
	windowMoveGesture := DragGesture().OnChanged(func(value Value) {
		window.moveWindow(value.translationX, value.translationY)
	}).OnEnded(func(value Value) {
		window.lastX = 0
		window.lastY = 0
	})

	shadowColor := Black
	shadowColor.A = 127
	shadowRect := Text("").SetSize(-1, -1).Background(shadowColor).Foreground(shadowColor)

	shadowLayer := VStack(
		Spacer().SetSize(-1, 1),
		HStack(
			Spacer().SetSize(2, -1),
			shadowRect,
		),
	)
	resizeGesture := DragGesture().OnChanged(func(value Value) {
		window.resizeWindow(value.translationX, value.translationY)
	}).OnEnded(func(value Value) {
		window.lastW = 0
		window.lastH = 0
	})

	window.titleText = Text(window.title).Foreground(White).Background(Grey).Align(Center).SetSize(-1, -1).Gesture(windowMoveGesture)

	windowFrame := VStack(
		HStack(
			Button("X", func(outlet *_Button) {
				window.Close()
			}).Foreground(White).Background(Red),
			Button("-", func(outlet *_Button) {
				// Todo
			}).Foreground(Grey).Background(Yellow),
			Button("+", func(outlet *_Button) {
				// Todo
			}).Foreground(White).Background(Green),
			window.titleText,
		).SetSize(-1, 1),
		ZStack(
			Text("").Background(White).Foreground(White).SetSize(-1, -1),
			window.body,
			Box(Text("⇲").Background(White).Foreground(Black).Gesture(resizeGesture)).Gravity(Right, Right).SetSize(-1, -1),
		))

	realLayer := VStack(
		HStack(
			windowFrame,
			Spacer().SetSize(2, -1),
		),
		Spacer().SetSize(-1, 1),
	)

	// Window view
	window.windowContainer = ZStack(shadowLayer, realLayer)
	window.redraw()
}

func (window *_Window) redraw() {
	window.staticCanvas = window.render(window.width, window.height)
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
	window.activeAreas = append(window.activeAreas, window.windowContainer.getChildrenGestures(0, 0)...)
}

func (window *_Window) getGestureInPoint(x, y int) Gesture {
	for i := len(window.activeAreas) - 1; i >= 0; i-- {
		area := window.activeAreas[i]
		if pointInArea(x, y, area) {
			return area.pointer
		}
	}
	return nil
}

func (window *_Window) eventHandler() {
	window.activeAreas = append(window.activeAreas, window.windowContainer.getChildrenGestures(0, 0)...)
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

func (window *_Window) getLogicalSize() (int, int) {
	return window.width, window.height
}

func (window *_Window) getActualSize() (int, int) {
	return window.width, window.height
}

func (window *_Window) getPos() (int, int) {
	return window.x, window.y
}

func (window *_Window) getGesture() Gesture {
	return nil
}

func (window *_Window) hasGesture() bool {
	return false
}

func (window *_Window) setPos(x, y int) {
	window.x = x
	window.y = y
}

func (window *_Window) render(width, height int) [][]proto.Cell {
	window.windowContainer.setPos(0, 0)
	return window.windowContainer.render(width, height)
}

func Window(title string, body View) *_Window {
	window := new(_Window)
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
