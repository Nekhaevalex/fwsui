package fwsui

import (
	"testing"
)

func TestHStack(t *testing.T) {
	text := Text("Hello, World!")
	spacer := Spacer()
	button := Text("Greeting")

	stack := HStack(text, spacer, button)
	stack.SetActualSize(Size{20, 10})
	canvas, err := stack.Render()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(canvas)
	t.Log(err)
	t.Log(stack)
}
