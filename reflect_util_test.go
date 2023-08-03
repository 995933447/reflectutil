package reflectutil

import (
	"fmt"
	"testing"
)

func TestCopySameFields2(t *testing.T) {
	type A struct {
	}
	type B struct {
		A *A
		T string
	}

	var (
		B2 B
		B1 B
	)
	B2.T = "123"
	if err := CopySameFields(&B2, &B1); err != nil {
		t.Log(err.Error())
	}

	t.Log(B1)
}

func TestCopySameFields(t *testing.T) {
	type A struct {
		A string
		b string
		B uint64
		C bool
		D func()
	}

	type B struct {
		A string
		C bool
		D func()
	}

	a := &A{
		A: "123",
		b: "234",
		B: 90,
		C: true,
		D: func() {
			fmt.Println("Hello world")
		},
	}

	var b B = B{}
	if err := CopySameFields(a, &b); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%+v\n", b)

	b.D()
}
