package fwsui

// Some useful local functions

func abs[T int | float32 | float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}
