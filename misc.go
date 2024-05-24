package fwsui

// Some useful local functions

func max[T int | uint | float32 | float64](x, y T) T {
	if x > y {
		return x
	}
	return y
}

func min[T int | uint | float32 | float64](x, y T) T {
	if x > y {
		return y
	}
	return x
}

func abs[T int | float32 | float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
