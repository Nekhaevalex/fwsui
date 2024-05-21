package fwsui

import (
	"log"
	"testing"

	"github.com/Nekhaevalex/fwsprotocol"
)

func TestAllocateCanvas(t *testing.T) {
	canvas, err := AllocateCanvas(Size{10, 5})
	if err != nil {
		log.Fatal(err)
	}
	if len(canvas) != 10 || len(canvas[0]) != 5 {
		log.Fatal("Wrong size")
	}

	canvas, err = AllocateCanvas(Size{0, 0})
	if err == nil {
		log.Fatal("Error not raised")
	}
	if canvas != nil {
		log.Fatal("Canvas not empty")
	}
}

func TestChange(t *testing.T) {
	cell := fwsprotocol.Cell{
		Ch:        'r',
		Fg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Bg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Attribute: fwsprotocol.Attr(0),
	}
	canvas, err := AllocateCanvas(Size{10, 6})
	if err != nil {
		log.Fatal(err)
	}
	canvas.Set(Point{5, 5}, cell)
	if canvas.Get(Point{5, 5}).Ch != 'r' {
		t.Fatal("Point was not saved")
	}
}

func TestSize(t *testing.T) {
	canvas, err := AllocateCanvas(Size{10, 6})
	if err != nil {
		log.Fatal(err)
	}
	if canvas.Size().Height != 6 || canvas.Size().Width != 10 {
		log.Fatal("Incorrect size")
	}
}

func TestInpaint(t *testing.T) {
	cell := fwsprotocol.Cell{
		Ch:        'r',
		Fg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Bg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Attribute: fwsprotocol.Attr(0),
	}
	screen, errScreen := AllocateCanvas(Size{30, 20})
	if errScreen != nil {
		log.Fatal(errScreen)
	}
	child, errChild := AllocateCanvas(Size{10, 10})
	if errChild != nil {
		log.Fatal(errChild)
	}
	// Fill child
	for x := 0; x < int(child.Size().Width); x++ {
		for y := 0; y < int(child.Size().Height); y++ {
			child.Set(Point{x, y}, cell)
		}
	}

	// Inpaint
	// Center
	screen.Inpaint(child, Vector{5, 5})
	// Out of bound
	screen.Inpaint(child, Vector{-100, -100})
	// Right end
	screen.Inpaint(child, Vector{20, 5})
	// Left end
	screen.Inpaint(child, Vector{-5, 5})
	// Top
	screen.Inpaint(child, Vector{5, -5})
	// Bottom
	screen.Inpaint(child, Vector{5, 15})
}

func TestResize(t *testing.T) {
	screen, errScreen := AllocateCanvas(Size{30, 20})
	if errScreen != nil {
		log.Fatal(errScreen)
	}
	child, errChild := AllocateCanvas(Size{10, 10})
	if errChild != nil {
		log.Fatal(errChild)
	}
	// Fill child
	cell := fwsprotocol.Cell{
		Ch:        'r',
		Fg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Bg:        fwsprotocol.Color{A: 1, R: 1, G: 1, B: 1},
		Attribute: fwsprotocol.Attr(0),
	}
	for x := 0; x < int(child.Size().Width); x++ {
		for y := 0; y < int(child.Size().Height); y++ {
			child.Set(Point{x, y}, cell)
		}
	}
	screen.Inpaint(child, Vector{0, 0})
	screen.Resize(Size{10, 10})
	if screen.Get(Point{5, 5}).Ch != 'r' {
		t.Fatal("Image not saved")
	}
}
