package fwsui

import proto "github.com/Nekhaevalex/fwsprotocol"

type KeyHandler interface {
	getInputChan() chan *proto.EventRequest
	enableInput()
}
