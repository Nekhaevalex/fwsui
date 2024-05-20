// Provides core container components such as AbstractContainer Container interface
// Also provides basic containers such as Box, HStack, VStack, ZStack
package fwsui

import proto "github.com/Nekhaevalex/fwsprotocol"

// Represents abstract container and it's key attributes such as position, sizes
// (inherited from AbstractView), gravity, child.
// Since each container is view it's inherited from AbstractView.
//
// First element that is special to containers is Gravity - they describe how
// child element should be aligned horizontaly and vericaly.
// Second special element is children - it stores contained views.
//
// Since each container must implement View interface it is possible to store
// containers inside containers.
//
// Note: AbstractContainer itself implements Container interface but not View.
type AbstractContainer struct {
	AbstractView
	gravity  Gravity
	children []View
}

// Recursively gets GestureDescriptor of child view for mapping child elements gestures.
// It is required for building map of active areas.
func (ac AbstractContainer) GetChildrenGestures() []GestureDescriptor {
	// allocating descriptor storage
	descriptors := make([]GestureDescriptor, 0, 1)
	for _, child := range ac.children {
		// each child is view and implements HasGesture method.
		if child.HasGesture() {
			childDescriptor := child.GetGesture().GetGestureDescriptor()
			childDescriptor.Position.Translate(Vector(ac.position))
			descriptors = append(descriptors, childDescriptor)
		}
		// Trying assert and if child is container - recursive call GetChildrenGestures
		if asserted, ok := child.(Container); ok {
			descriptors = append(descriptors, asserted.GetChildrenGestures()...)
		}
	}
	return descriptors
}

// Sets gravity for AbstractContainer object
func (ac *AbstractContainer) SetGravity(gravity Gravity) {
	ac.gravity = gravity
}

// Gets gravity of AbstractContainer object
func (ac *AbstractContainer) GetGravity() Gravity {
	return ac.gravity
}

// Applies specified AbstractContainer gravity to children which are containters.
func (ac AbstractContainer) ApplyChildGravity() {
	for _, child := range ac.children {
		if asserted, ok := child.(Container); ok {
			asserted.SetGravity(ac.gravity)
		}
	}
}

// Container - interface for implementing containers.
// Each container must provide several methods.
//
// Since each container should contain gravity variable (see AbstractContainer),
// Container should implement setter and getter.
//
// Each of View in Container may contain Gesture. Thats why active areas of
// gestures must be mapped for quick gesture position identification.
// Hence container must implement GetChildrenGestures which recursively map gestures.
type Container interface {
	GetChildrenGestures() []GestureDescriptor // Recursively gets GestureDescriptor of child view for mapping child elements gestures.
	GetGravity() Gravity
	SetGravity(gravity Gravity)
}

// Represents Box object with single child.
// Since box can be larger then child, child can be aligned to different sides.
type BoxObject struct {
	AbstractContainer
}

// Represents Box object with single child.
// Since box can be larger then child, child can be aligned to different sides.
// By default size is infinite.
// Default gravity is (Center, Center)
func Box(child View) *BoxObject {
	box := new(BoxObject)
	box.children = make([]View, 1)
	box.children = append(box.children, child)
	box.SetPosition(Point{0, 0})
	box.SetMinSize(Size{0, 0})
	box.SetMaxSize(Size{Infinite, Infinite})
	box.SetGesture(nil)
	box.SetGravity(Gravity{Center, Center})
	return box
}

// Solves child's size constraints when Box actual size is already known but
// child's size still can float.
// Returns child's estimated size.
func (box BoxObject) solveConstraintsSize() Size {
	// Alias to child
	child := box.children[0]

	// Solve 1D problem
	solveOnAxis := func(axis Axis) uint {
		actualSize := box.GetActualSize().GetComponent(axis)
		childMinSize := child.GetMinSize().GetComponent(axis)
		childMaxSize := child.GetMaxSize().GetComponent(axis)
		if childMinSize <= actualSize && childMaxSize <= actualSize {
			return childMaxSize
		} else if childMinSize <= actualSize && actualSize <= childMaxSize {
			return actualSize
		} else if childMinSize >= actualSize && childMaxSize > actualSize {
			return childMinSize
		} else {
			return childMinSize
		}
	}

	// Returns on 2 axis
	return Size{solveOnAxis(X), solveOnAxis(Y)}
}

// Solves starting position according to solved size (with solveConstraintsSize),
// and gravity.
func (box BoxObject) solveConstraintsPosition() Point {
	// Alias to child
	child := box.children[0]
	// Solve 1D problem
	solveOnAxis := func(axis Axis) int {
		// Coordinate of frame end (xStart + size)
		cEnd := box.position.GetComponent(axis) + int(box.actualSize.GetComponent(axis))
		// Coordinate of frame center (xStart + size / 2)
		cMiddle := box.position.GetComponent(axis) + int(box.actualSize.GetComponent(axis))/2
		switch box.gravity.GetComponent(axis).TransformAlignment() {
		case Left:
			// xStart
			return box.position.GetComponent(axis)
		case Center:
			// xCenter - childSize / 2
			return cMiddle - int(child.GetActualSize().GetComponent(axis))/2
		case Right:
			// xEnd - childSize
			return cEnd - int(child.GetActualSize().GetComponent(axis))
		default:
			return 0
		}
	}

	return Point{solveOnAxis(X), solveOnAxis(Y)}
}

// Renders box and it's child
func (box *BoxObject) Render() (Canvas, error) {
	// Allocating canvas
	canvas, error := AllocateCanvas(box.GetActualSize())
	if error != nil {
		return nil, error
	}
	// Alias child
	child := box.children[0]
	// Solve and apply child size
	child.SetActualSize(box.solveConstraintsSize())
	// Solve child position
	child.SetPosition(box.solveConstraintsPosition())
	// Get child's rendered canvas
	childCanvas, childError := child.Render()
	if childError != nil {
		return nil, childError
	}
	// Calculate drawing start position
	// Get frames of objects
	boxFrame := box.GetFrame()
	childFrame := child.GetFrame()
	// Cut child frame
	cuttedChildFrame := childFrame.Cut(*boxFrame) // In box coordinates
	startX, endX := cuttedChildFrame.GetStartEnd(X)
	startY, endY := cuttedChildFrame.GetStartEnd(Y)

	for x := startX; x < endX; x++ {
		for y := startY; y < endY; y++ {
			canvas[x][y] = childCanvas[x-child.GetPosition().X][y-child.GetPosition().Y]
		}
	}
	return canvas, nil
}

// Set size
func (box *BoxObject) Size(size Size) *BoxObject {
	box.SetSize(size)
	return box
}

// Set min size
func (box *BoxObject) MinSize(size Size) *BoxObject {
	box.SetMinSize(size)
	return box
}

// Set max size
func (box *BoxObject) MaxSize(size Size) *BoxObject {
	box.SetMaxSize(size)
	return box
}

// Set gravity
func (box *BoxObject) Gravity(gravity Gravity) *BoxObject {
	box.SetGravity(gravity)
	return box
}

// Represents unviersal stack object
type AbstractStackObject struct {
	AbstractContainer
	axis    Axis
	padding int
}

func AbstractStack(axis Axis, children ...View) *AbstractStackObject {
	stack := new(AbstractStackObject)
	stack.children = children
	stack.SetPosition(Point{0, 0})
	stack.SetMinSize(Size{0, 0})
	stack.SetMaxSize(Size{Infinite, Infinite})
	stack.SetGesture(nil)
	stack.SetGravity(Gravity{Center, Center})
	stack.padding = 0
	return stack
}

// Set size
func (stack *AbstractStackObject) Size(size Size) *AbstractStackObject {
	stack.SetSize(size)
	return stack
}

// Set min size
func (stack *AbstractStackObject) MinSize(size Size) *AbstractStackObject {
	stack.SetMinSize(size)
	return stack
}

// Set max size
func (stack *AbstractStackObject) MaxSize(size Size) *AbstractStackObject {
	stack.SetMaxSize(size)
	return stack
}

// Set padding
func (stack *AbstractStackObject) Padding(paddint int) *AbstractStackObject {
	stack.padding = paddint
	return stack
}

// Set gravity
func (stack *AbstractStackObject) Gravity(gravity Gravity) *AbstractStackObject {
	stack.SetGravity(gravity)
	return stack
}

func (stack *AbstractStackObject) Render() (Canvas, error) {
	canvas, error := AllocateCanvas(stack.GetActualSize())
	if error != nil {
		return nil, error
	}

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
