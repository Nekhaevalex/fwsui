/*
Provides constraint solvers used in containers and their internal functions
Solvers are compliant with FWSUI geometry types and views
*/

package fwsui

import "math"

// Returns true if all intervals are fixed along axis
func allFixed(axis Axis, intervals ...SizeInterval) bool {
	for _, interval := range intervals {
		if !interval.Fixed(axis) {
			return false
		}
	}
	return true
}

// Solves size constraints provided in intervals with respect to outerSize along
// specified axis using Kuruzov's method.
//
// For set with at least one unfixed size each size is calcluated as
// x_i = x_i^lower + (x_i^upper - x_i^lower) * alpha,
// where alpha = (x^upper - sum_i x_i^lower) / sum_i (x_i^upper - x_i^lower)
//
// Returns slice of 1D sizes.
func KuruzovSolver(axis Axis, intervals []SizeInterval, outerSize Size) []uint {
	// Check if all sizes are fixed
	// If all are fixed then return array of fixed sizes
	sizes := make([]uint, len(intervals))
	if allFixed(axis, intervals...) {
		for i, interval := range intervals {
			sizes[i] = interval.Max.GetComponent(axis)
		}
		return sizes
	}
	// Calculating alpha
	// retrieving xUpper from stack end subtracting padding space
	// We also assume that actual size of stack was already evaluated
	xUpper := float64(outerSize.GetComponent(axis))

	// first sum
	sum1 := 0.0
	for _, interval := range intervals {
		xILower := float64(interval.Min.GetComponent(axis))
		sum1 += xILower
	}

	// check if minimal sum exceeds size and if yes, return
	if sum1 >= xUpper {
		for i, interval := range intervals {
			sizes[i] = interval.Min.GetComponent(axis)
		}
		return sizes
	}

	// second sum
	sum2 := 0.0
	for _, interval := range intervals {
		xIUpper := float64(interval.Max.GetComponent(axis))
		// Override child max size to xUpper if its infinite
		if xIUpper == Infinite {
			xIUpper = xUpper
		}
		xILower := float64(interval.Min.GetComponent(axis))
		diff := xIUpper - xILower
		sum2 += diff
	}
	alpha := float64(xUpper-sum1) / float64(sum2)

	// Calculating x_i
	var checkSum uint = 0
	for i, interval := range intervals {
		xIUpper := float64(interval.Max.GetComponent(axis))
		if xIUpper == Infinite {
			xIUpper = xUpper
		}
		xILower := float64(interval.Min.GetComponent(axis))
		xI := xILower + (xIUpper-xILower)*alpha
		uxi := uint(math.Round(xI))
		checkSum += uxi
		sizes[i] = uxi
	}

	// If after rounding checksum exceeds outerSize
	diff := checkSum - outerSize.GetComponent(axis)
	if diff > 0 {
		for i, interval := range intervals {
			if !interval.Fixed(axis) && sizes[i]-interval.Min.GetComponent(axis) >= diff {
				sizes[i] -= diff
				break
			}
		}
	}

	return sizes
}

// Returns coordinate component along axis of object with size which is aligned
// according to alignment inside object with size outerSize
func CoordinatesSolver(axis Axis, size Size, outerSize Size, alignment Align) int {
	outerStart := 0
	vOuterSize := outerSize.ToVector().GetComponent(axis)
	vSize := size.ToVector().GetComponent(axis)

	switch alignment.TransformAlignment() {
	case Left:
		return outerStart
	case Center:
		return vOuterSize/2 - vSize/2
	case Right:
		return vOuterSize - vSize
	default:
		return outerStart
	}
}
