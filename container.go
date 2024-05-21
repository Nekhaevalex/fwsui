// Provides core container components such as AbstractContainer Container interface
// Also provides basic containers such as Box, HStack, VStack, ZStack
package fwsui

import (
	"errors"
	"log"
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
	box.children[0] = child
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
	if child == nil {
		log.Fatal("nil reference to child in box", box)
	}

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
		return nil, errors.Join(errors.New("BoxObject was not able to create canvas"), error)
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

func (stack *AbstractStackObject) renderPlaneStack() (Canvas, error) {
	// Allocate canvas
	canvas, error := AllocateCanvas(stack.GetActualSize())
	if error != nil {
		return nil, errors.Join(errors.New("AbstractStackObject was not able to create canvas (plane)"), error)
	}

	// Standard value checks
	// 1. Check if all sizes are fixed
	allFixedCheck := func() bool {
		fixedChildren := 0
		for _, child := range stack.children {
			if child.IsFixedSize() {
				fixedChildren++
			}
		}
		return fixedChildren == len(stack.children)
	}

	// Solving along stack axis
	solveConstraintsAlongAxis := func() {
		// Solving with Kuruzov's method:
		// x_i = x_i^lower + (x_i^upper - x_i^lower) * alpha
		// alpha = (x^upper - sum_i x_i^lower) / sum_i (x_i^upper - x_i^lower)

		// Calculating alpha
		// retrieving xUpper from stack end subtracting padding space
		// We also assume that actual size of stack was already evaluated
		xUpper := int(stack.GetActualSize().GetComponent(stack.axis)) - (stack.padding * (len(stack.children) - 1))
		// upper sum
		sum1 := 0
		for _, child := range stack.children {
			xILower := child.GetMinSize().GetComponent(stack.axis)
			sum1 += int(xILower)
		}
		// lower sum
		sum2 := 0
		for _, child := range stack.children {
			xIUpper := child.GetMaxSize().GetComponent(stack.axis)
			// Override child max size to xUpper if its infinite
			if xIUpper == Infinite {
				xIUpper = uint(xUpper)
			}
			xILower := child.GetMinSize().GetComponent(stack.axis)
			diff := xIUpper - xILower
			sum2 += int(diff)
		}
		alpha := float32(xUpper-sum1) / float32(sum2)

		// Calculating x_i/y_i
		for _, child := range stack.children {
			xIUpper := child.GetMaxSize().GetComponent(stack.axis)
			// Override child max size to xUpper if its infinite
			if xIUpper == Infinite {
				xIUpper = uint(xUpper)
			}
			xILower := child.GetMinSize().GetComponent(stack.axis)
			xI := xILower + (xIUpper-xILower)*uint(alpha)
			actualSize := Size{0, 0}
			actualSize.SetComponent(xI, stack.axis)

			yI := min(
				stack.GetActualSize().GetComponent(stack.axis.PlaneOrthogonal()),
				child.GetMaxSize().GetComponent(stack.axis.PlaneOrthogonal()),
			)
			yI = max(yI, child.GetMinSize().GetComponent(stack.axis.PlaneOrthogonal()))
			actualSize.SetComponent(yI, stack.axis.PlaneOrthogonal())
			child.SetActualSize(actualSize)
		}
	}

	// Solving if needed
	if !allFixedCheck() {
		solveConstraintsAlongAxis()
	}

	// With all actual sizes now known we need to solve gravity
	// Solving it with boxes
	// Translation vector
	tVector := Vector{0, 0}
	paddingVector := Vector{0, 0}
	paddingVector.SetComponent(stack.padding, stack.axis)
	for _, child := range stack.children {
		box := Box(child)
		box.SetSize(child.GetActualSize())
		box.SetActualSize(child.GetActualSize())
		box.SetGravity(stack.GetGravity())
		box.SetPosition(Point(tVector))
		tVector = tVector.Add(child.GetActualSize().ToVector().Project(stack.axis)).Add(paddingVector)
		// Render stage
		childCanvas, error := box.Render()
		if error != nil {
			return nil, error
		}

		childFrame := child.GetFrame().Cut(*stack.GetFrame())
		startX, endX := childFrame.GetStartEnd(X)
		startY, endY := childFrame.GetStartEnd(Y)

		for x := startX; x < endX; x++ {
			for y := startY; y < endY; y++ {
				canvas[x][y] = childCanvas[x-child.GetPosition().X][y-child.GetPosition().Y]
			}
		}
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
