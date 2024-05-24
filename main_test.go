package fwsui

import (
	"errors"
	"log"

	"github.com/nsf/termbox-go"
)

// Renders view of specified size to terminal
func PreviewView(view View, size Size) {
	// Build view
	boxedView := Box(view)
	boxedView.SetActualSize(size)
	boxedView.SetGravity(Gravity{Horizontal: Left, Vertical: Top})
	canvas, err := boxedView.Render()
	if err != nil {
		log.Fatal(errors.Join(errors.New("failed to render"), err))
	}

	//
	err = termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	termbox.SetOutputMode(termbox.OutputRGB)
	for x := 0; x < int(canvas.Size().Width); x++ {
		for y := 0; y < int(canvas.Size().Height); y++ {
			cell := canvas.Get(Point{X: x, Y: y})
			tbCell := cell.ToTerboxCell()
			termbox.SetCell(x, y, tbCell.Ch, tbCell.Fg, tbCell.Bg)
		}
	}
	termbox.Flush()
	event := termbox.PollEvent()
	log.Print(event)
}
