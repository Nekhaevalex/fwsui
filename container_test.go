package fwsui

import "testing"

func TestPlaneSolver(t *testing.T) {
	message := Text("Select action").
		Background(White).
		Foreground(Black).
		MaxSize(Size{Width: Infinite, Height: Infinite}).
		Align(Center)
	// spacer := Spacer()
	// message2 := Text("Hello, World!").
	// 	Background(Red).
	// 	MaxSize(Size{Width: Infinite, Height: Infinite}).
	// 	Align(Right)
	stack := HStack(message)
	stack.SetActualSize(Size{Width: 40, Height: 10})
	stack.Render()
	for _, view := range stack.children {
		t.Log(view.GetPosition(), view.GetActualSize())
	}
}
