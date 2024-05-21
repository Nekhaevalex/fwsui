/*
	Geometry Tests
*/

package fwsui

import (
	"math"
	"testing"
)

/******************************************************************************/
// Point
/******************************************************************************/

func TestPlaneOrthogonal(t *testing.T) {
	if X.PlaneOrthogonal() != Y {
		t.Fail()
	}
	if Y.PlaneOrthogonal() != X {
		t.Fail()
	}
	if Z.PlaneOrthogonal() != Z {
		t.Fail()
	}
}

func TestPointEqual(t *testing.T) {
	p1 := Point{10, 10}
	p2 := Point{10, 10}
	if !p1.Equal(p2) {
		t.Fatal()
	}
}

func TestTranslate(t *testing.T) {
	p := Point{10, 10}
	v := Vector{5, 5}
	o := Point{15, 15}
	if !p.Translate(v).Equal(o) {
		t.Fatal("Points are not equal:", p, o)
	}
}

func TestPointUnpack(t *testing.T) {
	p := Point{10, 10}
	x, y := p.Unpack()
	if !(x == 10 && y == 10) {
		t.Fatal("Points are not equal:", p)
	}
}

func TestPointGetComponent(t *testing.T) {
	p := Point{10, 15}
	if p.GetComponent(X) != 10 {
		t.Fail()
	}
	if p.GetComponent(Y) != 15 {
		t.Fail()
	}
	if p.GetComponent(Z) != 0 {
		t.Fail()
	}
}

func TestPointSetComponent(t *testing.T) {
	p := Point{10, 15}
	p.SetComponent(20, X)
	p.SetComponent(25, Y)

	if p.GetComponent(X) != 20 {
		t.Fail()
	}
	if p.GetComponent(Y) != 25 {
		t.Fail()
	}
	if p.GetComponent(Z) != 0 {
		t.Fail()
	}

	p2 := p.SetComponent(35, Y)
	if p2 != &p {
		t.Fail()
	}
}

/******************************************************************************/
// Vector
/******************************************************************************/

func TestVectorGetComponents(t *testing.T) {
	v := Vector{10, 15}
	if v.GetComponent(X) != 10 {
		t.Fail()
	}
	if v.GetComponent(Y) != 15 {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}
}

func TestVectorSetComponent(t *testing.T) {
	v := Vector{10, 15}
	v.SetComponent(20, X)
	v.SetComponent(25, Y)

	if v.GetComponent(X) != 20 {
		t.Fail()
	}
	if v.GetComponent(Y) != 25 {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}

	v2 := v.SetComponent(35, Y)
	if v2 != &v {
		t.Fail()
	}
}

func TestVectorEqual(t *testing.T) {
	p1 := Vector{10, 10}
	p2 := Vector{10, 10}
	if !p1.Equal(p2) {
		t.Fatal()
	}
}

func TestVectorAdd(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{5, 2}
	vV := Vector{15, 12}
	if !v1.Add(v2).Equal(vV) {
		t.FailNow()
	}
}

func TestVectorSub(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{5, 2}
	vV := Vector{5, 8}
	if !v1.Sub(v2).Equal(vV) {
		t.FailNow()
	}
}

func TestMul(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{30, 30}
	if !v1.Mul(3).Equal(v2) {
		t.FailNow()
	}
}

func TestVecMul(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{5, 2}
	vV := Vector{50, 20}
	if !v1.VecMul(v2).Equal(vV) {
		t.FailNow()
	}
}

func TestReverse(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{-10, -10}
	if !v1.Reverse().Equal(v2) {
		t.FailNow()
	}
}

func TestDoc(t *testing.T) {
	v1 := Vector{10, 10}
	v2 := Vector{15, 20}
	if v1.Dot(v2) != 350 {
		t.FailNow()
	}
}

func TestNorm(t *testing.T) {
	v1 := Vector{2, 5}
	if v1.Norm() != math.Sqrt(29) {
		t.FailNow()
	}
}

func TestProject(t *testing.T) {
	v1 := Vector{15, 35}
	if !v1.Project(X).Equal(Vector{15, 0}) {
		t.Fatal("Not equal along X")
	}
	if !v1.Project(Y).Equal(Vector{0, 35}) {
		t.Fatal("Not equal along Y")
	}
	if !v1.Project(Z).Equal(Vector{0, 0}) {
		t.Fatal("Not equal along Z")
	}
}

/******************************************************************************/
// Size
/******************************************************************************/

func TestInfinityChecks(t *testing.T) {
	size := Size{Infinite, Infinite}
	if !size.InfiniteWidth() {
		t.Fatal("Infinity check X failed")
	}
	if !size.InfiniteHeight() {
		t.Fatal("Infinity check Y failed")
	}
	x, y := size.IsInfinite()
	if !(x && y) {
		t.Fatal("Both infinity checks failed")
	}
}

func TestSizeEqual(t *testing.T) {
	s1 := Size{5, 10}
	s2 := Size{5, 10}
	if !s1.Equal(s2) {
		t.Fatal("Not equal")
	}
}

func TestSizeGetComponents(t *testing.T) {
	v := Size{10, 15}
	if v.GetComponent(X) != 10 {
		t.Fail()
	}
	if v.GetComponent(Y) != 15 {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}
}

func TestSizeSetComponent(t *testing.T) {
	v := Size{10, 15}
	v.SetComponent(20, X)
	v.SetComponent(25, Y)

	if v.GetComponent(X) != 20 {
		t.Fail()
	}
	if v.GetComponent(Y) != 25 {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}

	v2 := v.SetComponent(35, Y)
	if v2 != &v {
		t.Fail()
	}
}
