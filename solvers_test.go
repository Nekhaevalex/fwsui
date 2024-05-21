package fwsui

import (
	"testing"
)

func TestAllFixed(t *testing.T) {
	size := Size{5, 5}
	intervals := make([]SizeInterval, 10)
	for i := 0; i < 10; i++ {
		intervals[i] = SizeInterval{size, size}
	}
	if !allFixed(X, intervals...) {
		t.FailNow()
	}
	biggerSize := Size{10, 10}
	intervals[5] = SizeInterval{size, biggerSize}
	if allFixed(X, intervals...) {
		t.FailNow()
	}
}

func TestKuruzovSolver(t *testing.T) {
	// Normal solution
	intervals := []SizeInterval{
		{
			Min: Size{5, 5}, Max: Size{10, 10},
		},
		{
			Min: Size{10, 10}, Max: Size{10, 10},
		},
		{
			Min: Size{5, 5}, Max: Size{10, 10},
		},
	}
	constr := Size{27, 27}
	results := KuruzovSolver(X, intervals, constr)
	expectedResults := []uint{8, 10, 9}
	for i, result := range results {
		if result != expectedResults[i] {
			t.Fatal("Wrong solution:", results)
		}
	}

	// Out of bound case
	intervals = []SizeInterval{
		{
			Min: Size{10, 10}, Max: Size{10, 10},
		},
		{
			Min: Size{10, 10}, Max: Size{10, 10},
		},
		{
			Min: Size{10, 10}, Max: Size{10, 10},
		},
	}
	constr = Size{27, 27}
	results = KuruzovSolver(X, intervals, constr)
	expectedResults = []uint{10, 10, 10}
	for i, result := range results {
		if result != expectedResults[i] {
			t.Fatal("Wrong solution:", results)
		}
	}
}

func TestCoordinateSolver(t *testing.T) {
	outerSize := Size{8, 8}
	size := Size{10, 2}
	if CoordinatesSolver(X, size, outerSize, Left) != 0 {
		t.Fatal("Left failed")
	}
	if CoordinatesSolver(X, size, outerSize, Center) != -1 {
		t.Fatal("Center failed")
	}
	if CoordinatesSolver(X, size, outerSize, Right) != -2 {
		t.Fatal("Right failed")
	}
}
