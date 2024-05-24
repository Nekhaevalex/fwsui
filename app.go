package fwsui

import (
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

type AppObject struct {
	pid              int
	windowServerConn net.Conn
	scenes           map[proto.ID]Scene
	eventCatcherFl   bool
	requestActive    bool
	forwardReplies   []proto.Msg
	forwardRepliesM  sync.Mutex
	keyInputChan     *chan *proto.EventRequest
	quit             chan int
}

func (app *AppObject) establishConnection() {
	app.pid = os.Getpid()
	conn, err := net.Dial("unix", proto.FWS_SOCKET)
	if err != nil {
		log.Fatal(errors.Join(fmt.Errorf("app %d failed to connect to fws", app.pid), err))
	}
	app.windowServerConn = conn
	pid_cache := make([]byte, 4)
	binary.LittleEndian.PutUint32(pid_cache, uint32(app.pid))
	_, err = conn.Write(pid_cache)
	if err != nil {
		log.Fatal(errors.Join(fmt.Errorf("app %d failed sending pid to fws", app.pid), err))
	}
	rdy := make([]byte, 10)
	b, err := conn.Read(rdy)
	if err != nil {
		log.Fatal(errors.Join(fmt.Errorf("app %d failed recieving ready state from fws", app.pid), err))
	}
	str1 := string(rdy[:b])
	if str1 != "READY" {
		return
	}
}

func (app *AppObject) sendRequest(request proto.Request) proto.ID {
	app.requestActive = true
	defer func() {
		app.requestActive = false
	}()
	for {
		// Send requesst
		_, err := app.windowServerConn.Write(request.Encode())
		if err != nil {
			log.Fatal(err)
		}
		// Receive reply / ack
		var msg proto.Msg
		if !app.eventCatcherFl {
			// Branch if event catcher is not active yet -- receiving self
			buff := make([]byte, 1024)
			n, err := app.windowServerConn.Read(buff)
			if err != nil {
				log.Fatal(err)
			}
			msg = proto.Msg(buff[:n])
		} else {
			// Branch if event catcher is already active:
			// it will receive all incomming connections and if it's not event,
			// will reroute it here
			sendAgain := false
			waiting := true
			timer := time.NewTimer(100 * time.Millisecond)
			timeout := timer.C
			for waiting {
				select {
				case <-timeout:
					sendAgain = true
					waiting = false
					timer.Stop()
				default:
					if len(app.forwardReplies) > 0 {
						app.forwardRepliesM.Lock()
						msg = app.forwardReplies[0]
						app.forwardReplies = app.forwardReplies[1:]
						app.forwardRepliesM.Unlock()
						waiting = false
					}
				}
			}
			if sendAgain {
				continue
			}
		}
		reply := msg.Decode()
		switch final := reply.(type) {
		case *proto.RepeatRequest:
			continue
		case *proto.AckRequest:
			return final.Id
		case *proto.ReplyCreationRequest:
			return final.Id
		default:
			continue
		}
	}
}

func (app *AppObject) incomingMessagesHandler() {
	// Event catcher
	app.eventCatcherFl = true
	defer func() { app.eventCatcherFl = false }()
	for {
		buff := make([]byte, 4096)
		n, err := app.windowServerConn.Read(buff)
		if err != nil {
			switch err {
			case io.EOF:
				app.quit <- 1
				return
			default:
				log.Fatal(err)
			}
		}
		msg := proto.Msg(buff[:n])
		request := msg.Decode()
		switch typed_request := request.(type) {
		case *proto.EventRequest:
			// It's event and must be send to window handler
			go func() {
				lId := typed_request.Id
				channel := app.scenes[lId].GetEventChannel()
				channel <- typed_request
			}()
		default:
			// It's message addressed to sendRequest process but received here instead
			// Must be sent to sendRequest
			if app.requestActive {
				app.forwardRepliesM.Lock()
				app.forwardReplies = append(app.forwardReplies, msg)
				app.forwardRepliesM.Unlock()
			}
		}
	}
}

func (app *AppObject) setInput(channel *chan *proto.EventRequest) {
	app.keyInputChan = channel
}

func (app *AppObject) OpenWindow(window Scene) {
	window.BindApp(app)
	lid := window.RequestLayerID()
	app.scenes[lid] = window
	// Initial render
	app.scenes[lid].Move(Vector{5, 5})
	app.scenes[lid].Resize(Size{30, 10})
	canvas, err := app.scenes[lid].Render()
	if err != nil {
		log.Fatal(errors.Join(fmt.Errorf("failed to complete initial render of scene %d", lid)), err)
	}
	app.scenes[lid].SendRender(canvas)
	app.scenes[lid].Redraw()
	// Create event channel
	app.scenes[lid].GetEventChannel()
	// Raise event handler
	go app.scenes[lid].EventHandler()
}

func (app *AppObject) Quit() {
	app.quit <- 1
}

var appInstance *AppObject
var once sync.Once

func AppInstance() *AppObject {
	return appInstance
}

func App(initialScene ...Scene) *AppObject {
	once.Do(func() {
		appInstance = new(AppObject)
		appInstance.scenes = make(map[proto.ID]Scene)
		appInstance.quit = make(chan int)
		appInstance.eventCatcherFl = false
		appInstance.forwardReplies = make([]proto.Msg, 0)
		appInstance.establishConnection()
	})
	defer func() {
		log.Printf("Closing connection")
		appInstance.windowServerConn.Close()
	}()
	for _, window := range initialScene {
		appInstance.OpenWindow(window)
	}
	go appInstance.incomingMessagesHandler()
	// make loop unitl quit received
	for {
		if ok := <-appInstance.quit; ok > 0 {
			log.Printf("%d", ok)
			return appInstance
		}
	}
}
