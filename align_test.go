package fwsui

import "testing"

func TestTransformAlignment(t *testing.T) {
	if Left.TransformAlignment() != Left {
		t.Fatal("Left -> Left")
	}
	if Right.TransformAlignment() != Right {
		t.Fatal("Right -> Right")
	}
	if Center.TransformAlignment() != Center {
		t.Fatal("Center -> Center")
	}
	if Top.TransformAlignment() != Left {
		t.Fatal("Top -> Left")
	}
	if Bottom.TransformAlignment() != Right {
		t.Fatal("Bottom -> Right")
	}
}

func TestGravityGetComponents(t *testing.T) {
	v := Gravity{Left, Top}
	if v.GetComponent(X) != Left {
		t.Fail()
	}
	if v.GetComponent(Y) != Top {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}
}

func TestGravitySetComponent(t *testing.T) {
	v := Gravity{Left, Top}
	v.SetComponent(Right, X)
	v.SetComponent(Bottom, Y)

	if v.GetComponent(X) != Right {
		t.Fail()
	}
	if v.GetComponent(Y) != Bottom {
		t.Fail()
	}
	if v.GetComponent(Z) != 0 {
		t.Fail()
	}

	v2 := v.SetComponent(Top, Y)
	if v2 != &v {
		t.Fail()
	}
}
