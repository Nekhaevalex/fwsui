// Provides core container components such as AbstractContainer Container interface
// Also provides basic containers such as Box, HStack, VStack, ZStack
package fwsui

import (
	"errors"
)

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
func (ac AbstractContainer) ApplyChildrenGravity() {
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
	box.children[0] = child
	box.SetPosition(Point{0, 0})
	box.SetMinSize(Size{0, 0})
	box.SetMaxSize(Size{Infinite, Infinite})
	box.SetGesture(nil)
	box.SetGravity(Gravity{Center, Center})
	return box
}

// Renders box and it's child
func (box *BoxObject) Render() (Canvas, error) {
	// Allocating canvas
	canvas, error := AllocateCanvas(box.GetActualSize())
	if error != nil {
		return nil, errors.Join(errors.New("BoxObject was not able to create canvas"), error)
	}
	// Alias child
	child := box.children[0]
	// Solve and apply child size
	xSize := KuruzovSolver(X, []SizeInterval{child.GetSizeInterval()}, box.actualSize)[0]
	ySize := KuruzovSolver(Y, []SizeInterval{child.GetSizeInterval()}, box.actualSize)[0]
	child.SetActualSize(Size{xSize, ySize})
	// Solve child position
	x := CoordinatesSolver(X, child.GetActualSize(), box.GetActualSize(), box.gravity.Horizontal)
	y := CoordinatesSolver(Y, child.GetActualSize(), box.GetActualSize(), box.gravity.Vertical)
	child.SetPosition(Point{x, y})
	// Get child's rendered canvas
	childCanvas, childError := child.Render()
	if childError != nil {
		return nil, childError
	}
	canvas.Inpaint(childCanvas, Vector(child.GetPosition()))
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

func (stack *AbstractStackObject) renderPlaneStack() (Canvas, error) {
	// Allocate canvas
	canvas, error := AllocateCanvas(stack.GetActualSize())
	if error != nil {
		return nil, errors.Join(errors.New("AbstractStackObject was not able to create canvas (plane)"), error)
	}

	// Solve size constraints
	// Retrieving size intervals
	intervals := make([]SizeInterval, len(stack.children))
	for i, child := range stack.children {
		intervals[i] = child.GetSizeInterval()
	}
	// Apply padding to actual size
	actualSizeCopy := stack.GetActualSize()
	newSizeAlongAxis := stack.actualSize.GetComponent(stack.axis) - uint(stack.padding)*uint(len(stack.children)-1)
	actualSizeCopy.SetComponent(newSizeAlongAxis, stack.axis)
	// Solving stacked sizes
	stackedSizes := KuruzovSolver(stack.axis, intervals, actualSizeCopy)
	// Solving parallel sizes
	parallelSizes := make([]uint, len(stack.children))
	for i, child := range stack.children {
		childSize := KuruzovSolver(stack.axis.PlaneOrthogonal(), []SizeInterval{child.GetSizeInterval()}, stack.GetActualSize())[0]
		parallelSizes[i] = childSize
	}
	// Solving coordinates
	translationVector := Vector{0, 0}
	paddingVector := Vector{stack.padding, stack.padding}
	for i, child := range stack.children {
		// Set resulted size
		childActualSize := Size{0, 0}
		childActualSize.SetComponent(stackedSizes[i], stack.axis)
		childActualSize.SetComponent(parallelSizes[i], stack.axis.PlaneOrthogonal())
		child.SetActualSize(childActualSize)

		// Set unshifted child position
		x := CoordinatesSolver(stack.axis, childActualSize, childActualSize, stack.GetGravity().GetComponent(stack.axis))
		y := CoordinatesSolver(stack.axis.PlaneOrthogonal(), childActualSize, childActualSize, stack.GetGravity().GetComponent(stack.axis.PlaneOrthogonal()))
		unshiftedChildPos := Vector{0, 0}
		unshiftedChildPos.SetComponent(x, stack.axis)
		unshiftedChildPos.SetComponent(y, stack.axis.PlaneOrthogonal())

		translatedChildPos := translationVector.Add(unshiftedChildPos)

		child.SetPosition(Point(translatedChildPos))

		nextStartPos := translatedChildPos.Add(childActualSize.ToVector()).Add(paddingVector).Project(stack.axis)
		translationVector = nextStartPos

		childCanvas, childError := child.Render()
		if childError != nil {
			return nil, errors.Join(errors.New("child failed to render"), childError)
		}
		canvas.Inpaint(childCanvas, translationVector.Sub(unshiftedChildPos))
	}
	return canvas, nil
}

func (stack *AbstractStackObject) renderZStack() (Canvas, error) {
	canvas, error := AllocateCanvas(stack.GetActualSize())
	if error != nil {
		return nil, errors.Join(errors.New("AbstractStackObject was not able to create canvas (zstack)"), error)
	}
	for _, child := range stack.children {
		boxed := Box(child)
		boxed.SetGravity(stack.GetGravity())
		boxed.SetPosition(stack.GetPosition())
		layer, error := boxed.Render()
		if error != nil {
			return nil, error
		}
		for x := 0; x < int(stack.actualSize.Width); x++ {
			for y := 0; y < int(stack.actualSize.Height); y++ {
				canvas[x][y] = layer[x][y].Over(canvas[x][y])
			}
		}
	}
	return canvas, nil
}

func (stack *AbstractStackObject) Render() (Canvas, error) {
	switch stack.axis {
	case X, Y:
		return stack.renderPlaneStack()
	case Z:
		return stack.renderZStack()
	default:
		return nil, errors.New("impossible axis provided")
	}
}

func HStack(children ...View) *AbstractStackObject {
	return AbstractStack(X, children...)
}

func VStack(children ...View) *AbstractStackObject {
	return AbstractStack(Y, children...)
}

func ZStack(children ...View) *AbstractStackObject {
	return AbstractStack(Z, children...)
}
