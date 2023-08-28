package fwsui

import proto "github.com/Nekhaevalex/fwsprotocol"

type Container interface {
	getChildrenGestures(x, y int) []GestureDescriptor
}

type _Box struct {
	x, y, width, height int
	awidth, aheight     int
	gravityX            Align
	gravityY            Align
	child               View
}

func Box(child View) *_Box {
	box := new(_Box)
	box.child = child
	box.gravityX = Center
	box.gravityY = Center
	box.width, box.height = child.getLogicalSize()
	return box
}

func (box *_Box) getLogicalSize() (int, int) {
	return box.width, box.height
}

func (box *_Box) getActualSize() (int, int) {
	return box.awidth, box.aheight
}

func (box *_Box) getGesture() Gesture {
	return nil
}

func (box *_Box) hasGesture() bool {
	return false
}

func (box *_Box) getPos() (int, int) {
	return box.x, box.y
}

func (box *_Box) setPos(x, y int) {
	box.x = x
	box.y = y
}

func (box *_Box) render(width, height int) [][]proto.Cell {
	box.awidth = width
	box.aheight = height
	// Allocating canvas
	canvas := allocateCanvas(width, height)
	// Check if child view has floating size
	floating_x, floating_y := viewSizeFloating(box.child)
	c_size_x, c_size_y := box.child.getLogicalSize()
	// // Set new box size (from render directive arguments)
	// box.width = width
	// box.height = height

	// assign child sizes
	if floating_x {
		c_size_x = width
	}

	if floating_y {
		c_size_y = height
	}

	// Calculate shifts from gravity parameters
	var shift_x int
	var shift_y int
	switch box.gravityX {
	case Left:
		shift_x = 0
	case Right:
		shift_x = width - c_size_x
	case Center:
		shift_x = width/2 - c_size_x/2
	}
	switch box.gravityY {
	case Left:
		shift_y = 0
	case Right:
		shift_y = height - c_size_y
	case Center:
		shift_y = height/2 - c_size_y/2
	}

	// Calculate drawing start position
	start_x := max(0, shift_x)
	start_y := max(0, shift_y)
	last_x := min(width, shift_x+c_size_x)
	lasy_y := min(height, shift_y+c_size_y)
	box.child.setPos(start_x+box.x, start_y+box.y)
	child_canvas := box.child.render(c_size_x, c_size_y)
	for x := start_x; x < last_x; x++ {
		for y := start_y; y < lasy_y; y++ {
			canvas[x][y] = child_canvas[abs(x-shift_x)][abs(y-shift_y)]
		}
	}
	return canvas
}

func (box *_Box) SetSize(width, height int) *_Box {
	box.width = -1
	box.height = -1
	return box
}

func (box *_Box) Gravity(x, y Align) *_Box {
	box.gravityX = x
	box.gravityY = y
	return box
}

func (box *_Box) getChildrenGestures(x, y int) []GestureDescriptor {
	actors := make([]GestureDescriptor, 0, 1)
	if box.child.hasGesture() {
		actors = append(actors, box.child.getGesture().getGestureDescriptor(0, 0))
	}
	if asserted, ok := box.child.(Container); ok {
		actors = append(actors, asserted.getChildrenGestures(box.x, box.y)...)
	}
	return actors
}

type _HStack struct {
	x, y, width, height int
	awidth, aheight     int
	padding             int
	gravityX            Align
	gravityY            Align
	children            []View
}

// getGesture implements View.
func (*_HStack) getGesture() Gesture {
	return nil
}

// hasGesture implements View.
func (*_HStack) hasGesture() bool {
	return false
}

func (hstack *_HStack) getLogicalSize() (int, int) {
	return hstack.width, hstack.height
}

func (hstack *_HStack) getActualSize() (int, int) {
	return hstack.awidth, hstack.aheight
}

func (hstack *_HStack) setPos(x, y int) {
	hstack.x = x
	hstack.y = y
}

func (hstack *_HStack) getPos() (int, int) {
	return hstack.x, hstack.y
}

func (hstack *_HStack) render(width, height int) [][]proto.Cell {
	hstack.awidth = width
	hstack.aheight = height
	// get total fixed size over X axis, maximal size over Y axis and fixed size elements amount
	total_fixed_size_X := 0
	max_size_Y := 0
	fixed_size_amount := 0
	for _, child := range hstack.children {
		x, y := child.getLogicalSize()
		if x > 0 {
			total_fixed_size_X += x
			fixed_size_amount += 1
		}
		max_size_Y = max(max_size_Y, y)
	}
	if height < 0 {
		height = max_size_Y
	}
	// calculating horizontal space per floating object
	floating_objects_amount := len(hstack.children) - fixed_size_amount
	free_space_per_obj := 0
	if floating_objects_amount > 0 {
		free_space_per_obj = (width - total_fixed_size_X - (len(hstack.children)+1)*hstack.padding) / floating_objects_amount
	}
	max_x := min(width, total_fixed_size_X+floating_objects_amount*free_space_per_obj)
	canvas := allocateCanvas(width, height)
	x := hstack.padding
	y := 0
	for _, child := range hstack.children {
		w, h := child.getLogicalSize()
		if w < 0 {
			w = free_space_per_obj
		}
		if h < 0 {
			h = height
		}
		box := Box(child)
		box.gravityX = hstack.gravityX
		box.gravityY = hstack.gravityY
		box.setPos(x, y)
		sub_frame := box.render(w, h)
		for ix := x; ix < min(max_x, x+w); ix++ {
			for iy := y; iy < h; iy++ {
				canvas[ix][iy] = sub_frame[ix-x][iy-y]
			}
		}
		x += (w + hstack.padding)
	}
	return canvas
}

func (hstack *_HStack) getChildrenGestures(x, y int) []GestureDescriptor {
	actors := make([]GestureDescriptor, 0)
	for _, child := range hstack.children {
		if child.hasGesture() {
			actors = append(actors, child.getGesture().getGestureDescriptor(hstack.x, hstack.y))
		}
		if asserted, ok := child.(Container); ok {
			actors = append(actors, asserted.getChildrenGestures(hstack.x, hstack.y)...)
		}
	}
	return actors
}

func (hstack *_HStack) Padding(padding int) *_HStack {
	hstack.padding = padding
	return hstack
}

func (hstack *_HStack) SetSize(x, y int) *_HStack {
	hstack.width = x
	hstack.height = y
	return hstack
}

func (hstack *_HStack) Gravity(x, y Align) *_HStack {
	hstack.gravityX = x
	hstack.gravityY = y
	return hstack
}

func (hstack *_HStack) AddView(view View) *_HStack {
	hstack.children = append(hstack.children, view)
	return hstack
}

func HStack(children ...View) *_HStack {
	hstack := new(_HStack)
	hstack.children = children
	hstack.x = 0
	hstack.y = 0
	hstack.width = -1
	hstack.height = -1
	hstack.padding = 0
	hstack.gravityX = Center
	hstack.gravityY = Center
	return hstack
}

type _VStack struct {
	x, y, width, height int
	awidth, aheight     int
	padding             int
	gravityX            Align
	gravityY            Align
	children            []View
}

// getGesture implements View.
func (*_VStack) getGesture() Gesture {
	return nil
}

// hasGesture implements View.
func (*_VStack) hasGesture() bool {
	return false
}

func (vstack *_VStack) getLogicalSize() (int, int) {
	return vstack.width, vstack.height
}

func (vstack *_VStack) getActualSize() (int, int) {
	return vstack.awidth, vstack.aheight
}

func (vstack *_VStack) setPos(x, y int) {
	vstack.x = x
	vstack.y = y
}

func (vstack *_VStack) getPos() (int, int) {
	return vstack.x, vstack.y
}

func (vstack *_VStack) render(width, height int) [][]proto.Cell {
	vstack.awidth = width
	vstack.aheight = height
	// get total fixed size over X axis, maximal size over Y axis and fixed size elements amount
	total_fixed_size_Y := 0
	max_size_X := 0
	fixed_size_amount := 0
	for _, child := range vstack.children {
		x, y := child.getLogicalSize()
		if y > 0 {
			total_fixed_size_Y += y
			fixed_size_amount += 1
		}
		max_size_X = max(max_size_X, x)
	}
	if width < 0 {
		width = max_size_X
	}
	// calculating horizontal space per floating object
	floating_objects_amount := len(vstack.children) - fixed_size_amount
	free_space_per_obj := 0
	if floating_objects_amount > 0 {
		free_space_per_obj = (height - total_fixed_size_Y - len(vstack.children)*vstack.padding) / floating_objects_amount
	}
	max_y := min(height, total_fixed_size_Y+floating_objects_amount*free_space_per_obj)
	canvas := allocateCanvas(width, height)
	y := vstack.padding
	x := 0
	for _, child := range vstack.children {
		_, h := child.getLogicalSize()
		if h < 0 {
			h = free_space_per_obj
		}
		w := width
		box := Box(child)
		box.gravityX = vstack.gravityX
		box.gravityY = vstack.gravityY
		box.setPos(x, y)
		sub_frame := box.render(w, h)
		for ix := x; ix < w; ix++ {
			for iy := y; iy < min(max_y, y+h); iy++ {
				canvas[ix][iy] = sub_frame[ix-x][iy-y]
			}
		}
		y += (h + vstack.padding)
	}
	return canvas
}

func (vstack *_VStack) getChildrenGestures(x, y int) []GestureDescriptor {
	actors := make([]GestureDescriptor, 0)
	for _, child := range vstack.children {
		if child.hasGesture() {
			actors = append(actors, child.getGesture().getGestureDescriptor(vstack.x, vstack.y))
		}
		if asserted, ok := child.(Container); ok {
			actors = append(actors, asserted.getChildrenGestures(vstack.x, vstack.y)...)
		}
	}
	return actors
}

func (vstack *_VStack) Padding(padding int) *_VStack {
	vstack.padding = padding
	return vstack
}

func (vstack *_VStack) SetSize(x, y int) *_VStack {
	vstack.width = x
	vstack.height = y
	return vstack
}

func (vstack *_VStack) Gravity(x, y Align) *_VStack {
	vstack.gravityX = x
	vstack.gravityY = y
	return vstack
}

func (vstack *_VStack) AddView(view View) *_VStack {
	vstack.children = append(vstack.children, view)
	return vstack
}

func VStack(children ...View) *_VStack {
	vstack := new(_VStack)
	vstack.children = children
	vstack.x = 0
	vstack.y = 0
	vstack.width = -1
	vstack.height = -1
	vstack.padding = 0
	vstack.gravityX = Center
	vstack.gravityY = Center
	return vstack
}

type _ZStack struct {
	x, y, width, height int
	gravityX, gravityY  Align
	children            []View
}

// getGesture implements View.
func (*_ZStack) getGesture() Gesture {
	return nil
}

// hasGesture implements View.
func (*_ZStack) hasGesture() bool {
	return false
}

func (zstack *_ZStack) getLogicalSize() (int, int) {
	return zstack.width, zstack.height
}

func (zstack *_ZStack) getActualSize() (int, int) {
	return zstack.width, zstack.height
}

func (zstack *_ZStack) getPos() (int, int) {
	return zstack.x, zstack.y
}

func (zstack *_ZStack) setPos(x, y int) {
	zstack.x = x
	zstack.y = y
}

func (zstack *_ZStack) render(width, height int) [][]proto.Cell {
	canvas := allocateCanvas(width, height)
	for _, child := range zstack.children {
		boxed := Box(child)
		boxed.gravityX = zstack.gravityX
		boxed.gravityY = zstack.gravityY
		boxed.setPos(zstack.x, zstack.y)
		layer := boxed.render(width, height)
		for i := 0; i < width; i++ {
			for j := 0; j < height; j++ {
				canvas[i][j] = layer[i][j].Over(canvas[i][j])
			}
		}
	}
	return canvas
}

func (zstack *_ZStack) getChildrenGestures(x, y int) []GestureDescriptor {
	actors := make([]GestureDescriptor, 0)
	for _, child := range zstack.children {
		if child.hasGesture() {
			actors = append(actors, child.getGesture().getGestureDescriptor(zstack.x, zstack.y))
		}
		if asserted, ok := child.(Container); ok {
			actors = append(actors, asserted.getChildrenGestures(zstack.x, zstack.y)...)
		}
	}
	return actors
}

func ZStack(children ...View) *_ZStack {
	zstack := new(_ZStack)
	zstack.children = children
	zstack.x = 0
	zstack.y = 0
	zstack.width = -1
	zstack.height = -1
	zstack.gravityX = Center
	zstack.gravityY = Center
	return zstack
}
