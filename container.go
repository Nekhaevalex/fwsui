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
// Note: AbstractContainer itself implements both Container and View interfaces.
type AbstractContainer struct {
	AbstractView
	gravity  Gravity
	children []View
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

func (ac AbstractContainer) FindGesture(me MouseEvent) Gesture {
	if !ac.GetFrame().IsInside(me.Position) {
		return nil
	}
	for _, c := range ac.children {
		g := c.FindGesture(me)
		if g != nil {
			return g
		}
	}
	return nil
}

// Container - interface for implementing containers.
// Each container must provide several methods.
//
// Since each container should contain gravity variable (see AbstractContainer),
// Container should implement setter and getter.
//
// Each of View in Container may contain Gesture.
type Container interface {
	GetGravity() Gravity
	SetGravity(gravity Gravity)
	Render() (Canvas, error)
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
	box.SetGestures(nil)
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
	child.SetPosition(Point(Vector{x, y}.Add(Vector(box.GetPosition()))))
	// Get child's rendered canvas
	childCanvas, childError := child.Render()
	if childError != nil {
		return nil, childError
	}
	canvas.Inpaint(childCanvas, Vector{x, y})
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
	stack.SetGestures(nil)
	stack.SetGravity(Gravity{Center, Center})
	stack.axis = axis
	stack.padding = 0
	return stack
}

// Adds new element to the end of stack
func (stack *AbstractStackObject) Add(new View) *AbstractStackObject {
	stack.children = append(stack.children, new)
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

	// Defining axis
	longAxis := stack.axis
	shortAxis := stack.axis.PlaneOrthogonal()

	// Solve size constraints
	// Retrieving size intervals
	intervals := make([]SizeInterval, len(stack.children))
	for i, child := range stack.children {
		intervals[i] = child.GetSizeInterval()
	}
	// Apply padding to actual size
	actualSizeCopy := stack.GetActualSize()
	newSizeAlongAxis := stack.actualSize.GetComponent(longAxis) - uint(stack.padding)*uint(len(stack.children)-1)
	actualSizeCopy.SetComponent(newSizeAlongAxis, longAxis)
	// Solving stacked sizes
	stackedSizes := KuruzovSolver(longAxis, intervals, actualSizeCopy)
	// Solving parallel sizes
	parallelSizes := make([]uint, len(stack.children))
	var parallelSizeMax uint = 0
	for i, child := range stack.children {
		childSize := KuruzovSolver(shortAxis, []SizeInterval{child.GetSizeInterval()}, stack.GetActualSize())[0]
		if i > 0 {
			parallelSizeMax = max(parallelSizeMax, childSize)
		} else {
			parallelSizeMax = childSize
		}
		parallelSizes[i] = childSize
	}
	// Solving coordinates
	translationVector := Vector{0, 0}
	paddingVector := Vector{stack.padding, stack.padding}
	for i, child := range stack.children {
		// Set resulted size
		childActualSize := Size{0, 0}
		childActualSize.SetComponent(stackedSizes[i], longAxis)
		childActualSize.SetComponent(parallelSizes[i], shortAxis)
		child.SetActualSize(childActualSize)
		childBoxSize := childActualSize
		childBoxSize.SetComponent(parallelSizeMax, shortAxis)

		// Set unshifted child position

		x := CoordinatesSolver(longAxis, childActualSize, childBoxSize, stack.GetGravity().GetComponent(longAxis))
		y := CoordinatesSolver(shortAxis, childActualSize, childBoxSize, stack.GetGravity().GetComponent(shortAxis))
		unshiftedChildPos := Vector{0, 0}
		unshiftedChildPos.SetComponent(x, longAxis)
		unshiftedChildPos.SetComponent(y, shortAxis)

		translatedChildPos := translationVector.Add(unshiftedChildPos)

		// Shift child with self position before render to prevent unshifted gestures
		child.SetPosition(Point(Vector(Point(translatedChildPos)).Add(Vector(stack.GetPosition()))))

		nextStartPos := translatedChildPos.
			Add(childActualSize.ToVector()).
			Add(paddingVector).
			Project(longAxis)
		translationVector = nextStartPos

		childCanvas, childError := child.Render()
		if childError != nil {
			return nil, errors.Join(errors.New("child failed to render"), childError)
		}
		canvas.Inpaint(childCanvas, translatedChildPos)
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
		boxed.SetPosition(stack.GetPosition())
		boxed.SetActualSize(stack.GetActualSize())
		boxed.SetGravity(stack.GetGravity())
		layer, error := boxed.Render()
		if error != nil {
			return nil, error
		}
		canvas.Inpaint(layer, Vector{0, 0})
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
