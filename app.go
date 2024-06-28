package fwsui

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	proto "github.com/Nekhaevalex/fwsprotocol"
)

var (
	ErrorAppInit                  = errors.New("app failed to initialize")                                                // app failed to initialize
	ErrorAppEarlyInit             = errors.New("app was initializies out of App constructor")                             // app was initializies out of App constructor
	ErrorAppNoContext             = errors.New("app context is nil")                                                      // app context is nil
	ErrorMessengerInit            = errors.New("messenger failed to initialize")                                          // messenger failed to initialize
	ErrorMessengerConnect         = errors.New("messenger failed to establish connection")                                // messenger failed to establish connection
	ErrorMessengerShutDown        = errors.New("messenger failed to shutdown")                                            // messenger failed to shutdown
	ErrorMessengerReceive         = errors.New("messenger failed to receive message")                                     // messenger failed to receive message
	ErrorMessengerHandler         = errors.New("messenger's incomming messege handler failed")                            // messenger's incomming messege handler failed
	ErrorMessengerSend            = errors.New("messenger failed to send message")                                        // messenger failed to send message
	ErrorMessengerForwardedClosed = errors.New("messenger failed to receive forwarded message because channel is closed") // messenger failed to receive forwarded message because channel is closed
	ErrorMessengerTimeout         = errors.New("parallel receiver reached timeout")                                       // parallel receiver reached timeout
	ErrorMessengerNoCancel        = errors.New("messenger has no cancel function")                                        // messenger has no cancel function
)

// Object for communicating with window server.
// Runs async message receiver.
// Provides function to send message on demand.
type Messenger struct {
	conn                          net.Conn           // Connection with window server
	awaitingReply, receiverActive bool               // state flags
	cancel                        context.CancelFunc // Function to be called when process finishes
	forwarded                     chan proto.Request // buffered channel for forwarded replies
}

// Establishes connection
func (m *Messenger) establishConnection(network, address string, pid int) error {
	conn, err := net.Dial(network, address)
	if err != nil {
		return fmt.Errorf("messenger failed to dial fws at %s via %s protocol on behalf of process %d, make sure fws is running", address, network, pid)
	}
	pid_cache := make([]byte, 4)
	binary.LittleEndian.PutUint32(pid_cache, uint32(pid))
	_, err = conn.Write(pid_cache)
	if err != nil {
		return fmt.Errorf("messenger failed to send pid to fws at %s via %s protocol on behalf of process %d", address, network, pid)
	}
	rdy := make([]byte, 10)
	b, err := conn.Read(rdy)
	if err != nil {
		return fmt.Errorf("messenger failed to receive reply from fws at %s via %s protocol on behalf of process %d", address, network, pid)
	}
	str1 := string(rdy[:b])
	if str1 != "READY" {
		return fmt.Errorf("messenger received string '%s' from fws at %s via %s protocol on behalf of process %d instead of string 'READY'", str1, address, network, pid)
	}
	m.conn = conn
	return nil
}

func (m *Messenger) receiveMessage() (proto.Request, error) {
	buff := make([]byte, 4096)
	n, err := m.conn.Read(buff)
	if err != nil {
		switch err {
		case io.EOF:
			return nil, err
		default:
			return nil, errors.Join(ErrorMessengerReceive, err)
		}
	}

	msg := proto.Msg(buff[:n])
	request := msg.Decode()
	return request, nil
}

// Shuts down messenger
func (m *Messenger) ShutDown() error {
	err := m.conn.Close()
	if err != nil {
		return errors.Join(ErrorMessengerShutDown, err)
	}
	if m.cancel != nil {
		m.cancel()
		return nil
	}
	return ErrorMessengerNoCancel
}

// incomingMessagesHandler - handler which catches all incomming messages from window server.
// It's quite important that it catches ALL messages including those which are awaited by SendRequest.
// Hopefully it will forward them to SendRequest.
func (m *Messenger) handleIncommingMessages(ctx context.Context, route func(*proto.EventRequest) error) {
	defer func() {
		m.receiverActive = false
		close(m.forwarded)
	}()

	m.receiverActive = true
	for {
		select {
		case <-ctx.Done():
			return
		default:
			request, err := m.receiveMessage()
			if err != nil {
				log.Fatal(errors.Join(ErrorMessengerHandler, err))
			}
			switch tRequest := request.(type) {
			case *proto.EventRequest:
				go func() {
					err := route(tRequest)
					if err != nil {
						log.Fatal(errors.Join(ErrorMessengerHandler, err))
					}
				}()
			default:
				if m.awaitingReply {
					m.forwarded <- tRequest
				}
			}
		}
	}
}

func (m *Messenger) parallelReceive() (proto.Request, error) {
	if !m.receiverActive {
		reply, err := m.receiveMessage()
		if err != nil {
			return nil, err
		}
		return reply, nil
	} else {
		timer := time.NewTimer(10 * time.Millisecond)
		timeout := timer.C
		for {
			select {
			case <-timeout:
				timer.Stop()
				return nil, ErrorMessengerTimeout
			case reply, ok := <-m.forwarded:
				if !ok {
					return nil, ErrorMessengerForwardedClosed
				}
				return reply, nil
			}
		}
	}
}

func (m *Messenger) SendRequest(request proto.Request) (proto.Request, error) {
	defer func() {
		m.awaitingReply = false
	}()

	m.awaitingReply = true
	for {
		_, err := m.conn.Write(request.Encode())
		if err != nil {
			return nil, errors.Join(ErrorMessengerSend, err)
		}

		reply, err := m.parallelReceive()
		if err != nil {
			if errors.Is(err, ErrorMessengerTimeout) {
				continue
			} else {
				return nil, err
			}
		}
		return reply, nil
	}
}

// Creates and sets up new Messenger object.
// Requires network - connection type (refer to net.Dial for more information).
// address - connection address in string format.
// route - callback with for EventRequest routing
func NewMessenger(network, address string, pid int, route func(*proto.EventRequest) error) (*Messenger, error) {
	messenger := new(Messenger)
	err := messenger.establishConnection(network, address, pid)
	if err != nil {
		return nil, errors.Join(ErrorMessengerInit, ErrorMessengerConnect, err)
	}
	messenger.forwarded = make(chan proto.Request, 100)
	ctx, cancel := context.WithCancel(context.Background())
	messenger.cancel = cancel
	go messenger.handleIncommingMessages(ctx, route)
	return messenger, nil
}

// Represents Application.
// Application is a contacting point for communicating with window server.
// It catches all messages from server and routes them and also provides public method for
// allowing GUI elements to send their messages to window server.
//
// AppObject provides following methods:
//   - OpenWindow - method for opening new window/scene
//   - SendRequest – method for sending requests to window server
//   - SetInput – sets new keyboard input consumer
//
// Note: Only one AppObject allowed per process.
// As a result you can call AppInstance() function which always returns pointer to
// current AppObject.
type AppObject struct {
	Pid          int                // Stores process ID
	Scenes       map[proto.ID]Scene // Child Scenes storage
	messenger    *Messenger         // Messenger structure for communicating with FWS
	keyInputChan *chan *proto.EventRequest
	ctx          context.Context    // App context
	cancel       context.CancelFunc // App context cancel
}

// AppObject instance. Only one allowed per process.
// Private, but can be retrieved via AppInstance() function
var appInstance *AppObject

// Provides singletone functionality for AppObject
var once sync.Once

// Initializes new AppObject
func createAppObject() (*AppObject, error) {
	var gerr error = nil
	once.Do(func() {
		appInstance = new(AppObject)
		appInstance.Pid = os.Getpid()
		appInstance.Scenes = make(map[proto.ID]Scene)
		appInstance.ctx, appInstance.cancel = context.WithCancel(context.Background())
		// Trying to establish connection with window server
		if err := appInstance.establishConnection(); err != nil {
			gerr = err
		}
	})
	return appInstance, gerr
}

// Established connection with window server
func (app *AppObject) establishConnection() error {
	var err error
	app.messenger, err = NewMessenger(
		"unix",
		proto.FWS_SOCKET,
		app.Pid,
		func(er *proto.EventRequest) error {
			lId := er.Id
			scene, ok := app.Scenes[lId]
			if !ok {
				return fmt.Errorf("layer with ID %d not found", lId)
			}
			channel := scene.GetEventChannel()
			channel <- er
			return nil
		},
	)
	return err
}

func AppInstance() *AppObject {
	appInstance, err := createAppObject()
	if err != nil {
		log.Panic(errors.Join(ErrorAppEarlyInit, err))
	}
	return appInstance
}

// Creates new AppObject with initial Scenes runned
func App(initialScene ...Scene) *AppObject {
	appInstance, err := createAppObject()
	if err != nil {
		log.Fatal(errors.Join(ErrorAppInit, err))
	}
	defer func() {
		appInstance.messenger.ShutDown()
	}()
	for _, window := range initialScene {
		appInstance.EmitScene(window)
	}
	if appInstance.ctx != nil {
		done := <-appInstance.ctx.Done()
		if done == struct{}{} {
			return appInstance
		}
	} else {
		log.Fatal(errors.Join(ErrorAppInit, ErrorAppNoContext))
	}
	return appInstance
}

// SendRequest - public methods that allows any method inside GUI to send messages to window server.
// Returns window server reply.
func (app *AppObject) SendRequest(request proto.Request) (proto.Request, error) {
	return app.messenger.SendRequest(request)
}

func (app *AppObject) SetInput(channel *chan *proto.EventRequest) {
	app.keyInputChan = channel
}

// Emits new Scene object on screen (including all necessary preparations and initializations)
func (app *AppObject) EmitScene(window Scene) {
	lid := window.RequestLayerID()
	app.Scenes[lid] = window
	// Initial render
	app.Scenes[lid].Move(Vector{5, 5})
	app.Scenes[lid].Resize(Size{30, 10})
	canvas, err := app.Scenes[lid].Render()
	if err != nil {
		log.Fatal(errors.Join(fmt.Errorf("failed to complete initial render of scene %d", lid)), err)
	}
	app.Scenes[lid].SendRender(canvas)
	app.Scenes[lid].Redraw()
	// Create event channel
	app.Scenes[lid].GetEventChannel()
	// Raise event handler
	go app.Scenes[lid].Emit()
}

func (app *AppObject) Quit() {
	defer func() {
		app.messenger.ShutDown()
	}()
	app.cancel()
}
