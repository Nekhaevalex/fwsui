# F Window System UI Library

# Intro
This library is purposed for building FWS apps using Go.
Although neither FWS nor this library should not be used in real life tasks due to lack of anything like reliability unless you want to make it better or simply make couple of apps for FWS.

# Installation
1. Install FWS: https://github.com/Nekhaevalex/fws (simply clone it, build fws.go and run fws in any terminal that supports 24-bit colors)
2. Make new go module with for your app.
3. Install library: `go get github.com/Nekhaevalex/fwsui`
4. Write your app.
5. Make sure FWS instance is running.
6. Run your app.

# Hello, World!
Simple Hello World app example:
```go
package main

import "github.com/Nekhaevalex/fwsui"

func main() {
    App(
        Window("Hello, World!", Text("Hello, World!")).OnClose(func() {
	    	AppInstance().Quit()
	    }),
    )
}
```

# Components
FWSUI consists of following core components types:

* App
* Scene
* Container
* View

and some service components:
* Gesture
* KeyHandler

## Core components
### App
`App` provides object that can establish connection with F Window Server and route messages between it's scenes and server.

Only one `App` object can exist (singleton). You can pass any amount of objects implementing `Scene` interface (e.g. `Window`) to `App(...)` function which will be shown when your app starts.

`App` object can be received any time with globaly available `AppInstance()` function.

`App` object provides 2 methods:
1. `OpenWindow(scene Scene)` which shows new `Scene` object.
2. `Quit()` which shuts down the app.

### Scene
`Scene` – interface for implementing standalone objects that can be shown on screen and handle incomming events.

Simple example of `Scene` objects is `Window`.

#### Window
`Window` provides simple window object that can be closed, moved, resized. It receives 2 objects:
1. Window title (which will be written on the titlebar)
2. View that will be shown inside the window

### Container
Container is any object that can order one or more Views and render them.
There are 4 containers available:
1. VStack – verical stack of Views
2. HStack - horizontal stack of Views
3. ZStack - multilayered stack of Views (from the bottom)
4. Box - container for single view.

You can set up Gravity for each view – it defines the alignment of objects in cells. For Y-axis gravity Left equal to top, Right to Bottom.

### View
View in minimal object that can be rendered on screen.
There are 4 views so far:
1. Spacer - transparent object that can occupy specified space
2. Text - single line text box
3. Button - single line clickable button with specified action on click.
4. TextField - single line field for text input.

### Gesture
Gesture objects can be passed to Text object and do some specified action if triggered. There are 4 gestures so far:
1. LClickGesture - Left mouse click
2. MClickGesture - Middle mouse click
3. RClickGesture - Right mouse click
4. DragGesture - Drag gesture

### KeyHandler
Will be described later. Used for handling keyboard keys. Used only in TextField now.