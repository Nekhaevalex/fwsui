package fwsui

import (
	"context"
	"errors"
	"log"

	proto "github.com/Nekhaevalex/fwsprotocol"
	"github.com/nsf/termbox-go"
)

var (
	ErrorSceneNoCancel  = errors.New("cancel function is nil")
	ErrorSceneNotFound  = errors.New("scene not found in app")
	ErrorSceneCloseFail = errors.New("scene failed to close")
)

// Represents abstract scene object.
// Doesn't implement EventHandler hence doesn't implement Scene interface.
// Needs additional EventHandler definition.
type AbstractScene struct {
	Position       Point                    // Represents Scene frame global position
	LayerID        proto.ID                 // Layer ID on window server
	Content        View                     // Scene contents pointer (via interface)
	CurrentGesture Gesture                  // Current gesture pointer
	Events         chan *proto.EventRequest // Incomming channel for events from WS
	cancel         context.CancelFunc       // Function for canceling EventHandler
}

// Requests Layer ID from window server
func (scene *AbstractScene) RequestLayerID() proto.ID {
	if AppInstance() == nil {
		log.Fatal("AppObject pointer nil")
	}
	// Construct initial window creations request
	newWindowRequest := &proto.NewWindowRequest{
		Pid:    AppInstance().Pid,
		X:      scene.Position.X,
		Y:      scene.Position.Y,
		Width:  int(scene.Content.GetActualSize().Width),
		Height: int(scene.Content.GetActualSize().Height),
	}
	reply, err := AppInstance().SendRequest(newWindowRequest)
	if err != nil {
		newWindowReply, ok := reply.(*proto.ReplyCreationRequest)
		if !ok {
			log.Fatal("unknown messsge received: ", reply)
		}
		scene.LayerID = newWindowReply.Id
	}
	return scene.LayerID
}

// Returns event channel. If doesn't exists, allocates it
func (scene *AbstractScene) GetEventChannel() chan *proto.EventRequest {
	if scene.Events == nil {
		scene.Events = make(chan *proto.EventRequest)
	}
	return scene.Events
}

func (window *AbstractScene) EventHandler(ctx context.Context) {
	for {
		if window.Events == nil {
			log.Fatal("Events channel is nil")
		}
		select {
		case <-ctx.Done():
			return
		case event := <-window.Events:
			switch event.Type {
			case termbox.EventMouse:
				// Event interference prevention!
				// Here we should isolate current event from interfering with other events.

				// First we retrieve received event as MouseEvent (simply for standartization)
				mouseEvent := FromEventRequest(*event)
				// Next we retrieve gesture which is located on this position.
				// Important: in fact it can be any event. Event is retrieved based on it's position.
				// If we retrieved new event while our current event is not finished yet, event interference may happen.

				// Actions to prevent event interference:
				// 0. Check if current gesture is not nil. If nil then assign found gesture.
				// 1. Check if found (retrieved) gesture is the same as current gesture, stored in Scene object.
				//    If pointer are equal, no problem.
				// 2. Check if found gestuire is nill. It means that event didn't trigger any other Gestures.
				// 3. If there is any gesture found which is not equal to current gesture
				//    – check if current gesture ended and if yes - update current gesture.
				foundGesture := window.FindGesture(mouseEvent)
				if window.CurrentGesture != nil {
					if window.CurrentGesture != foundGesture && foundGesture != nil {
						if window.CurrentGesture.Ended() {
							window.CurrentGesture = foundGesture
						}
					}
				} else {
					window.CurrentGesture = foundGesture
				}
				// [Experimental]
				if window.CurrentGesture != nil {
					window.CurrentGesture.Update(event)
					canvas, err := window.Render()
					if err != nil {
						log.Fatal("Failed to rerender on update", err)
					}
					window.SendRender(canvas)
					window.Redraw()
					if window.CurrentGesture.Ended() {
						window.CurrentGesture = nil
					}
				}
			case termbox.EventKey:
				if *AppInstance().keyInputChan != nil {
					*AppInstance().keyInputChan <- event
				}
			}
		}
	}
}

func (scene *AbstractScene) Emit() {
	ctx, cancel := context.WithCancel(context.Background())
	scene.cancel = cancel
	go scene.EventHandler(ctx)
}

func (scene *AbstractScene) Close() error {
	if scene.cancel == nil {
		return ErrorSceneNoCancel
	}
	scene.cancel()
	if _, ok := AppInstance().Scenes[scene.LayerID]; ok {
		delete(AppInstance().Scenes, scene.LayerID)
	} else {
		return ErrorSceneNotFound
	}
	deleteRequest := &proto.DeleteRequest{Id: scene.LayerID}
	_, err := AppInstance().SendRequest(deleteRequest)
	if err != nil {
		return errors.Join(ErrorSceneCloseFail, err)
	}
	return nil
}

// Moves scene
func (scene *AbstractScene) Move(translation Vector) {
	moveRequest := &proto.MoveRequest{
		Id: scene.LayerID,
		X:  translation.X,
		Y:  translation.Y,
	}

	AppInstance().SendRequest(moveRequest)
	render := &proto.RenderRequest{Id: scene.LayerID}
	AppInstance().SendRequest(render)
	scene.Position.Translate(translation)
}

// Resizes scene
func (scene *AbstractScene) Resize(size Size) {
	if size.Width >= 15 && size.Height >= 5 {
		scene.Content.SetActualSize(size)
		resizeRequest := &proto.ResizeRequest{
			Id:     scene.LayerID,
			Width:  int(scene.Content.GetActualSize().Width),
			Height: int(scene.Content.GetActualSize().Height),
		}
		AppInstance().SendRequest(resizeRequest)
		scene.Redraw()
	}
}

func (scene AbstractScene) Redraw() {
	render := &proto.RenderRequest{Id: scene.LayerID}
	AppInstance().SendRequest(render)
}

func (scene AbstractScene) Render() (Canvas, error) {
	canvas, err := scene.Content.Render()
	if err != nil {
		return nil, errors.Join(errors.New("AbstractScene was not able to render content"), err)
	}
	return canvas, err
}

func (scene AbstractScene) SendRender(canvas Canvas) {
	draw_request := &proto.DrawFillRequest{
		Id:     scene.LayerID,
		Width:  int(scene.Content.GetActualSize().Width),
		Height: int(scene.Content.GetActualSize().Height),
		Img:    canvas,
	}
	AppInstance().SendRequest(draw_request)
}

// Returns top gesture available in this location
func (scene *AbstractScene) FindGesture(event MouseEvent) Gesture {
	if scene.Content == nil {
		return nil
	}
	return scene.Content.FindGesture(event)
}

// Scene – interface for implementing multiple standalone objects that can be
// shown on screen and handle incomming events
// Unlike Container, Scene is not View
// Scene communicates with window server
type Scene interface {
	RequestLayerID() proto.ID                  // Method for requesting new layer ID from Window Server
	GetEventChannel() chan *proto.EventRequest // Method for returning events incomming connection
	EventHandler(ctx context.Context)          // Handler for incomming events
	Emit()
	Close() error
	// Typical canvas routines
	Resize(size Size)         // Resize scene
	Move(translation Vector)  // Move scene without render
	Redraw()                  // Force rendraw scene without render
	Render() (Canvas, error)  // Render scene
	SendRender(canvas Canvas) // Send canvas to window server
}

type WindowObject struct {
	AbstractScene
	title      string
	titleText  *TextObject
	content    View
	lastShift  Vector
	onClose    func()
	onMinimize func()
	onMaximize func()
}

func Window(title string, content View) *WindowObject {
	window := new(WindowObject)
	window.title = title
	window.content = content
	// Building window
	// Gestures
	windowMoveGesture := DragGesture(termbox.MouseLeft).
		OnChanged(func(value Value) {
			window.Move(value.Translation.Sub(window.lastShift))
			window.lastShift = value.Translation
		}).
		OnEnded(func(value Value) {
			window.lastShift = Vector{0, 0}
		})
	resizeGesture := DragGesture(termbox.MouseLeft).OnChanged(func(value Value) {
		oldSize := window.Content.GetActualSize().ToVector()
		shift := value.Translation.Sub(window.lastShift)
		newActSize := oldSize.Add(shift).ToSize()
		window.Resize(newActSize)
		window.lastShift = value.Translation
	}).OnEnded(func(value Value) {
		window.lastShift = Vector{0, 0}
	})
	// Items
	window.titleText = Text(window.title).
		Foreground(White).
		Background(Grey).
		Align(Center).
		MaxSize(Size{Infinite, Infinite}).
		Gesture(windowMoveGesture)
	// Window itself
	window.Content = VStack(
		HStack(
			Button("X", func(outlet *ButtonObject) {
				go func() {
					window.Close()
					if window.onClose != nil {
						window.onClose()
					}
				}()
			}).
				Foreground(White).
				Background(Red),

			Button("-", func(outlet *ButtonObject) {
				log.Printf("- detected")
				if window.onMinimize != nil {
					window.onMinimize()
				}
			}).
				Foreground(Grey).
				Background(Yellow),

			Button("+", func(outlet *ButtonObject) {
				log.Printf("+ detected")
				if window.onMaximize != nil {
					window.onMaximize()
				}
			}).
				Foreground(White).
				Background(Green),
			window.titleText,
		).
			MaxSize(Size{Infinite, 1}),
		ZStack(
			Rectangle(Size{1, 12}).
				Color(White),
			window.content,
			Box(
				Text(">>>").
					Foreground(Black).
					Background(White).
					Gesture(resizeGesture),
			).
				Gravity(Gravity{Right, Bottom}).
				MaxSize(Size{Infinite, Infinite}),
		),
	)
	return window
}

func (window *WindowObject) OnClose(action func()) *WindowObject {
	window.onClose = action
	return window
}

func (window *WindowObject) SetTitle(s string) *WindowObject {
	window.title = s
	if window.Content != nil {
		window.titleText.SetText(s)
	}
	return window
}

func (window *WindowObject) SetSize(size Size) *WindowObject {
	window.Content.SetActualSize(size)
	return window
}
