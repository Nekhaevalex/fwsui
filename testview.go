package fwsui

func tf() {
	array := make([]View, 10)

	text := Text("Hello, World!")
	spacer := Spacer()
	button := Button("Greet", func(outlet *ButtonObject) {})
	tfValue := ""
	textField := TextField(&tfValue, "Input")

	array = append(array, text)
	array = append(array, button)
	array = append(array, spacer)
	array = append(array, textField)
}
