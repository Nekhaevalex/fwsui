package fwsui

import (
	"context"

	proto "github.com/Nekhaevalex/fwsprotocol"
	"github.com/nsf/termbox-go"
)

type KeyHandler interface {
	EventHandler()
	Activate()
	Deactivate()
	SetKeyHandler(handler func(key termbox.Key))
}

type AbstractKeyHandler struct {
	input     *chan *proto.EventRequest
	ctx       context.Context
	cancel    context.CancelFunc
	active    bool
	handleKey func(event proto.EventRequest)
}

func (akh *AbstractKeyHandler) EventHandler() {
	if akh.ctx == nil || akh.cancel == nil {
		return
	}

	for {
		select {
		case <-akh.ctx.Done():
			return
		case key := <-*akh.input:
			akh.handleKey(*key)
		}
	}
}

func (akh AbstractKeyHandler) IsActive() bool {
	return akh.active
}

func (akh *AbstractKeyHandler) Activate() {
	AppInstance().SetInput(akh.input)
	akh.ctx, akh.cancel = context.WithCancel(context.Background())
	go akh.EventHandler()
	akh.active = true
}

func (akh *AbstractKeyHandler) Deactivate() {
	akh.cancel()
	akh.active = false
}

func (akh *AbstractKeyHandler) SetKeyHandler(handler func(event proto.EventRequest)) {
	akh.handleKey = handler
}
