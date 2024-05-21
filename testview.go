package fwsui

func tf() {
	array := make([]View, 10)

	text := Text("Hello, World!")
	spacer := Spacer()
	button := Button("Greet", func(outlet *ButtonObject) {})
	tfValue := ""
	textField := TextField(&tfValue, "Input")

	stack := AbstractStack(X, text, spacer, button, textField)
	stack.Render()
}
